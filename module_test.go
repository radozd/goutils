package goutils

import (
	"fmt"
	"log"
	"testing"

	"github.com/radozd/goutils/caches"
	"github.com/radozd/goutils/collections"
	"github.com/radozd/goutils/logger"
	"github.com/radozd/goutils/vt100"
)

func TestProcessInfo(t *testing.T) {
	fmt.Println(ProcessInfo())
}

func TestTerminal(t *testing.T) {
	logger.EchoConsole = true
	logger.VT100Console = true
	l := logger.NewLogger()

	log.Println("{test} 'string' #number# to *do*")
	log.Printf("`warning` @newer@ file at the same path: `%s`\n", "/Total *R*ekall.mp4")
	vt100.Printf("`warning` @newer@ file at the same path: `%s`\n", "/Total *R*ekall.mp4")
	vt100.ColorizeParams = true
	vt100.Printf("`warning` @newer@ file at the same path: `%s`\n", "/Total *R*ekall.mp4")
	l.Close()
}

func TestZstd(t *testing.T) {
	str := "test data test data test data test data test data test data"
	bytes, _ := caches.ZstdPack([]byte(str))
	unp, err := caches.ZstdUnpack(bytes)
	if err != nil {
		t.Error(err)
	}
	if str != string(unp) {
		t.Error("zstd: no match")
	}
	//os.WriteFile("/Users/ko/Dev/goutils.go/test3.zst", bytes, os.ModePerm)
}

func TestUniq(t *testing.T) {
	sl := collections.MergeSlices([]string{"1", "2", "3"}, []string{"2", "3", "4"})
	if len(sl) != 4 || sl[0] != "1" || sl[1] != "2" || sl[2] != "3" || sl[3] != "4" {
		t.Error("bad merge:", sl)
	}
}
