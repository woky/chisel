package slicer_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/canonical/chisel/internal/archive"
	"github.com/canonical/chisel/internal/setup"
	"github.com/canonical/chisel/internal/slicer"
	"github.com/canonical/chisel/internal/testutil"
)

type testPackageInfo struct {
	name    string
	version string
	archive string
	content []testutil.TarEntry
}

type slicerTest struct {
	summary string
	arch    string
	release map[string]string
	slices  []setup.SliceKey
	hackopt func(c *C, opts *slicer.RunOptions)
	result  map[string]string
	error   string
	pkgs    []testPackageInfo
}

var packageEntries = map[string][]testutil.TarEntry{
	"copyright-symlink-libssl3": {
		{Header: tar.Header{Name: "./"}},
		{Header: tar.Header{Name: "./usr/"}},
		{Header: tar.Header{Name: "./usr/lib/"}},
		{Header: tar.Header{Name: "./usr/lib/x86_64-linux-gnu/"}},
		{Header: tar.Header{Name: "./usr/lib/x86_64-linux-gnu/libssl.so.3", Mode: 00755}},
		{Header: tar.Header{Name: "./usr/share/"}},
		{Header: tar.Header{Name: "./usr/share/doc/"}},
		{Header: tar.Header{Name: "./usr/share/doc/copyright-symlink-libssl3/"}},
		{Header: tar.Header{Name: "./usr/share/doc/copyright-symlink-libssl3/copyright"}},
	},
	"copyright-symlink-openssl": {
		{Header: tar.Header{Name: "./"}},
		{Header: tar.Header{Name: "./etc/"}},
		{Header: tar.Header{Name: "./etc/ssl/"}},
		{Header: tar.Header{Name: "./etc/ssl/openssl.cnf"}},
		{Header: tar.Header{Name: "./usr/"}},
		{Header: tar.Header{Name: "./usr/bin/"}},
		{Header: tar.Header{Name: "./usr/bin/openssl", Mode: 00755}},
		{Header: tar.Header{Name: "./usr/share/"}},
		{Header: tar.Header{Name: "./usr/share/doc/"}},
		{Header: tar.Header{Name: "./usr/share/doc/copyright-symlink-openssl/"}},
		{Header: tar.Header{Name: "./usr/share/doc/copyright-symlink-openssl/copyright", Linkname: "../libssl3/copyright"}},
	},
}

// filesystem entries of copyright file from base-files package that will be
// automatically injected into every slice
var copyrightEntries = map[string]string{
	"/usr/":                               "dir 0755",
	"/usr/share/":                         "dir 0755",
	"/usr/share/doc/":                     "dir 0755",
	"/usr/share/doc/base-files/":          "dir 0755",
	"/usr/share/doc/base-files/copyright": "file 0644 cdb5461d",
}

