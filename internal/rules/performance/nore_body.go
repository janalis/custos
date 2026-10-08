package performance

import (
	"regexp"
	"strings"

	"custos/internal/meta"
)

// NotOptimalRegularExpressions: pattern body checks (D14–D19).

// noreShortClasses lists the D14 keys and their shorthand, in report order.
var noreShortClasses = [][2]string{
	{`[0-9]`, `\d`}, {`[:digit:]`, `\d`},
	{`[^0-9]`, `\D`}, {`[^\d]`, `\D`},
	{`[:word:]`, `\w`}, {`[A-Za-z0-9_]`, `\w`},
	{`[^\w]`, `\W`}, {`[^A-Za-z0-9_]`, `\W`},
	{`[^\s]`, `\S`},
}

var (
	noreSimpleGroup = regexp.MustCompile(`\[[^\[\]]+\]`)
	noreInnerGroup  = regexp.MustCompile(`([^\\])\([^()]*[^()\\]\)([^+*])`)
	noreOuterGroup  = regexp.MustCompile(`(?:^|[^>])\(([^()]*)\)([+*])(?:$|[^+])`)
	noreTagDot      = regexp.MustCompile(`>\.[*+]\??<`)
)

// noreNormalize returns B' (D14): class ranges spelled in canonical order.
func noreNormalize(body string) string {
	b := strings.ReplaceAll(body, "a-zA-Z", "A-Za-z")
	return strings.ReplaceAll(b, "0-9A-Za-z", "A-Za-z0-9")
}

