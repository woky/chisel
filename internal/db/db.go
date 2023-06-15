package db

import (
	"os"
	"path/filepath"

	"github.com/canonical/chisel/internal/jsonwall"
	"github.com/klauspost/compress/zstd"
)

const schema = "0.1"

func New() *jsonwall.DBWriter {
	options := jsonwall.DBWriterOptions{Schema: schema}
	return jsonwall.NewDBWriter(&options)
}

func dbPath(root string) string {
	return filepath.Join(root, ".chisel.db")
}

func Save(dbw *jsonwall.DBWriter, root string) error {
	f, err := os.OpenFile(dbPath(root), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	// chmod the existing file
	if err := f.Chmod(0644); err != nil {
		return err
	}
	zw, err := zstd.NewWriter(f)
	if err != nil {
		return err
	}
	if _, err := dbw.WriteTo(zw); err != nil {
		return err
	}
	zw.Close()
	return nil
}

func Load(root string) (*jsonwall.DB, error) {
	f, err := os.Open(dbPath(root))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := zstd.NewReader(f)
	if err != nil {
		return nil, err
	}
	db, err := jsonwall.ReadDB(zr)
	if err != nil {
		return nil, err
	}
	return db, nil
}
