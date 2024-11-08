package goutils

import (
	"fmt"
	"log"
	"testing"

	"github.com/radozd/goutils/caches"
	"github.com/radozd/goutils/logger"
	"github.com/radozd/goutils/slices"
)

func TestProcessInfo(t *testing.T) {
	fmt.Println(ProcessInfo())
}

func TestTerminal(t *testing.T) {
	logger.EchoConsole = true
	l := logger.NewLogger()
	defer l.Close()

	log.Println("{test} `string` *number* to do")
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
	sl := slices.MergeSlices([]string{"1", "2", "3"}, []string{"2", "3", "4"})
	if len(sl) != 4 || sl[0] != "1" || sl[1] != "2" || sl[2] != "3" || sl[3] != "4" {
		t.Error("bad merge:", sl)
	}
}

func TestMapKeys(t *testing.T) {
	sl := slices.MapKeys(map[string]bool{"2": false, "1": false, "3": true})
	if len(sl) != 3 || sl[0] != "1" || sl[1] != "2" || sl[2] != "3" {
		t.Error("bad map:", sl)
	}
}
