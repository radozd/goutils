package text

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var uniReplacer = strings.NewReplacer(
	"\u00A0", " ", // NBSP
	"\u202F", " ", // NNBSP (narrow no-break space)
	"\u2009", " ", // THIN SPACE
	"\u2002", " ", // EN SPACE
	"\u2003", " ", // EM SPACE

	"\u2010", "-", // HYPHEN
	"\u2011", "-", // NON-BREAKING HYPHEN
	"\u2012", "-", // FIGURE DASH
	"\u2013", "-", // EN DASH
	"\u2014", "-", // EM DASH
	"\u2212", "-", // MINUS SIGN

	"“", "\"", "”", "\"", "«", "\"", "»", "\"", "„", "\"",
	"’", "'", "‘", "'", "‛", "'",

	"⁄", "/",

	"œ", "oe",
	"æ", "ae",

	"µ", "u",
	"×", "x",

	"°", " ", "º", " ", "˚", " ",
	"€", " ",
	"£", " ",
	"¥", " ",
	"®", " ",
	"©", " ",
)

var superSubReplacer = strings.NewReplacer(
	"⁰", "0", "¹", "1", "²", "2", "³", "3", "⁴", "4", "⁵", "5", "⁶", "6", "⁷", "7", "⁸", "8", "⁹", "9",
	"₀", "0", "₁", "1", "₂", "2", "₃", "3", "₄", "4", "₅", "5", "₆", "6", "₇", "7", "₈", "8", "₉", "9",
	"⁺", "+", "⁻", "-", "⁽", "(", "⁾", ")",
	"₊", "+", "₋", "-", "₍", "(", "₎", ")",
)

var vulgarFracReplacer = strings.NewReplacer(
	"½", "1/2",
	"⅓", "1/3", "⅔", "2/3",
	"¼", "1/4", "¾", "3/4",
	"⅕", "1/5", "⅖", "2/5", "⅗", "3/5", "⅘", "4/5",
	"⅙", "1/6", "⅚", "5/6",
	"⅛", "1/8", "⅜", "3/8", "⅝", "5/8", "⅞", "7/8",
)

var spacesReplacer = strings.NewReplacer(
	" ( ", "(",
	" )", ")",
	"( ", "(",
	" / ", "/",
	" /", "/",
	"/ ", "/",
	" ,", ",",
	" .", ".",
	" :", ":",
	" 000", "000",
	".000", "000",
	" \n", "\n",
)

var separateThousands *regexp.Regexp = regexp.MustCompile(`([ ();:/])(\d+) (\d{1,3}00)\b`)
var separateThousands2 *regexp.Regexp = regexp.MustCompile(`([ ();:/])(\d) (\d{3})\b`)

func NormalizeSpaces(s string) string {
	s = spacesReplacer.Replace(s)
	s = separateThousands.ReplaceAllString(s, "$1$2$3")
	s = separateThousands2.ReplaceAllString(s, "$1$2$3")
	return s
}

func StripAccents(s string) string {
	if s == "" {
		return s
	}
	decomp := norm.NFD.String(s)

	var b strings.Builder
	b.Grow(len(decomp))
	for _, r := range decomp {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return norm.NFC.String(b.String())
}

func NormalizeString(s string) string {
	s = strings.ToLower(s)
	s = uniReplacer.Replace(s)
	s = vulgarFracReplacer.Replace(s)
	s = superSubReplacer.Replace(s)
	s = StripAccents(s)
	s = NormalizeSpaces(s)
	return s
}
