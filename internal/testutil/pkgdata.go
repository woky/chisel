package testutil

import (
	"archive/tar"
	"bytes"
	"embed"
	"path/filepath"
	"time"

	"github.com/blakesmith/ar"
	"github.com/klauspost/compress/zstd"
)

var PackageData = map[string][]byte{}

func init() {
	for name, entries := range pkgs {
		for i, _ := range entries {
			if !entries[i].noFixup {
				fixupEntry(&entries[i])
			}
		}
		data, err := buildDeb(entries)
		if err != nil {
			panic(err)
		}
		PackageData[name] = data
	}
}

type tarEntry struct {
	header  tar.Header
	noFixup bool
	content []byte
}

var pkgs = map[string][]tarEntry{
	"base-files": {
		{
			header: tar.Header{
				Name: "./",
			},
		},
		{
			header: tar.Header{
				Name: "./bin/",
			},
		},
		{
			header: tar.Header{
				Name: "./boot/",
			},
		},
		{
			header: tar.Header{
				Name: "./dev/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/debian_version",
			},
			content: readPkgdataFile(
				"base-files/etc/debian_version",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/default/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/dpkg/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/dpkg/origins/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/dpkg/origins/debian",
			},
			content: readPkgdataFile(
				"base-files/etc/dpkg/origins/debian",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/dpkg/origins/ubuntu",
			},
			content: readPkgdataFile(
				"base-files/etc/dpkg/origins/ubuntu",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/host.conf",
			},
			content: readPkgdataFile(
				"base-files/etc/host.conf",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/issue",
			},
			content: readPkgdataFile(
				"base-files/etc/issue",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/issue.net",
			},
			content: readPkgdataFile(
				"base-files/etc/issue.net",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/legal",
			},
			content: readPkgdataFile(
				"base-files/etc/legal",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/lsb-release",
			},
			content: readPkgdataFile(
				"base-files/etc/lsb-release",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/profile.d/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/profile.d/01-locale-fix.sh",
			},
			content: readPkgdataFile(
				"base-files/etc/profile.d/01-locale-fix.sh",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/skel/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/update-motd.d/",
			},
		},
		{
			header: tar.Header{
				Name: "./etc/update-motd.d/00-header",
				Mode: 00755,
			},
			content: readPkgdataFile(
				"base-files/etc/update-motd.d/00-header",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/update-motd.d/10-help-text",
				Mode: 00755,
			},
			content: readPkgdataFile(
				"base-files/etc/update-motd.d/10-help-text",
			),
		},
		{
			header: tar.Header{
				Name: "./etc/update-motd.d/50-motd-news",
				Mode: 00755,
			},
			content: readPkgdataFile(
				"base-files/etc/update-motd.d/50-motd-news",
			),
		},
		{
			header: tar.Header{
				Name: "./home/",
			},
		},
		{
			header: tar.Header{
				Name: "./lib/",
			},
		},
		{
			header: tar.Header{
				Name: "./lib/systemd/",
			},
		},
		{
			header: tar.Header{
				Name: "./lib/systemd/system/",
			},
		},
		{
			header: tar.Header{
				Name: "./lib/systemd/system/motd-news.service",
			},
			content: readPkgdataFile(
				"base-files/lib/systemd/system/motd-news.service",
			),
		},
		{
			header: tar.Header{
				Name: "./lib/systemd/system/motd-news.timer",
			},
			content: readPkgdataFile(
				"base-files/lib/systemd/system/motd-news.timer",
			),
		},
		{
			header: tar.Header{
				Name: "./proc/",
			},
		},
		{
			header: tar.Header{
				Name: "./root/",
				Mode: 00700,
			},
		},
		{
			header: tar.Header{
				Name: "./run/",
			},
		},
		{
			header: tar.Header{
				Name: "./sbin/",
			},
		},
		{
			header: tar.Header{
				Name: "./sys/",
			},
		},
		{
			header: tar.Header{
				Name: "./tmp/",
				Mode: 01777,
			},
		},
		{
			header: tar.Header{
				Name: "./usr/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/bin/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/bin/hello",
				Mode: 00775,
			},
			content: readPkgdataFile(
				"base-files/usr/bin/hello",
			),
		},
		{
			header: tar.Header{
				Name: "./usr/games/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/include/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/lib/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/lib/os-release",
			},
			content: readPkgdataFile(
				"base-files/usr/lib/os-release",
			),
		},
		{
			header: tar.Header{
				Name: "./usr/sbin/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/dict/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/doc/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/doc/base-files/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/doc/base-files/copyright",
			},
			content: readPkgdataFile(
				"base-files/usr/share/doc/base-files/copyright",
			),
		},
		{
			header: tar.Header{
				Name: "./usr/share/info/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/man/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/share/misc/",
			},
		},
		{
			header: tar.Header{
				Name: "./usr/src/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/backups/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/cache/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/lib/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/lib/dpkg/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/lib/misc/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/local/",
				Mode: 02775,
			},
		},
		{
			header: tar.Header{
				Name: "./var/lock/",
				Mode: 01777,
			},
		},
		{
			header: tar.Header{
				Name: "./var/log/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/run/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/spool/",
			},
		},
		{
			header: tar.Header{
				Name: "./var/tmp/",
				Mode: 01777,
			},
		},
		{
			header: tar.Header{
				Name:     "./etc/os-release",
				Linkname: "../usr/lib/os-release",
			},
		},
	},
}

