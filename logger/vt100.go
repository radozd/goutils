package logger

import (
	"fmt"

	"github.com/radozd/goutils/text"
)

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

func Color(code int) string {
	return fmt.Sprintf("\033[%dm", code)
}

const Bold string = "\033[1m"
const ResetAttr string = "\033[0m"

func colorizeVT100(s string) string {
	do := func(s string, brackets string, color int) string {
		a := string(brackets[0])
		b := string(brackets[1])

		for {
			var s2 string
			if color != BOLD {
				s2 = text.ReplaceBetweenInc(s, a, b, Color(color)+text.TakeBetween(s, a, b)+ResetAttr)
			} else {
				s2 = text.ReplaceBetweenInc(s, a, b, Bold+text.TakeBetween(s, a, b)+ResetAttr)
			}
			if s == s2 {
				break
			}
			s = s2
		}
		return s
	}

	s = do(s, "{}", BG_RED)

	s = do(s, "``", RED)
	s = do(s, "''", GREEN)
	s = do(s, "##", BLUE)
	s = do(s, "@@", YELLOW)

	s = do(s, "**", BOLD)

	return s
}

func colorizeHtml(s string) string {

	do := func(s string, brackets string, color string) string {
		a := string(brackets[0])
		b := string(brackets[1])

		for {
			s2 := text.ReplaceBetweenInc(s, a, b, "<span class=\""+color+"\">"+text.TakeBetween(s, a, b)+"</span>")
			if s == s2 {
				break
			}
			s = s2
		}
		return s
	}

	s = do(s, "{}", "BG_RED")

	s = do(s, "``", "RED")
	s = do(s, "''", "GREEN")
	s = do(s, "##", "BLUE")
	s = do(s, "@@", "YELLOW")

	s = do(s, "**", "BOLD")

	return s
}

func VT100Sprintf(format string, a ...any) string {
	return fmt.Sprintf(colorizeVT100(format), a...)
}

func VT100Printf(format string, a ...any) {
	fmt.Print(VT100Sprintf(format, a...))
}

func HtmlSprintf(format string, a ...any) string {
	return fmt.Sprintf(colorizeHtml(format), a...)
}
