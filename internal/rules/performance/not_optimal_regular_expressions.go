package performance

import (
	"strings"
	"unicode/utf8"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// notOptimalRegularExpressions runs a family of checks on PCRE patterns
// passed to the preg_* functions. The checks live in nore_*.go:
// delimiters/modifiers (nore_modifiers.go), pattern body (nore_body.go) and
// call shapes / plain string API replacements (nore_calls.go).
type notOptimalRegularExpressions struct{}

func init() { register(notOptimalRegularExpressions{}) }

func (notOptimalRegularExpressions) ID() string { return "NotOptimalRegularExpressions" }

func (notOptimalRegularExpressions) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var noreFunctions = []string{
	"preg_filter", "preg_grep", "preg_match_all", "preg_match",
	"preg_replace_callback", "preg_replace", "preg_split", "preg_quote",
}

// noreCase is one pattern literal being checked for one call.
type noreCase struct {
	ctx   *analysis.Context
	call  *syntax.FuncCall
	fn    string
	lit   syntax.Expr // the resolved literal L
	body  string
	mods  string
	array bool // array mode (D5)

	sink  *[]string       // dry run: collect literal findings instead of reporting
	prior map[string]bool // literal findings already made by an earlier call
}

func (c *noreCase) report(sev meta.Severity, msg string) {
	key := string(sev) + "\x00" + msg
	if c.sink != nil {
		*c.sink = append(*c.sink, key)
		return
	}
	if c.prior[key] {
		return
	}
	c.ctx.ReportSeverity(c.lit.Span(), sev, msg)
}

func (notOptimalRegularExpressions) Check(ctx *analysis.Context, n syntax.Node) {
	call, fn, cands, array := noreCandidates(ctx, n)
	for _, lit := range cands {
		c := &noreCase{ctx: ctx, call: call, fn: fn, lit: lit, array: array}
		if !call.Span().Contains(lit.Span()) {
			// Reached through value discovery: an earlier call using the
			// same literal already reported its literal findings.
			c.prior = noreEarlierFindings(ctx, call, lit)
		}
		if c.checkLiteral() && !array {
			c.checkCall()
		}
	}
}

// noreCandidates applies D1/D2: the call, its function name, the candidate
// literals and whether the call is in array mode.
func noreCandidates(ctx *analysis.Context, n syntax.Node) (*syntax.FuncCall, string, []syntax.Expr, bool) {
	call, fn := perfCall(ctx, n, noreFunctions...) // D1
	if call == nil || util.ArgCount(call) == 0 {
		return nil, "", nil, false
	}
	first, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || first.Value == nil {
		return nil, "", nil, false
	}
	var cands []syntax.Expr // D2
	arr, array := first.Value.(*syntax.Array)
	if array {
		for _, it := range arr.Items {
			if it == nil || it.Value == nil {
				continue
			}
			if lit := util.SingleStringLiteral(ctx.File, it.Value); lit != nil {
				cands = append(cands, lit)
			}
		}
	} else if lit := util.SingleStringLiteral(ctx.File, first.Value); lit != nil {
		cands = append(cands, lit)
	}
	return call, fn, cands, array
}

// checkLiteral runs D3–D19 on the literal; it reports whether the pattern
// parsed (so the call checks may run).
func (c *noreCase) checkLiteral() bool {
	raw, _, ok := util.QuotedStringRaw(c.lit) // D3
	if !ok || raw == "" {
		return false
	}
	body, mods, ok := noreSplitDelimiters(raw) // D4
	if !ok {
		if c.fn != "preg_quote" { // D6
			c.report(meta.SeverityWarning, "Pattern has no valid delimiters.")
		}
		return false
	}
	c.body, c.mods = body, mods
	c.checkModifiers()
	c.checkBody()
	return true
}

// noreEarlierFindings returns the literal findings that calls placed before
// call make on lit, so each finding on a shared literal is reported once.
func noreEarlierFindings(ctx *analysis.Context, call *syntax.FuncCall, lit syntax.Expr) map[string]bool {
	var keys []string
	limit := call.Span().Start
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if n.Span().Start >= limit {
			return false
		}
		if n.Kind() != syntax.KFuncCall {
			return true
		}
		other, fn, cands, array := noreCandidates(ctx, n)
		for _, l := range cands {
			if l.Span() == lit.Span() {
				d := &noreCase{ctx: ctx, call: other, fn: fn, lit: l, array: array, sink: &keys}
				d.checkLiteral()
			}
		}
		return true
	})
	if len(keys) == 0 {
		return nil
	}
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

// noreSplitDelimiters implements D4: it returns the pattern body and the
// trailing modifier letters.
func noreSplitDelimiters(r string) (body, mods string, ok bool) {
	// PHP skips leading whitespace, then rejects alphanumeric, backslash and
	// NUL delimiters.
	r = strings.TrimLeft(r, " \t\n\v\f\r")
	if r == "" {
		return "", "", false
	}
	open := r[0]
	if open == '\\' || open == 0 || isASCIILetter(open) || open >= '0' && open <= '9' {
		return "", "", false
	}
	openLen := 1
	var close string
	switch open {
	case '{':
		close = "}"
	case '<':
		close = ">"
	case '(':
		close = ")"
	case '[':
		close = "]"
	default:
		// the delimiter may be a multi-byte character
		_, openLen = utf8.DecodeRuneInString(r)
		if openLen >= len(r) {
			return "", "", false
		}
		close = r[:openLen]
	}
	// t = start of the maximal run of trailing ASCII letters
	t := len(r)
	for t > 0 && isASCIILetter(r[t-1]) {
		t--
	}
	lo := t - len(close)
	if lo < openLen {
		lo = openLen
	}
	// the last occurrence of close starting at or after lo: its remaining
	// suffix lies in the trailing letter run (j+len(close) >= t)
	for j := len(r) - len(close); j >= lo; j-- {
		if r[j:j+len(close)] == close {
			return r[openLen:j], r[j+len(close):], true
		}
	}
	return "", "", false
}

func isASCIILetter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }

// noreHasLineTerminator reports whether s contains \n, \r, U+0085, U+2028
// or U+2029.
func noreHasLineTerminator(s string) bool {
	return strings.ContainsAny(s, "\n\r\u0085  ")
}

// noreCountDiff returns count(a) - count(b) (non-overlapping occurrences).
func noreCountDiff(s, a, b string) int { return strings.Count(s, a) - strings.Count(s, b) }