//go:embed all:pkgdata
var pkgdataFS embed.FS

var zeroTime time.Time
var epochStartTime time.Time = time.Unix(0, 0)

func fixupEntry(entry *tarEntry) {
	hdr := &entry.header
	if hdr.Typeflag == 0 {
		if hdr.Linkname != "" {
			hdr.Typeflag = tar.TypeSymlink
		} else if hdr.Name[len(hdr.Name)-1] == '/' {
			hdr.Typeflag = tar.TypeDir
		} else {
			hdr.Typeflag = tar.TypeReg
		}
	}
	if hdr.Mode == 0 {
		switch hdr.Typeflag {
		case tar.TypeDir:
			hdr.Mode = 0755
		case tar.TypeSymlink:
			hdr.Mode = 0777
		default:
			hdr.Mode = 0644
		}
	}
	if hdr.Size == 0 && entry.content != nil {
		hdr.Size = int64(len(entry.content))
	}
	if hdr.Uid == 0 && hdr.Uname == "" {
		hdr.Uname = "root"
	}
	if hdr.Gid == 0 && hdr.Gname == "" {
		hdr.Gname = "root"
	}
	if hdr.ModTime == zeroTime {
		hdr.ModTime = epochStartTime
	}
	if hdr.Format == 0 {
		hdr.Format = tar.FormatGNU
	}
}

func readPkgdataFile(path string) []byte {
	path = filepath.Join("pkgdata", path)
	content, err := pkgdataFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}

func makeTar(entries []tarEntry) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, entry := range entries {
		if err := tw.WriteHeader(&entry.header); err != nil {
			return nil, err
		}
		if entry.content != nil {
			if _, err := tw.Write(entry.content); err != nil {
				return nil, err
			}
		}
	}
	return buf.Bytes(), nil
}

func compressBytesZstd(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zstd.NewWriter(&buf)
	if _, err = writer.Write(input); err != nil {
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildDeb(entries []tarEntry) ([]byte, error) {
	var buf bytes.Buffer

	tarData, err := makeTar(entries)
	if err != nil {
		return nil, err
	}
	compTarData, err := compressBytesZstd(tarData)
	if err != nil {
		return nil, err
	}

	writer := ar.NewWriter(&buf)
	if err := writer.WriteGlobalHeader(); err != nil {
		return nil, err
	}
	dataHeader := ar.Header{
		Name: "data.tar.zst",
		Size: int64(len(compTarData)),
	}
	if err := writer.WriteHeader(&dataHeader); err != nil {
		return nil, err
	}
	if _, err = writer.Write(compTarData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
