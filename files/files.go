package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func AppDir() string {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(path)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsEmptyDirectory(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func FormatSize(size uint64) string {
	sizeMB := (int)(size / 1024 / 1024)
	if sizeMB < 1024 {
		return strconv.Itoa(sizeMB) + "MB"
	}
	if sizeMB < 1024*1024 {
		return strconv.Itoa(int(sizeMB/1024)) + "GB"
	}

	return fmt.Sprintf("%.1fTB", float64(sizeMB)/1024/1024)
}
