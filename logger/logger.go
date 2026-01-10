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

// раскрашиваем только строку форматирования или сначала форматируем, потом раскрашиваем.
// Во втором случае надо экранировать спецсимволы.
var VT100ColorizeParams bool = false

type VT100Writer struct {
	w io.Writer
}

func NewVT100Writer(w io.Writer) VT100Writer {
	return VT100Writer{w: w}
}

func (vt VT100Writer) Write(p []byte) (n int, err error) {
	s := colorizeVT100(string(p))
	if n, err := vt.w.Write([]byte(s)); err != nil {
		return n, err
	}
	return len(p), nil
}

type VT100DummyWriter struct {
	w io.Writer
}

func NewVT100DummyWriter(w io.Writer) VT100DummyWriter {
	return VT100DummyWriter{w: w}
}

func (vt VT100DummyWriter) Write(p []byte) (n int, err error) {
	s := stripVT100(string(p))
	if n, err := vt.w.Write([]byte(s)); err != nil {
		return n, err
	}
	return len(p), nil
}

func NewLogger() *os.File {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir, name := filepath.Split(path)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	log_file, err := os.OpenFile(filepath.Join(dir, name+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	var mw io.Writer
	if EchoConsole {
		if VT100Console {
			mw = io.MultiWriter(NewVT100Writer(os.Stdout), NewVT100DummyWriter(log_file))
		} else {
			mw = io.MultiWriter(os.Stdout, log_file)
		}
	} else {
		mw = io.MultiWriter(log_file)
	}
	log.SetOutput(mw)

	return log_file
}