var slicerTests = []slicerTest{{
	summary: "Basic slicing",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/usr/bin/hello:
						/usr/bin/hallo: {copy: /usr/bin/hello}
						/bin/hallo:     {symlink: ../usr/bin/hallo}
						/etc/passwd:    {text: data1}
						/etc/dir/sub/:  {make: true, mode: 01777}
		`,
	},
	result: map[string]string{
		"/usr/":          "dir 0755",
		"/usr/bin/":      "dir 0755",
		"/usr/bin/hello": "file 0775 eaf29575",
		"/usr/bin/hallo": "file 0775 eaf29575",
		"/bin/":          "dir 0755",
		"/bin/hallo":     "symlink ../usr/bin/hallo",
		"/etc/":          "dir 0755",
		"/etc/dir/":      "dir 0755",
		"/etc/dir/sub/":  "dir 01777",
		"/etc/passwd":    "file 0644 5b41362b",
	},
}, {
	summary: "Glob extraction",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/**/he*o:
		`,
	},
	result: map[string]string{
		"/usr/":          "dir 0755",
		"/usr/bin/":      "dir 0755",
		"/usr/bin/hello": "file 0775 eaf29575",
	},
}, {
	summary: "Create new file under extracted directory",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						# Note the missing /tmp/ here.
						/tmp/new: {text: data1}
		`,
	},
	result: map[string]string{
		"/tmp/":    "dir 01777", // This is the magic.
		"/tmp/new": "file 0644 5b41362b",
	},
}, {
	summary: "Create new nested file under extracted directory",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						# Note the missing /tmp/ here.
						/tmp/new/sub: {text: data1}
		`,
	},
	result: map[string]string{
		"/tmp/":        "dir 01777", // This is the magic.
		"/tmp/new/":    "dir 0755",
		"/tmp/new/sub": "file 0644 5b41362b",
	},
}, {
	summary: "Create new directory under extracted directory",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						# Note the missing /tmp/ here.
						/tmp/new/: {make: true}
		`,
	},
	result: map[string]string{
		"/tmp/":     "dir 01777", // This is the magic.
		"/tmp/new/": "dir 0755",
	},
}, {
	summary: "Conditional architecture",
	arch:    "amd64",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1, arch: amd64}
						/tmp/file2: {text: data1, arch: i386}
						/tmp/file3: {text: data1, arch: [i386, amd64]}
						/usr/bin/hello1: {copy: /usr/bin/hello, arch: amd64}
						/usr/bin/hello2: {copy: /usr/bin/hello, arch: i386}
						/usr/bin/hello3: {copy: /usr/bin/hello, arch: [i386, amd64]}
		`,
	},
	result: map[string]string{
		"/tmp/":           "dir 01777",
		"/tmp/file1":      "file 0644 5b41362b",
		"/tmp/file3":      "file 0644 5b41362b",
		"/usr/":           "dir 0755",
		"/usr/bin/":       "dir 0755",
		"/usr/bin/hello1": "file 0775 eaf29575",
		"/usr/bin/hello3": "file 0775 eaf29575",
	},
}, {
	summary: "Script: write a file",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1, mutable: true}
					mutate: |
						content.write("/tmp/file1", "data2")
		`,
	},
	result: map[string]string{
		"/tmp/":      "dir 01777",
		"/tmp/file1": "file 0644 d98cf53e",
	},
}, {
	summary: "Script: read a file",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1}
						/foo/file2: {text: data2, mutable: true}
					mutate: |
						data = content.read("/tmp/file1")
						content.write("/foo/file2", data)
		`,
	},
	result: map[string]string{
		"/tmp/":      "dir 01777",
		"/tmp/file1": "file 0644 5b41362b",
		"/foo/":      "dir 0755",
		"/foo/file2": "file 0644 5b41362b",
	},
}, {
	summary: "Script: use 'until' to remove file after mutate",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1, until: mutate}
						/foo/file2: {text: data2, mutable: true}
					mutate: |
						data = content.read("/tmp/file1")
						content.write("/foo/file2", data)
		`,
	},
	result: map[string]string{
		"/tmp/":      "dir 01777",
		"/foo/":      "dir 0755",
		"/foo/file2": "file 0644 5b41362b",
	},
}, {
	summary: "Script: use 'until' to remove wildcard after mutate",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/usr/bin**:  {until: mutate}
						/etc/passwd: {until: mutate, text: data1}
		`,
	},
	result: map[string]string{
		"/usr/": "dir 0755",
		"/etc/": "dir 0755",
	},
}, {
	summary: "Script: 'until' does not remove non-empty directories",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/usr/bin/: {until: mutate}
						/usr/bin/hallo: {copy: /usr/bin/hello}
		`,
	},
	result: map[string]string{
		"/usr/":          "dir 0755",
		"/usr/bin/":      "dir 0755",
		"/usr/bin/hallo": "file 0775 eaf29575",
	},
}, {
	summary: "Script: cannot write non-mutable files",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1}
					mutate: |
						content.write("/tmp/file1", "data2")
		`,
	},
	error: `slice base-files_myslice: cannot write file which is not mutable: /tmp/file1`,
}, {
	summary: "Script: cannot read unlisted content",
	slices:  []setup.SliceKey{{"base-files", "myslice2"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice1:
					contents:
						/tmp/file1: {text: data1}
				myslice2:
					mutate: |
						content.read("/tmp/file1")
		`,
	},
	error: `slice base-files_myslice2: cannot read file which is not selected: /tmp/file1`,
}, {
	summary: "Script: can read globbed content",
	slices:  []setup.SliceKey{{"base-files", "myslice1"}, {"base-files", "myslice2"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice1:
					contents:
						/usr/bin/*:
				myslice2:
					mutate: |
						content.read("/usr/bin/hello")
		`,
	},
}, {
	summary: "Relative content root directory must not error",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/tmp/file1: {text: data1}
					mutate: |
						content.read("/tmp/file1")
		`,
	},
	hackopt: func(c *C, opts *slicer.RunOptions) {
		dir, err := os.Getwd()
		c.Assert(err, IsNil)
		opts.TargetDir, err = filepath.Rel(dir, opts.TargetDir)
		c.Assert(err, IsNil)
	},
}, {
	summary: "Can list parent directories of normal paths",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/a/b/c: {text: foo}
						/x/y/: {make: true}
					mutate: |
						content.list("/")
						content.list("/a")
						content.list("/a/b")
						content.list("/x")
						content.list("/x/y")
		`,
	},
}, {
	summary: "Cannot list unselected directory",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/a/b/c: {text: foo}
					mutate: |
						content.list("/a/d")
		`,
	},
	error: `slice base-files_myslice: cannot list directory which is not selected: /a/d/`,
}, {
	summary: "Cannot list file path as a directory",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/a/b/c: {text: foo}
					mutate: |
						content.list("/a/b/c")
		`,
	},
	error: `slice base-files_myslice: content is not a directory: /a/b/c`,
}, {
	summary: "Can list parent directories of globs",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/**/bin/h?llo:
					mutate: |
						content.list("/usr/bin")
		`,
	},
}, {
	summary: "Cannot list directories not matched by glob",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/**/bin/h?llo:
					mutate: |
						content.list("/etc")
		`,
	},
	error: `slice base-files_myslice: cannot list directory which is not selected: /etc/`,
}, {
	summary: "Duplicate copyright symlink is ignored",
	slices:  []setup.SliceKey{{"copyright-symlink-openssl", "bins"}},
	release: map[string]string{
		"slices/mydir/copyright-symlink-libssl3.yaml": `
			package: copyright-symlink-libssl3
			slices:
				libs:
					contents:
						/usr/lib/x86_64-linux-gnu/libssl.so.3:
		`,
		"slices/mydir/copyright-symlink-openssl.yaml": `
			package: copyright-symlink-openssl
			slices:
				bins:
					essential:
						- copyright-symlink-libssl3_libs
						- copyright-symlink-openssl_config
					contents:
						/usr/bin/openssl:
				config:
					contents:
						/etc/ssl/openssl.cnf:
		`,
	},
}, {
	summary: "Can list unclean directory paths",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/a/b/c: {text: foo}
						/x/y/: {make: true}
					mutate: |
						content.list("/////")
						content.list("/a/")
						content.list("/a/b/../b/")
						content.list("/x///")
						content.list("/x/./././y")
		`,
	},
}, {
	summary: "Cannot read directories",
	slices:  []setup.SliceKey{{"base-files", "myslice"}},
	release: map[string]string{
		"slices/mydir/base-files.yaml": `
			package: base-files
			slices:
				myslice:
					contents:
						/x/y/: {make: true}
					mutate: |
						content.read("/x/y")
		`,
	},
	error: `slice base-files_myslice: content is not a file: /x/y`,
}, {
	summary: "Can install packages from multiple archives",
	pkgs: []testPackageInfo{
		{
			name:    "electron",
			version: "1.897",
			archive: "leptons",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./mass/"}},
				{
					Header:  tar.Header{Name: "./mass/electron"},
					Content: []byte("9.1093837015E−31 kg\n"),
				},
				{Header: tar.Header{Name: "./usr/"}},
				{Header: tar.Header{Name: "./usr/share/"}},
				{Header: tar.Header{Name: "./usr/share/doc/"}},
				{Header: tar.Header{Name: "./usr/share/doc/electron/"}},
				{Header: tar.Header{Name: "./usr/share/doc/electron/copyright"}},
			},
		},
		{
			name:    "proton",
			version: "1.886",
			archive: "hadrons",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./mass/"}},
				{
					Header:  tar.Header{Name: "./mass/proton"},
					Content: []byte("1.67262192369E−27 kg\n"),
				},
			},
		},
	},
	release: map[string]string{
		"chisel.yaml": `
			format: chisel-v1
			archives:
				leptons:
					version: 1
					suites: [main]
					components: [main, universe]
				hadrons:
					version: 1
					suites: [main]
					components: [main]
		`,
		"slices/mydir/electron.yaml": `
			package: electron
			slices:
				mass:
					contents:
						/mass/electron:
		`,
		"slices/mydir/proton.yaml": `
			package: proton
			slices:
				mass:
					contents:
						/mass/proton:
		`,
	},
	slices: []setup.SliceKey{
		{"electron", "mass"},
		{"proton", "mass"},
	},
	result: map[string]string{
		"/mass/":                            "dir 0755",
		"/mass/electron":                    "file 0644 a1258e30",
		"/mass/proton":                      "file 0644 a2390d10",
		"/usr/":                             "dir 0755",
		"/usr/share/":                       "dir 0755",
		"/usr/share/doc/":                   "dir 0755",
		"/usr/share/doc/electron/":          "dir 0755",
		"/usr/share/doc/electron/copyright": "file 0644 empty",
	},
}, {
	summary: "Can pick latest packages from multiple archives",
	pkgs: []testPackageInfo{
		{
			name:    "cheetah",
			version: "109.4",
			archive: "vertebrates",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./speed/"}},
				{
					Header:  tar.Header{Name: "./speed/cheetah"},
					Content: []byte("109.4 km/h\n"),
				},
			},
		},
		{
			name:    "ostrich",
			version: "100.0",
			archive: "vertebrates",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./speed/"}},
				{
					Header:  tar.Header{Name: "./speed/ostrich"},
					Content: []byte("100.0 km/h\n"),
				},
			},
		},
		{
			name:    "cheetah",
			version: "120.7",
			archive: "mammals",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./speed/"}},
				{
					Header:  tar.Header{Name: "./speed/cheetah"},
					Content: []byte("120.7 km/h\n"),
				},
			},
		},
		{
			name:    "ostrich",
			version: "90.0",
			archive: "birds",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./speed/"}},
				{
					Header:  tar.Header{Name: "./speed/ostrich"},
					Content: []byte("90.0 km/h\n"),
				},
			},
		},
	},
	slices: []setup.SliceKey{
		{"cheetah", "speed"},
		{"ostrich", "speed"},
	},
	release: map[string]string{
		"chisel.yaml": `
			format: chisel-v1
			archives:
				vertebrates:
					version: 1
					suites: [main]
					components: [main, universe]
				mammals:
					version: 1
					suites: [main]
					components: [main]
				birds:
					version: 1
					suites: [main]
					components: [main]
		`,
		"slices/mydir/cheetah.yaml": `
			package: cheetah
			slices:
				speed:
					contents:
						/speed/cheetah:
		`,
		"slices/mydir/ostrich.yaml": `
			package: ostrich
			slices:
				speed:
					contents:
						/speed/ostrich:
		`,
	},
	result: map[string]string{
		"/speed/":        "dir 0755",
		"/speed/cheetah": "file 0644 e98b0879",
		"/speed/ostrich": "file 0644 c8fa2806",
	},
}, {
	summary: "Pick package from archive with highest priority",
	pkgs: []testPackageInfo{
		{
			name:    "hello",
			version: "2.0-beta",
			archive: "edge",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./hello"}, Content: []byte("Hello, The Edge\n")},
			},
		},
		{
			name:    "hello",
			version: "1.8",
			archive: "candidate",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./hello"}, Content: []byte("Hello, The Candidate\n")},
			},
		},
		{
			name:    "hello",
			version: "1.2",
			archive: "stable",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./hello"}, Content: []byte("Hello, The Stable\n")},
			},
		},
		{
			name:    "hello",
			version: "1.0",
			archive: "obsolete",
			content: []testutil.TarEntry{
				{Header: tar.Header{Name: "./"}},
				{Header: tar.Header{Name: "./hello"}, Content: []byte("Hello, The Obsolete\n")},
			},
		},
	},
	release: map[string]string{
		"chisel.yaml": `
			format: chisel-v1
			archives:
				edge:
					version: 1
					suites: [main]
					components: [main]
				candidate:
					version: 1
					suites: [main]
					components: [main]
					priority: 5
				stable:
					version: 1
					suites: [main]
					components: [main]
					priority: 10
				obsolete:
					version: 1
					suites: [main]
					components: [main]
					priority: 10
		`,
		"slices/mydir/hello.yaml": `
			package: hello
			slices:
				all:
					contents:
						/hello:
		`,
	},
	slices: []setup.SliceKey{
		{"hello", "all"},
	},
	result: map[string]string{
		"/hello": "file 0644 b5621b65",
	},
}}

const defaultChiselYaml = `
	format: chisel-v1
	archives:
		ubuntu:
			version: 22.04
			components: [main, universe]
`

type testPackage struct {
	version string
	content []byte
}

type testArchive struct {
	options archive.Options
	pkgs    map[string]testPackage
}

func (a *testArchive) Options() *archive.Options {
	return &a.options
}

func (a *testArchive) Fetch(pkg string) (io.ReadCloser, error) {
	if data, ok := a.pkgs[pkg]; ok {
		return ioutil.NopCloser(bytes.NewBuffer(data.content)), nil
	}
	return nil, fmt.Errorf("attempted to open %q package", pkg)
}

func (a *testArchive) Version(pkg string) string {
	if data, ok := a.pkgs[pkg]; ok {
		return data.version
	}
	return ""
}

func (s *S) TestRun(c *C) {
	for _, test := range slicerTests {
		c.Logf("Summary: %s", test.summary)

		if _, ok := test.release["chisel.yaml"]; !ok {
			test.release["chisel.yaml"] = string(defaultChiselYaml)
		}

		releaseDir := c.MkDir()
		for path, data := range test.release {
			fpath := filepath.Join(releaseDir, path)
			err := os.MkdirAll(filepath.Dir(fpath), 0755)
			c.Assert(err, IsNil)
			err = ioutil.WriteFile(fpath, testutil.Reindent(data), 0644)
			c.Assert(err, IsNil)
		}

		release, err := setup.ReadRelease(releaseDir)
		c.Assert(err, IsNil)

		selection, err := setup.Select(release, test.slices)
		c.Assert(err, IsNil)

		archives := map[string]archive.Archive{}
		for name, setupArchive := range release.Archives {
			archive := &testArchive{
				options: archive.Options{
					Label:      setupArchive.Name,
					Version:    setupArchive.Version,
					Suites:     setupArchive.Suites,
					Components: setupArchive.Components,
					Priority:   setupArchive.Priority,
					Pro:        setupArchive.Pro,
					Arch:       test.arch,
				},
			}
			if test.pkgs != nil {
				archive.pkgs = map[string]testPackage{}
			}
			archives[name] = archive
		}

		if test.pkgs != nil {
			for _, pkgInfo := range test.pkgs {
				archive := archives[pkgInfo.archive].(*testArchive)
				c.Assert(archive, NotNil)
				deb, err := testutil.MakeDeb(pkgInfo.content)
				c.Assert(err, IsNil)
				archive.pkgs[pkgInfo.name] = testPackage{pkgInfo.version, deb}
			}
		} else {
			pkgs := map[string]testPackage{
				"base-files": testPackage{"1", testutil.PackageData["base-files"]},
			}
			for name, entries := range packageEntries {
				deb, err := testutil.MakeDeb(entries)
				c.Assert(err, IsNil)
				pkgs[name] = testPackage{"1", deb}
			}
			archive := archives["ubuntu"].(*testArchive)
			c.Assert(archive, NotNil)
			archive.pkgs = pkgs
		}

		targetDir := c.MkDir()
		options := slicer.RunOptions{
			Selection: selection,
			Archives:  archives,
			TargetDir: targetDir,
		}
		if test.hackopt != nil {
			test.hackopt(c, &options)
		}
		err = slicer.Run(&options)
		if test.error == "" {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, ErrorMatches, test.error)
			continue
		}

		if test.result != nil {
			result := make(map[string]string, len(copyrightEntries)+len(test.result))
			if test.pkgs == nil {
				for k, v := range copyrightEntries {
					result[k] = v
				}
			}
			for k, v := range test.result {
				result[k] = v
			}
			c.Assert(testutil.TreeDump(targetDir), DeepEquals, result)
		}
	}
}
