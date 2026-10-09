package notoptimalregularexpressions

import (
	"regexp"
	"strings"
	"unicode"

	"custos/internal/diagnostic"
	phpversion "custos/internal/php/version"
)

var noreClassEscape = regexp.MustCompile(`\\[\\dDwWsS]`)

func (c *noreCase) checkModifiers() {
	mods, body := c.mods, c.body
	pat := c.decoded // D9–D11 count the characters PCRE sees
	has := func(m byte) bool { return strings.IndexByte(mods, m) >= 0 }

	if has('e') { // D7
		c.report(diagnostic.SeverityError, "The /e flag was removed from PCRE; use a callback replacement.")
	}
	if c.fn != "preg_quote" && mods != "" { // D8
		allowed := "eimsuxADJSUX"
		if c.ctx.PHP >= phpversion.PHP82 {
			allowed += "n"
		}
		if c.ctx.PHP >= phpversion.PHP84 {
			allowed += "r"
		}
		for i := 0; i < len(mods); i++ {
			if strings.IndexByte(allowed, mods[i]) < 0 {
				c.report(diagnostic.SeverityError, "'"+mods[i:i+1]+"' is not a valid PCRE modifier.")
			}
		}
	}
	if has('D') { // D9
		if has('m') {
			c.report(diagnostic.SeverityInfo, "The /D flag has no effect together with /m.")
		}
		if pat != "" && !noreHasMeta(pat, '$') {
			c.report(diagnostic.SeverityInfo, "The /D flag is pointless: the pattern has no '$'.")
		}
	}
	if has('s') && pat != "" { // D10
		if !noreHasMeta(pat, '.') {
			c.report(diagnostic.SeverityInfo, "The /s flag is pointless: the pattern has no '.'.")
		}
	}
	if has('i') && pat != "" { // D11
		n := noreClassEscape.ReplaceAllString(pat, "")
		if strings.IndexFunc(n, unicode.IsLetter) < 0 {
			c.report(diagnostic.SeverityInfo, "The /i flag is pointless: the pattern has no letters.")
		}
	}
	if has('r') { // D12
		if !has('i') {
			c.report(diagnostic.SeverityError, "The /r flag needs /i to take effect.")
		}
		if !has('u') {
			c.report(diagnostic.SeverityError, "The /r flag needs /u to take effect.")
		}
	}
	if !has('u') && body != "" && c.fn != "preg_quote" { // D13
		lt := noreHasLineTerminator(body)
		if !lt && noreNonASCIIUnsafe(body, has('i')) {
			c.report(diagnostic.SeverityError, "Non-ASCII characters in the pattern need the /u flag.")
		} else if !lt {
			// custos: escapes are counted in the decoded pattern: the
			// source text '/A\\\P/' is the pattern A\\P (an escaped
			// backslash, then P), not a \P escape.
			src := body
			if pat != "" {
				src = pat
			}
			n := strings.ReplaceAll(src, `\\`, "")
			if strings.Contains(n, `\p`) || strings.Contains(n, `\P`) || strings.Contains(n, `\X`) {
				c.report(diagnostic.SeverityError, `Unicode escapes (\p, \P, \X) need the /u flag.`)
			}
		}
	}
}

// noreHasMeta reports whether the decoded pattern uses want as a
// metacharacter (D9b, D10): outside a character class, not escaped (each
// backslash escapes the next character, so `\\.` is an escaped backslash
// then a dot) and not inside a \Q…\E literal run.
func noreHasMeta(pat string, want byte) bool {
	inClass := false
	for i := 0; i < len(pat); i++ {
		ch := pat[i]
		switch {
		case ch == '\\' && i+1 < len(pat) && pat[i+1] == 'Q':
			end := strings.Index(pat[i+2:], `\E`)
			if end < 0 {
				return false
			}
			i += end + 3
		case ch == '\\':
			i++
		case inClass:
			if ch == '[' && i+1 < len(pat) && pat[i+1] == ':' {
				if end := strings.Index(pat[i+2:], ":]"); end >= 0 {
					i += end + 3 // a POSIX class [:name:]
				}
			} else if ch == ']' {
				inClass = false
			}
		case ch == '[':
			inClass = true
			if i+1 < len(pat) && pat[i+1] == '^' {
				i++
			}
			if i+1 < len(pat) && pat[i+1] == ']' {
				i++ // a leading ] is a member
			}
		case ch == want:
			return true
		}
	}
	return false
}

// noreNonASCIIUnsafe reports whether a non-ASCII character of body behaves
// differently without /u (D13a, custos): inside a character class (each
// byte becomes a class member), followed by a quantifier (it applies to
// the last byte only), or a letter under /i (no case folding without /u).
// A plain literal sequence matches byte for byte either way.
func noreNonASCIIUnsafe(body string, caseless bool) bool {
	inClass := false
	classStart := -1
	rs := []rune(body)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == '\\':
			i++
			continue
		case !inClass && r == '[':
			inClass, classStart = true, i
			if i+1 < len(rs) && rs[i+1] == '^' {
				i++
				classStart = i
			}
			continue
		case inClass && r == ']' && i != classStart+1:
			inClass = false
			continue
		}
		if r <= 0x7f {
			continue
		}
		if inClass || caseless && unicode.IsLetter(r) {
			return true
		}
		if i+1 < len(rs) && strings.ContainsRune("*+?{", rs[i+1]) {
			return true
		}
	}
	return false
}
