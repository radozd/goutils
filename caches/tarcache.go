package caches

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/radozd/goutils/files"
)

type TarCache struct {
	Name string
	file *os.File
	lock sync.Mutex
}

func NewTarCache(fname string) *TarCache {
	return &TarCache{
		Name: fname,
	}
}

func (c *TarCache) Open() error {
	if !files.Exists(c.Name) {
		f, err := os.Create(c.Name)
		if err != nil {
			return err
		}
		f.Close()
	}

	f, err := os.OpenFile(c.Name, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	c.file = f
	return nil
}

func (c *TarCache) Close() error {
	return c.file.Close()
}

func bufEmpty(buf []byte) bool {
	for _, v := range buf {
		if v != 0 {
			return false
		}
	}
	return true
}

// We might have zero, one or two end blocks.
// The standard is two, but we should try to handle other cases.
func (c *TarCache) seekToAppend() error {
	fi, err := c.file.Stat()
	if err != nil {
		return err
	}
	if fi.Size() < 1024 {
		return nil
	}

	buf := make([]byte, 512)

	if _, err = c.file.Seek(-1024, io.SeekEnd); err != nil {
		return err
	}
	if num, err := c.file.Read(buf); num != 512 || err != nil {
		return err
	}
	if bufEmpty(buf) {
		c.file.Seek(-1024, io.SeekEnd)
		return nil
	}

	if num, err := c.file.Read(buf); num != 512 || err != nil {
		return err
	}
	if bufEmpty(buf) {
		c.file.Seek(-512, io.SeekEnd)
	}
	return nil
}

func (c *TarCache) PutFile(path string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		return err
	}

	if err = c.seekToAppend(); err != nil {
		return err
	}

	// add file
	tw := tar.NewWriter(c.file)

	header := &tar.Header{
		Name:    filepath.Base(path),
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := io.Copy(tw, file); err != nil {
		return err
	}
	return tw.Close()
}

func (c *TarCache) PutBytes(path string, data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if err := c.seekToAppend(); err != nil {
		return nil
	}

	// add file
	tw := tar.NewWriter(c.file)

	header := &tar.Header{
		Name:    path,
		Size:    int64(len(data)),
		Mode:    0600,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(data); err != nil {
		return err
	}
	return tw.Close()
}

func (c *TarCache) GetBytes(path string, first bool) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, err := c.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var buf []byte = nil
	tr := tar.NewReader(c.file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hdr.Name == path {
			buf, _ = io.ReadAll(tr)
			if first {
				break
			}
		}
	}
	return buf, nil
}

func (c *TarCache) ListFiles() ([]string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, err := c.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	dic := make(map[string]bool)
	tr := tar.NewReader(c.file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		dic[hdr.Name] = true
	}

	list := make([]string, 0)
	for f := range dic {
		list = append(list, f)
	}
	return list, nil
}

func (c *TarCache) MemCache() (map[string][]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, err := c.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	cache := make(map[string][]byte)
	tr := tar.NewReader(c.file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err == nil {
			cache[hdr.Name], err = io.ReadAll(tr)
		}
		if err != nil {
			return nil, err
		}
	}

	return cache, nil
}
