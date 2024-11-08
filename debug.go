package goutils

import (
	"fmt"
	"runtime"

	"github.com/radozd/goutils/files"
)

func ProcessInfo() string {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	return fmt.Sprint(
		"Version     : ", runtime.Version(), "\n",
		"NumCPU      : ", runtime.NumCPU(), "\n",
		"NumGoroutine: ", runtime.NumGoroutine(), "\n",
		"HeapObjects : ", m.HeapObjects, "\n",
		"HeapAlloc   : ", files.FormatSize(m.HeapAlloc), "\n",
		"AppDir      : ", files.AppDir())
}
