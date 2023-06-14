package deb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/blakesmith/ar"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"

	"github.com/canonical/chisel/internal/fsutil"
	"github.com/canonical/chisel/internal/strdist"
)

type ExtractOptions struct {
	Package   string
	TargetDir string
	Extract   map[string][]ExtractInfo
	Globbed   map[string][]string
}

type TargetSelector struct {
	Source string
	Target string
	ID     string
}

type ExtractInfo struct {
	Path     string
	Mode     uint
	Optional bool
	ID       string
}

type TargetInfo struct {
	Path      string
	Mode      int64
	Matches []ExtractInfo
}

type extractContext struct {
	options      *ExtractOptions
	extractPaths map[string][]ExtractInfo
	extractGlobs [][]ExtractInfo
	pendingPaths map[string]bool
	skippedDirs  map[string]int
}

func isPathValid(path string) bool {
	return strings.HasPrefix(path, "/") && path != "/"
}

func (ctx *extractContext) addExtractPath(sourcePath string, exInfos []ExtractInfo) error {
	if !isPathValid(sourcePath) {
		return fmt.Errorf("invalid source path: %#v", sourcePath)
	}
	if len(exInfos) == 0 {
		return fmt.Errorf("source path %#v has no target paths", sourcePath)
	}
	pending := false
	if strings.ContainsAny(sourcePath, "*?") {
		ctx.extractGlobs = append(ctx.extractGlobs, exInfos)
		for _, exInfo := range exInfos {
			if exInfo.Path != sourcePath || exInfo.Mode != 0 {
				return fmt.Errorf("when using wildcards source and target paths must match: %s", sourcePath)
			}
			if !pending && !exInfo.Optional {
				pending = true
			}
		}
	} else {
		ctx.extractPaths[sourcePath] = exInfos
		for _, exInfo := range exInfos {
			if !isPathValid(exInfo.Path) {
				return fmt.Errorf("invalid target path: %#v", exInfo.Path)
			}
			if exInfo.Optional {
				for parent := fsutil.Dir(exInfo.Path); parent != "/"; parent = fsutil.Dir(parent) {
					ctx.extractPaths[parent] = append(ctx.extractPaths[parent], ExtractInfo{
						Path:     parent,
						Optional: true,
					})
				}
			} else if !pending {
				pending = true
			}
		}
	}
	if pending {
		ctx.pendingPaths[sourcePath] = true
	}
	return nil
}

func newExtractContext(options *ExtractOptions) (*extractContext, error) {
	ctx := &extractContext{
		options:      options,
		extractPaths: make(map[string][]ExtractInfo),
		pendingPaths: make(map[string]bool),
	}
	for extractPath, extractInfos := range options.Extract {
		if err := ctx.addExtractPath(extractPath, extractInfos); err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

func Extract(pkgReader io.Reader, options *ExtractOptions) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("cannot extract from package %q: %w", options.Package, err)
		}
	}()

	logf("Extracting files from package %q...", options.Package)

	ctx, err := newExtractContext(options)
	if err != nil {
		return err
	}

	_, err = os.Stat(options.TargetDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("target directory does not exist")
	} else if err != nil {
		return err
	}

	arReader := ar.NewReader(pkgReader)
	var dataReader io.Reader
	for dataReader == nil {
		arHeader, err := arReader.Next()
		if err == io.EOF {
			return fmt.Errorf("no data payload")
		}
		if err != nil {
			return err
		}
		switch arHeader.Name {
		case "data.tar.gz":
			gzipReader, err := gzip.NewReader(arReader)
			if err != nil {
				return err
			}
			defer gzipReader.Close()
			dataReader = gzipReader
		case "data.tar.xz":
			xzReader, err := xz.NewReader(arReader)
			if err != nil {
				return err
			}
			dataReader = xzReader
		case "data.tar.zst":
			zstdReader, err := zstd.NewReader(arReader)
			if err != nil {
				return err
			}
			defer zstdReader.Close()
			dataReader = zstdReader
		}
	}
	return ctx.extractData(dataReader, options)
}

