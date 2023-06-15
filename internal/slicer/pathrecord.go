package slicer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/canonical/chisel/internal/db"
	"github.com/canonical/chisel/internal/deb"
	"github.com/canonical/chisel/internal/fsutil"
	"github.com/canonical/chisel/internal/strdist"
)

type pathRecorder interface {
	addSlicePath(slice, path string)
	addSliceGlob(slice, glob string)
	onData(source string, size int64) (deb.ConsumeData, error)
	onCreate(source, target, link string, mode fs.FileMode) error
	addTarget(target, link string, mode fs.FileMode, data []byte)
	markMutated(target string)
	removeTarget(target string)
	updateTargets(root string) error
	updateDB(writeDB WriteDB) error
}

type contentInfo struct {
	size   int64
	digest string
}

func computeDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return fmt.Sprintf("%x", digest)
}

type pathRecImpl struct {
	pathSlices     map[string]db.StringSortedSet
	globSlices     map[string]db.StringSortedSet
	targetToSource map[string]string
	sourceContent  map[string]contentInfo
	targets        map[string]*db.Path
	mutatedTargets map[string]bool
}

var _ pathRecorder = (*pathRecImpl)(nil)

func newPathRecorder() pathRecorder {
	return &pathRecImpl{
		pathSlices:     make(map[string]db.StringSortedSet),
		globSlices:     make(map[string]db.StringSortedSet),
		targetToSource: make(map[string]string),
		sourceContent:  make(map[string]contentInfo),
		targets:        make(map[string]*db.Path),
		mutatedTargets: make(map[string]bool),
	}
}

func (rec *pathRecImpl) addSlicePath(slice, path string) {
	rec.pathSlices[path] = rec.pathSlices[path].AddStrings(slice)
}

func (rec *pathRecImpl) addSliceGlob(slice, glob string) {
	rec.globSlices[glob] = rec.pathSlices[glob].AddStrings(slice)
}

func (rec *pathRecImpl) onData(source string, size int64) (deb.ConsumeData, error) {
	consume := func(reader io.Reader) error {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		digest := computeDigest(data)
		rec.sourceContent[source] = contentInfo{size, digest}
		return nil
	}
	return consume, nil
}

func (rec *pathRecImpl) onCreate(source, target, link string, mode fs.FileMode) error {
	info := db.Path{
		Path: target,
		Mode: mode,
		Link: link,
	}
	rec.targets[target] = &info
	rec.targetToSource[target] = source
	return nil
}

func (rec *pathRecImpl) addTarget(target, link string, mode fs.FileMode, data []byte) {
	info := db.Path{
		Path: target,
		Mode: mode,
		Link: link,
	}
	if data != nil {
		info.Size = int64(len(data))
		info.SHA256 = computeDigest(data)
	}
	rec.targets[target] = &info
	for parent := fsutil.Dir(target); parent != "/"; parent = fsutil.Dir(parent) {
		if _, ok := rec.targets[parent]; ok {
			break
		}
		rec.targets[parent] = &db.Path{
			Path: parent,
			Mode: fs.ModeDir | 0755,
		}
	}
}

func (rec *pathRecImpl) markMutated(target string) {
	rec.mutatedTargets[target] = true
}

func (rec *pathRecImpl) removeTarget(target string) {
	delete(rec.targets, target)
}

func (rec *pathRecImpl) completeTarget(info *db.Path) {
	info.Mode = info.Mode & 07777

	source := rec.targetToSource[info.Path]
	if content, ok := rec.sourceContent[source]; ok {
		info.Size = content.size
		info.SHA256 = content.digest
	}

	slices := rec.pathSlices[info.Path]
	for glob, globSlices := range rec.globSlices {
		if strdist.GlobPath(glob, info.Path) {
			slices = slices.AddStrings(globSlices...)
		}
	}

	path := info.Path
	for len(slices) > 0 && path != "/" {
		newSlices := []string{}
		for _, sl := range slices {
			if tmp, ok := info.Slices.AddString(sl); ok {
				info.Slices = tmp
				newSlices = append(newSlices, sl)
			}
		}
		slices = newSlices
		path = fsutil.Dir(path)
		info = rec.targets[path]
	}
}

func (rec *pathRecImpl) refreshTarget(info *db.Path, root string) error {
	if !rec.mutatedTargets[info.Path] || info.SHA256 == "" {
		return nil
	}
	local := filepath.Join(root, info.Path)
	data, err := ioutil.ReadFile(local)
	if err != nil {
		return err
	}
	finalDigest := computeDigest(data)
	if info.SHA256 != finalDigest {
		info.FinalSHA256 = finalDigest
	}
	return nil
}

func (rec *pathRecImpl) updateTargets(root string) (err error) {
	for _, info := range rec.targets {
		rec.completeTarget(info)
		if err = rec.refreshTarget(info, root); err != nil {
			break
		}
	}
	return
}

func (rec *pathRecImpl) updateDB(writeDB WriteDB) error {
	for _, info := range rec.targets {
		if err := writeDB(info); err != nil {
			return fmt.Errorf("cannot write path to db: %w", err)
		}
		for _, sl := range info.Slices {
			content := db.Content{sl, info.Path}
			if err := writeDB(content); err != nil {
				return fmt.Errorf("cannot write content to db: %w", err)
			}
		}
	}
	return nil
}
