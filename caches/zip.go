package caches

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

// CreateZipFile creates new zip file from list
func CreateZipFile(zipPath string, files []string) error {
	archive, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	zw := zip.NewWriter(archive)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		w, err := zw.Create(filepath.Base(file))
		if err != nil {
			f.Close()
			return err
		}
		if _, err := io.Copy(w, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	if err = zw.Close(); err != nil {
		return err
	}

	return nil
}

func ZlibPack(buf []byte) ([]byte, error) {
	var err error
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err = w.Write(buf); err != nil {
		w.Close()
		return nil, err
	}
	err = w.Close()
	return b.Bytes(), err
}

func ZlibUnpack(buf []byte) ([]byte, error) {
	var err error
	b := bytes.NewReader(buf)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if _, err = io.Copy(&out, r); err != nil {
		r.Close()
		return nil, err
	}
	err = r.Close()
	return out.Bytes(), err
}

func ZstdPack(buf []byte) ([]byte, error) {
	// Unless SingleSegment is set, framessizes < 256 are nto stored.
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression), zstd.WithSingleSegment(len(buf) < 256))
	if err != nil {
		return nil, err
	}
	return enc.EncodeAll(buf, nil), nil
}

func ZstdUnpack(buf []byte) ([]byte, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	return dec.DecodeAll(buf, nil)
}
