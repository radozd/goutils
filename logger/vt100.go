package logger

import (
	"fmt"

	"github.com/radozd/goutils/text"
)

// List of possible colors
const (
	BLACK      = 30
	RED        = 31
	GREEN      = 32
	YELLOW     = 33
	BLUE       = 34
	MAGENTA    = 35
	CYAN       = 36
	WHITE      = 37
	BG_BLACK   = 40
	BG_RED     = 41
	BG_GREEN   = 42
	BG_YELLOW  = 43
	BG_BLUE    = 44
	BG_MAGENTA = 45
	BG_CYAN    = 46
	BG_WHITE   = 47
	BOLD       = 1000
)

func getColor(code int) string {
	return fmt.Sprintf("\033[%dm", code)
}

const setBold string = "\033[1m"
const resetAttr string = "\033[0m"

func colorizeVT100(s string) string {
	do := func(s string, brackets string, color int) string {
		a := string(brackets[0])
		b := string(brackets[1])

		for {
			var s2 string
			if color != BOLD {
				s2 = text.ReplaceBetweenInc(s, a, b, getColor(color)+text.TakeBetween(s, a, b)+resetAttr)
			} else {
				s2 = text.ReplaceBetweenInc(s, a, b, setBold+text.TakeBetween(s, a, b)+resetAttr)
			}
			if s == s2 {
				break
			}
			s = s2
		}
		return s
	}

	s = do(s, "{}", BG_RED) // error
	s = do(s, "[]", RED)    // important
	s = do(s, "``", GREEN)  // string
	s = do(s, "##", BLUE)   // number

	s = do(s, "**", BOLD)

	return s
}

func V100Sprintf(format string, a ...any) string {
	s := fmt.Sprintf(format, a...)
	return colorizeVT100(s)
}

func V100Printf(format string, a ...any) {
	fmt.Print(V100Sprintf(format, a...))
}