func (c *noreCase) checkBody() {
	body, mods := c.body, c.mods
	has := func(m byte) bool { return strings.IndexByte(mods, m) >= 0 }

	if body != "" { // D14
		norm := noreNormalize(body)
		hint := "same result without /u"
		if has('u') {
			hint = "matches more under /u"
		}
		for _, kv := range noreShortClasses {
			if strings.Contains(norm, kv[0]) {
				h := hint
				if strings.HasPrefix(kv[0], `[^\`) {
					h = "same result" // a negated shorthand is its complement in every mode
				}
				c.report(meta.SeverityInfo, "Write '"+kv[0]+"' as '"+kv[1]+"' ("+h+").")
			}
		}
	}
	if strings.IndexByte(body, '[') >= 0 { // D15
		if run, class, ok := noreRepeatedClass(body); ok {
			c.report(meta.SeverityInfo, "Collapse '"+run+"' into '"+class+"' with a counted quantifier.")
		}
	}
	if strings.HasPrefix(c.fn, "preg_match") && len(c.call.Args.Args) == 2 && body != "" &&
		noreCountDiff(body, `\0`, `\\0`) <= 0 { // D16
		if strings.HasPrefix(body, ".*") {
			c.report(meta.SeverityInfo, "Drop the leading '.*'; it does not change whether the pattern matches.")
		}
		if strings.HasSuffix(body, ".*") {
			c.report(meta.SeverityInfo, "Drop the trailing '.*'; it does not change whether the pattern matches.")
		}
	}
	for _, g := range noreSimpleGroup.FindAllString(body, -1) { // D17
		s := g[1 : len(g)-1]
		switch {
		case strings.Contains(s, `\w`) && strings.Contains(s, `\d`):
			c.report(meta.SeverityError, "Class ["+s+`] is redundant: \d is already covered by \w.`)
		case strings.Contains(s, `\W`) && strings.Contains(s, `\D`):
			c.report(meta.SeverityError, "Class ["+s+`] is redundant: \D is already covered by \W.`)
		}
	}
	if body != "" && len(body) <= noreNestedMaxLen { // D18
		for _, f := range noreNestedQuantifiers(body) {
			c.report(meta.SeverityError, "Nested quantifier ("+f[0]+")"+f[1]+" risks catastrophic backtracking.")
		}
	}
	if !has('s') && strings.IndexByte(body, '>') >= 0 && !noreHasLineTerminator(body) && noreTagDot.MatchString(body) { // D19
		c.report(meta.SeverityInfo, "Tag content matched with '.' likely needs the /s flag.")
	}
}

// D18 folds innermost groups one pass at a time (quadratic in the number of
// groups), so a huge hostile pattern could stall the analysis: patterns
// longer than noreNestedMaxLen, or needing more than noreNestedMaxFolds
// folds, are not examined. Real patterns stay far below both.
const (
	noreNestedMaxLen   = 64 << 10
	noreNestedMaxFolds = 256
)

// noreNestedQuantifiers implements D18 and returns (alternative, quantifier)
// pairs.
func noreNestedQuantifiers(body string) [][2]string {
	n := strings.ReplaceAll(body, "(?:", "(")
	for folds := 0; ; folds++ {
		m := noreInnerGroup.FindStringSubmatch(n)
		if m == nil {
			break
		}
		if folds == noreNestedMaxFolds {
			return nil
		}
		// custos: a mandatory inner group stays as a placeholder atom
		// (`(\s+(?:a|b))*` is not `(\s+)*`); only an optional one (`?`)
		// may vanish.
		keep := "#"
		if m[2] == "?" {
			keep = ""
		}
		n = strings.ReplaceAll(n, m[0], m[1]+keep+m[2])
	}
	var out [][2]string
	for _, m := range noreOuterGroup.FindAllStringSubmatch(n, -1) {
		for _, alt := range strings.Split(m[1], "|") {
			if len(alt) == 3 && alt[0] == '\\' && strings.IndexByte("dDwWsS", alt[1]) >= 0 && (alt[2] == '*' || alt[2] == '+') {
				out = append(out, [2]string{alt, m[2]})
				break
			}
		}
	}
	return out
}

// noreScanner answers "next ']' / '}' at or after i" in O(1), so that D15
// stays linear: scanning ahead from every '[' was quadratic on patterns
// with many unclosed brackets or braces.
type noreScanner struct {
	s                    string
	nextClose, nextBrace []int32 // -1: none
}

func newNoreScanner(s string) *noreScanner {
	sc := &noreScanner{s: s, nextClose: make([]int32, len(s)+1), nextBrace: make([]int32, len(s)+1)}
	c, b := int32(-1), int32(-1)
	sc.nextClose[len(s)], sc.nextBrace[len(s)] = -1, -1
	for i := len(s) - 1; i >= 0; i-- {
		switch s[i] {
		case ']':
			c = int32(i)
		case '}':
			b = int32(i)
		}
		sc.nextClose[i], sc.nextBrace[i] = c, b
	}
	return sc
}

// classAt matches `[X]` (X: one or more non-`]`) at i and returns its end.
func (sc *noreScanner) classAt(i int) (int, bool) {
	if i >= len(sc.s) || sc.s[i] != '[' {
		return 0, false
	}
	j := int(sc.nextClose[i+1])
	if j <= i+1 {
		return 0, false
	}
	return j + 1, true
}

// quantAt matches an optional quantifier (`*`, `+`, `?`, `{…}`) at i.
func (sc *noreScanner) quantAt(i int) int {
	if i >= len(sc.s) {
		return i
	}
	switch sc.s[i] {
	case '*', '+', '?':
		return i + 1
	case '{':
		if j := int(sc.nextBrace[i+1]); j > i+1 {
			return j + 1
		}
	}
	return i
}

// repetition matches `[X]q?[X]q?` at i with both classes identical.
func (sc *noreScanner) repetition(i int) (end int, class string, ok bool) {
	e1, ok := sc.classAt(i)
	if !ok {
		return 0, "", false
	}
	s := sc.s
	class = s[i:e1]
	j := sc.quantAt(e1)
	// An identical class at j closes exactly len(class) bytes later: check
	// that in O(1) before comparing the text.
	if e2, ok := sc.classAt(j); !ok || e2-j != len(class) || s[j:e2] != class {
		return 0, "", false
	}
	return sc.quantAt(j + len(class)), class, true
}

// noreRepeatedClass implements D15: the leftmost run of one or more
// repetitions, and the class of its last repetition.
func noreRepeatedClass(s string) (run, class string, ok bool) {
	sc := newNoreScanner(s)
	for i := 0; i < len(s); i++ {
		end, cl, ok := sc.repetition(i)
		if !ok {
			continue
		}
		for {
			e2, c2, ok := sc.repetition(end)
			if !ok {
				break
			}
			end, cl = e2, c2
		}
		return s[i:end], cl, true
	}
	return "", "", false
}
