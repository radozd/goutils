package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// в консоль будет писаться то же, что и в лог
var EchoConsole bool = false

// раскрашиваем вывод в консоль тегами. в логе теги остаются
var VT100Console bool = false

type VT100Writer struct {
	w io.Writer
}

func NewVT100Writer(w io.Writer) VT100Writer {
	return VT100Writer{w: w}
}

func (vt VT100Writer) Write(p []byte) (n int, err error) {
	s := colorizeVT100(string(p))
	return vt.w.Write([]byte(s))
}

func NewLogger() *os.File {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir, name := filepath.Split(path)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	logFile, err := os.OpenFile(filepath.Join(dir, name+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	var mw io.Writer
	if EchoConsole {
		if VT100Console {
			mw = io.MultiWriter(NewVT100Writer(os.Stdout), logFile)
		} else {
			mw = io.MultiWriter(os.Stdout, logFile)
		}
	} else {
		mw = io.MultiWriter(logFile)
	}
	log.SetOutput(mw)

	return logFile
}
