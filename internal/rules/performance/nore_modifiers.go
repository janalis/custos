package performance

import (
	"regexp"
	"strings"
	"unicode"

	"custos/internal/meta"
	"custos/internal/phpver"
)

// NotOptimalRegularExpressions: modifier checks (D7–D13).

var (
	noreEscapedDotBracket = regexp.MustCompile(`\\[.\[\]]`)
	noreBracketRun        = regexp.MustCompile(`\[[^\]]+\]`)
	noreClassEscape       = regexp.MustCompile(`\\[\\dDwWsS]`)
)

func (c *noreCase) checkModifiers() {
	mods, body := c.mods, c.body
	has := func(m byte) bool { return strings.IndexByte(mods, m) >= 0 }

	if has('e') { // D7
		c.report(meta.SeverityError, "The /e flag was removed from PCRE; use a callback replacement.")
	}
	if c.fn != "preg_quote" && mods != "" { // D8
		allowed := "eimsuxADJSUX"
		if c.ctx.PHP >= phpver.PHP82 {
			allowed += "n"
		}
		if c.ctx.PHP >= phpver.PHP84 {
			allowed += "r"
		}
		for i := 0; i < len(mods); i++ {
			if strings.IndexByte(allowed, mods[i]) < 0 {
				c.report(meta.SeverityError, "'"+mods[i:i+1]+"' is not a valid PCRE modifier.")
			}
		}
	}
	if has('D') { // D9
		if has('m') {
			c.report(meta.SeverityInfo, "The /D flag has no effect together with /m.")
		}
		if body != "" && noreCountDiff(body, "$", `\$`) == 0 {
			c.report(meta.SeverityInfo, "The /D flag is pointless: the pattern has no '$'.")
		}
	}
	if has('s') && body != "" { // D10
		n := noreEscapedDotBracket.ReplaceAllString(body, "")
		n = noreBracketRun.ReplaceAllString(n, "")
		if noreCountDiff(n, ".", `\.`) == 0 {
			c.report(meta.SeverityInfo, "The /s flag is pointless: the pattern has no '.'.")
		}
	}
	if has('i') && body != "" { // D11
		n := noreClassEscape.ReplaceAllString(body, "")
		if strings.IndexFunc(n, unicode.IsLetter) < 0 {
			c.report(meta.SeverityInfo, "The /i flag is pointless: the pattern has no letters.")
		}
	}
	if has('r') { // D12
		if !has('i') {
			c.report(meta.SeverityError, "The /r flag needs /i to take effect.")
		}
		if !has('u') {
			c.report(meta.SeverityError, "The /r flag needs /u to take effect.")
		}
	}
	if !has('u') && body != "" && c.fn != "preg_quote" { // D13
		lt := noreHasLineTerminator(body)
		if !lt && strings.IndexFunc(body, func(r rune) bool { return r > 0x7f }) >= 0 {
			c.report(meta.SeverityError, "Non-ASCII characters in the pattern need the /u flag.")
		} else if !lt {
			n := strings.ReplaceAll(body, `\\`, "")
			if strings.Contains(n, `\p`) || strings.Contains(n, `\P`) || strings.Contains(n, `\X`) {
				c.report(meta.SeverityError, `Unicode escapes (\p, \P, \X) need the /u flag.`)
			}
		}
	}
}
