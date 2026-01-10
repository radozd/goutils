package vt100

import (
	"fmt"
	"strings"

	"github.com/radozd/goutils/text"
)

// раскрашиваем только строку форматирования или сначала форматируем, потом раскрашиваем.
// Во втором случае надо экранировать спецсимволы.
var ColorizeParams bool = false

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

func StripMarkers(s string) string {
	do := func(s string, brackets string) string {
		a := string(brackets[0])
		b := string(brackets[1])

		for {
			s2 := text.ReplaceBetweenInc(s, a, b, text.TakeBetween(s, a, b))
			if s == s2 {
				break
			}
			s = s2
		}
		return s
	}

	s = do(s, "{}")

	s = do(s, "``")
	s = do(s, "''")
	s = do(s, "##")
	s = do(s, "@@")

	s = do(s, "**")

	return s
}

func EscapeMarkers(s string) string {
	repl := []struct{ from, to string }{
		{"*", "∗"}, // U+2217 ASTERISK OPERATOR
		{"#", "＃"}, // U+FF03 FULLWIDTH NUMBER SIGN
		{"'", "’"}, // U+2019 RIGHT SINGLE QUOTATION MARK
		{"{", "｛"}, // U+FF5B FULLWIDTH LEFT CURLY BRACKET
		{"}", "｝"}, // U+FF5D FULLWIDTH RIGHT CURLY BRACKET
	}
	for _, r := range repl {
		if strings.Contains(s, r.from) {
			s = strings.ReplaceAll(s, r.from, r.to)
		}
	}
	return s
}

func ColorizeVT100(s string) string {
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

func ColorizeHtml(s string) string {
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

func Sprintf(format string, a ...any) string {
	if ColorizeParams {
		s := fmt.Sprintf(format, a...)
		return ColorizeVT100(s)
	}
	return fmt.Sprintf(ColorizeVT100(format), a...)
}

func Printf(format string, a ...any) {
	fmt.Print(Sprintf(format, a...))
}

func HtmlSprintf(format string, a ...any) string {
	if ColorizeParams {
		s := fmt.Sprintf(format, a...)
		return ColorizeHtml(s)
	}
	return fmt.Sprintf(ColorizeHtml(format), a...)
}