type matchResult struct {
	targets      map[string]TargetInfo
	matchedPaths []string
	matchedGlobs []string
}

func (ctx *extractContext) matchTargets(sourcePath string, sourceMode int64) (result matchResult) {
	result.targets = make(map[string]TargetInfo)
	if exInfos, ok := ctx.extractPaths[sourcePath]; ok {
		result.matchedPaths = append(result.matchedPaths, sourcePath)
		for _, exInfo := range exInfos {
			dpInfo, ok := result.targets[exInfo.Path]
			if !ok {
				dpInfo = TargetInfo{
					Path: exInfo.Path,
					Mode: sourceMode,
				}
			}
			if exInfo.Mode != 0 {
				dpInfo.Mode = (sourceMode &^ 07777) | int64(exInfo.Mode&07777)
			}
			dpInfo.Matches = append(dpInfo.Matches, exInfo)
			result.targets[exInfo.Path] = dpInfo
		}
	}
	for _, exInfos := range ctx.extractGlobs {
		glob := exInfos[0].Path
		if strdist.GlobPath(glob, sourcePath) {
			result.matchedGlobs = append(result.matchedGlobs, glob)
			dpInfo, ok := result.targets[sourcePath]
			if !ok {
				dpInfo = TargetInfo{
					Path: sourcePath,
					Mode: sourceMode,
				}
			}
			dpInfo.Matches = append(dpInfo.Matches, exInfos...)
			result.targets[sourcePath] = dpInfo
		}
	}
	return
}

func (ctx *extractContext) extractData(dataReader io.Reader, options *ExtractOptions) error {

	oldUmask := syscall.Umask(0)
	defer func() {
		syscall.Umask(oldUmask)
	}()

	tarReader := tar.NewReader(dataReader)
	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		sourcePath := tarHeader.Name
		if len(sourcePath) < 3 || sourcePath[0] != '.' || sourcePath[1] != '/' {
			continue
		}
		sourcePath = sourcePath[1:]

		match := ctx.matchTargets(sourcePath, tarHeader.Mode)

		if len(match.targets) != 0 {
			for _, path := range match.matchedPaths {
				delete(ctx.pendingPaths, path)
			}
			for _, glob := range match.matchedGlobs {
				delete(ctx.pendingPaths, glob)
				if ctx.options.Globbed != nil {
					ctx.options.Globbed[glob] = append(ctx.options.Globbed[glob], sourcePath)
				}
			}
		} else {
			continue
		}

		var contentCache []byte
		var contentIsCached = true
		if contentIsCached {
			// Read and cache the content so it may be reused.
			// As an alternative, to avoid having an entire file in
			// memory at once this logic might open the first file
			// written and copy it every time. For now, the choice
			// is speed over memory efficiency.
			data, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return err
			}
			contentCache = data
		}

		var pathReader io.Reader = tarReader
		for _, dpInfo := range match.targets {
			tmpHeader := tar.Header{Typeflag: tarHeader.Typeflag, Mode: dpInfo.Mode}
			if contentIsCached {
				pathReader = bytes.NewReader(contentCache)
			}
			err := fsutil.Create(&fsutil.CreateOptions{
				Path: filepath.Join(options.TargetDir, dpInfo.Path),
				Mode: tmpHeader.FileInfo().Mode(),
				Data: pathReader,
				Link: tarHeader.Linkname,
				Dirs: true,
			})
			if err != nil {
				return err
			}
		}
	}

	if len(ctx.pendingPaths) > 0 {
		pendingList := make([]string, 0, len(ctx.pendingPaths))
		for pendingPath := range ctx.pendingPaths {
			pendingList = append(pendingList, pendingPath)
		}
		if len(pendingList) == 1 {
			return fmt.Errorf("no content at %s", pendingList[0])
		} else {
			sort.Strings(pendingList)
			return fmt.Errorf("no content at:\n- %s", strings.Join(pendingList, "\n- "))
		}
	}

	return nil
}
