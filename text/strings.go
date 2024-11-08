package text

import (
	"regexp"
	"strings"
	"unicode"
)

func StringAfterPrefix(s string, prefix string) string {
	if prefix == "" {
		return s
	}

	if i := strings.Index(s, prefix); i >= 0 {
		return s[i+len(prefix):]
	}
	return ""
}

func StringBeforeSuffix(s string, suffix string) string {
	if suffix == "" {
		return s
	}

	if i := strings.Index(s, suffix); i > 0 {
		return s[:i]
	}
	return ""
}

func TextContainsFullSubstring(text string, substr string) bool {
	words := strings.Fields(text)
	substr_words := strings.Fields(substr)

	if len(substr_words) > len(words) {
		return false
	}

	for i := 0; i <= len(words)-len(substr_words); i++ {
		match := true
		for j := 0; j < len(substr_words); j++ {
			if words[i+j] != substr_words[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TextContainsAnyFullSubstring(text string, substr []string) bool {
	for _, w := range substr {
		if TextContainsFullSubstring(text, w) {
			return true
		}
	}
	return false
}

func StringInArray(s string, variants []string) bool {
	for _, v := range variants {
		if s == v {
			return true
		}
	}
	return false
}

func MaxCommonString(a, b string) string {
	l := len([]rune(a))
	if l > len([]rune(b)) {
		l = len([]rune(b))
	}

	var dst string
	for i := 0; i < l; i++ {
		if []rune(a)[i] == []rune(b)[i] {
			dst += string([]rune(a)[i])
			continue
		}
		break
	}
	return strings.TrimSpace(dst)
}

var multispace *regexp.Regexp = regexp.MustCompile(`\s+`)

func RemoveMultiSpaces(s string) string {
	return strings.TrimSpace(multispace.ReplaceAllString(s, " "))
}

func CompressString(s string) string {
	return strings.TrimSpace(RemoveMultiSpaces(s))
}

func replaceChars(s string, charSet string, new string) string {
	for _, ch := range charSet {
		s = strings.ReplaceAll(s, string(ch), new)
	}
	return s
}

func insertRune(slice []rune, r rune, index int) []rune {
	return append(slice[:index], append([]rune{r}, slice[index:]...)...)
}

const DefaultGarbageReplace string = `!"#$&'*+,-/:;<=>?@[\]_{|}«°»—“”€`
const FilteredGarbageReplace string = `!"#$&'*,-:;<=>?@[\]_{|}«°»—“”€`

func preprocessTextToSplit(text string, replace string) string {
	tmp := strings.ReplaceAll(text, ",", ".") // 1,23 -> 1.23

	tmp = replaceChars(tmp, replace, " ") // убираем мусор, который точно не понадобится

	// в результате, от
	// "abc, de,11.2 po1nts. (to) 2 robots, 1.2 picas2/3. 3mg."
	// остается что-то типа
	// "abc. de.11.2 po1nts. (to) 2 robots. 1.2 picas2 3. 3mg"

	runes := []rune(tmp)
	for i := 1; i < len(runes)-1; i++ { // убираем лишние точки, которые не являются десятичными
		if runes[i] == '.' {
			if unicode.IsDigit(runes[i-1]) && unicode.IsDigit(runes[i+1]) {
				continue
			}
			runes[i] = ' '
		}
	}

	// а теперь так
	// "abc  de 11.2 po1nts   to  2 robots  1.2 picas2 3  3mg"

	return string(runes)
}

// и слова и числа. десятичные запятые превращаются в точки
func SplitToPartsEx(text string, replace string) []string {
	tmp := preprocessTextToSplit(text, replace)
	tmp = replaceChars(tmp, "()", " ") // убираем скобки тоже, они тут не используются
	tmp = strings.Trim(tmp, ".")

	runes := []rune(tmp)

	// надо отделить число от слова: 3mg
	letter := func(r rune) bool { return r != '.' && !unicode.IsDigit(r) && !unicode.IsSpace(r) }

x:
	for i := 1; i < len(runes)-1; i++ {
		if !unicode.IsDigit(runes[i]) {
			continue
		}

		p := runes[i-1]
		n := runes[i+1]
		if !letter(p) && letter(n) {
			runes = insertRune(runes, ' ', i+1)
			goto x
		}

		// компрессы и всякое подобное почти всегда имеет части вида 2х2
		// (или 2 х 2, или 2cm x 2cm, или 7,5x7,5 cm, или 10*10 cm)
		// жуткий случай вида 10cmx4m не обрабатываем
		if i > 1 && p == 'x' && unicode.IsSpace(runes[i-2]) {
			runes = insertRune(runes, ' ', i)
			goto x
		}
	}
	// и получается
	// "abc  de 11.2 po1nts   to  2 robots  1.2 picas2 3  3 mg"

	tmp = strings.TrimSpace(string(runes))
	return strings.Fields(tmp)
}

func SplitToParts(text string) []string {
	return SplitToPartsEx(text, DefaultGarbageReplace)
}

// знаки и числа вырезаются
// minWordLen убирает короткие слова
func SplitToWords(text string, minWordLen int) []string {
	parts := SplitToParts(text)

	words := make([]string, 0, len(parts))
	for _, s := range parts {
		rs := []rune(s)
		if len(rs) >= minWordLen && !unicode.IsDigit(rs[0]) {
			words = append(words, s)
		}
	}
	return words
}

func NormalizeUnicode(s string) string {
	s = strings.ReplaceAll(s, "µ", "u")
	s = strings.ReplaceAll(s, "½", "1/2")
	s = strings.ReplaceAll(s, "⅓", "1/3")
	s = strings.ReplaceAll(s, "²", "2")
	s = strings.ReplaceAll(s, "³", "3")
	s = strings.ReplaceAll(s, "œ", "oe")
	s = strings.ReplaceAll(s, "æ", "ae")
	s = strings.ReplaceAll(s, "’", "'")
	s = strings.ReplaceAll(s, "×", "x")
	s = replaceChars(s, "“”«»", "\"")
	s = replaceChars(s, "°€®©", " ")
	return s
}

func NormalizeSpaces(s string) string {
	s = strings.ReplaceAll(s, "( ", "(")
	s = strings.ReplaceAll(s, " )", ")")
	s = strings.ReplaceAll(s, " /", "/")
	s = strings.ReplaceAll(s, "/ ", "/")
	s = strings.ReplaceAll(s, " ,", ",")
	s = strings.ReplaceAll(s, " .", ".")
	s = strings.ReplaceAll(s, " :", ":")
	return s
}

func NormalizeString(s string) string {
	s = strings.ToLower(s)
	s = NormalizeUnicode(s)
	s = NormalizeSpaces(s)
	return s
}

// берет содержимое строки между `start` и `end`. если `end` пуст, то берет до конца строки
func TakeBetween(text string, start string, end string) string {
	i1 := strings.Index(text, start)
	if i1 < 0 {
		return ""
	}
	i1 = i1 + len(start)
	i2 := len(text) - i1
	if end != "" {
		i2 = strings.Index(text[i1:], end)
		if i2 < 0 {
			return ""
		}
	}
	return text[i1 : i1+i2]
}

// заменяет содержимое между `start` и `end` на новое
func ReplaceBetween(text string, start string, end string, repl string) string {
	i1 := strings.Index(text, start)
	if i1 < 0 {
		return text
	}
	i1 = i1 + len(start)
	i2 := strings.Index(text[i1:], end)
	if i2 < 0 {
		return text
	}
	return text[:i1] + repl + text[i1+i2:]
}

// вырезает из строки содержимое между `start` и `end`
func RemoveBetween(text string, start string, end string) string {
	return ReplaceBetween(text, start, end, "")
}

// заменяет содержимое между `start` и `end` на новое, вырезая `start` и `end`
func ReplaceBetweenInc(text string, start string, end string, repl string) string {
	i1 := strings.Index(text, start)
	if i1 < 0 {
		return text
	}
	i1 = i1 + len(start)
	i2 := strings.Index(text[i1:], end)
	if i2 < 0 {
		return text
	}
	return text[:i1-len(start)] + repl + text[i1+i2+len(end):]
}

// вырезает из строки содержимое между `start` и `end`, включая `start` и `end`
func RemoveBetweenInc(text string, start string, end string) string {
	return ReplaceBetweenInc(text, start, end, "")
}

func ReplaceWhole(s string, what string, with string) string {
	if i1 := strings.Index(s, what); i1 >= 0 {
		i2 := i1 + len(what)
		good_start := i1 == 0 || strings.Contains(" \n(,./", string(s[i1-1]))
		good_end := i2 == len(s) || strings.Contains(" \n),./", string(s[i2]))

		if good_start && good_end {
			s = s[:i1] + with + s[i2:]
		}
	}
	return s
}
