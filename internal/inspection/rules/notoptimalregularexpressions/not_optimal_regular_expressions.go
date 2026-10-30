package notoptimalregularexpressions

import (
	"strings"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// notOptimalRegularExpressions runs a family of checks on PCRE patterns
// passed to the preg_* functions. The checks live in nore_*.go:
// delimiters/modifiers (nore_modifiers.go), pattern body (nore_body.go) and
// call shapes / plain string API replacements (nore_calls.go).
type notOptimalRegularExpressions struct{}

func (notOptimalRegularExpressions) ID() string { return "NotOptimalRegularExpressions" }
func (notOptimalRegularExpressions) Semantic()  {}
func (notOptimalRegularExpressions) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var noreFunctions = []string{
	"preg_filter", "preg_grep", "preg_match_all", "preg_match",
	"preg_replace_callback", "preg_replace", "preg_split", "preg_quote",
}

// noreCase is one pattern literal being checked for one call.
type noreCase struct {
	ctx  *analysis.Context
	call *syntax.FuncCall
	fn   string
	lit  syntax.Expr // the resolved literal L
	body string
	mods string
	// decoded is the body of the PHP string value (escapes resolved),
	// for the checks that count pattern characters (D9–D11).
	decoded string
	array   bool // array mode (D5)
	// proven: the pattern argument certainly evaluates to lit, so a D22
	// rewrite may drop it.
	proven bool

	sink  *[]string       // dry run: collect literal findings instead of reporting
	prior map[string]bool // literal findings already made by an earlier call
}

func (c *noreCase) report(sev diagnostic.Severity, msg string) {
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
		c.proven = noreProvenPattern(norePatternArg(call, fn).Value, lit)
		if c.checkLiteral() && !array {
			c.checkCall()
		}
	}
}

// noreCandidates applies D1/D2: the call, its function name, the candidate
// literals and whether the call is in array mode.
func noreCandidates(ctx *analysis.Context, n syntax.Node) (*syntax.FuncCall, string, []syntax.Expr, bool) {
	call, fn := semanticquery.GlobalCall(ctx, n, noreFunctions...) // D1
	if call == nil || astquery.ArgCount(call) == 0 {
		return nil, "", nil, false
	}
	first := norePatternArg(call, fn)
	if first == nil {
		return nil, "", nil, false
	}
	var cands []syntax.Expr // D2
	arr, array := first.Value.(*syntax.Array)
	if array {
		for _, it := range arr.Items {
			if it == nil || it.Value == nil {
				continue
			}
			if lit := semanticquery.LocalStringLiteral(ctx.Types(), it.Value); lit != nil {
				cands = append(cands, lit)
			}
		}
	} else if lit := semanticquery.LocalStringLiteral(ctx.Types(), first.Value); lit != nil {
		cands = append(cands, lit)
	}
	return call, fn, cands, array
}

// norePatternArg returns the argument holding the pattern: the first one
// when positional, else the one named after the parameter (`pattern`, or
// `str` for preg_quote). nil when it cannot be told (spread, placeholder).
func norePatternArg(call *syntax.FuncCall, fn string) *syntax.Arg {
	name := "pattern"
	if fn == "preg_quote" {
		name = "str"
	}
	for i, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Value == nil || arg.Unpack {
			return nil
		}
		if i == 0 && arg.Name == nil || arg.Name != nil && arg.Name.Value == name {
			return arg
		}
	}
	return nil
}

// noreProvenPattern reports whether the pattern argument a can only be lit:
// the literal itself, or a class constant named without late static binding.
// Value discovery through a variable, a property or `static::` picks one
// possible value (a parameter default, an initial assignment), which is
// enough to check that literal but not to replace the call (D22).
func noreProvenPattern(a, lit syntax.Expr) bool {
	a = syntax.UnwrapParens(a)
	if a == lit {
		return true
	}
	cc, ok := a.(*syntax.ClassConstFetch)
	if !ok {
		return false
	}
	class, ok := cc.Class.(*syntax.Name)
	return ok && !strings.EqualFold(class.Value, "static")
}

// checkLiteral runs D3–D19 on the literal; it reports whether the pattern
// parsed (so the call checks may run).
func (c *noreCase) checkLiteral() bool {
	raw, _, ok := astquery.QuotedStringRaw(c.lit) // D3
	if !ok || raw == "" {
		return false
	}
	body, mods, ok := noreSplitDelimiters(raw) // D4
	// The string value decides what PHP sees: "\x2fa\x2f" and "/a/i\n" are
	// delimited although their source text is not.
	v, _ := astquery.QuotedStringValue(c.lit)
	vbody, _, vok := noreSplitDelimiters(v)
	if !ok {
		if c.fn != "preg_quote" && !vok { // D6
			c.report(diagnostic.SeverityWarning, "Pattern has no valid delimiters.")
		}
		return false
	}
	if c.fn == "preg_quote" {
		return true // custos: its argument is literal text, not a pattern (E3)
	}
	c.body, c.mods = body, mods
	c.decoded = body
	if vok {
		// D9–D11, D14 and D15 look at the pattern PCRE sees: "\$" and
		// "\x24" are a `$`.
		c.decoded = vbody
	}
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
// trailing modifier letters (PHP ignores spaces and line breaks among them).
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
	// t = start of the maximal run of trailing modifier characters
	t := len(r)
	for t > 0 && (isASCIILetter(r[t-1]) || noreModifierSpace(rune(r[t-1]))) {
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
			return r[openLen:j], strings.Map(noreDropModifierSpace, r[j+len(close):]), true
		}
	}
	return "", "", false
}
func isASCIILetter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }

// noreModifierSpace: the characters PHP skips in the modifier list.
func noreModifierSpace(r rune) bool { return r == ' ' || r == '\n' || r == '\r' }

func noreDropModifierSpace(r rune) rune {
	if noreModifierSpace(r) {
		return -1
	}
	return r
}

// noreEscaped reports whether s[i] follows an odd number of backslashes.
func noreEscaped(s string, i int) bool {
	n := 0
	for i > n && s[i-n-1] == '\\' {
		n++
	}
	return n%2 == 1
}

// noreHasUnescaped reports whether s contains key at a position that is
// not escaped by a backslash.
func noreHasUnescaped(s, key string) bool {
	for from := 0; ; {
		i := strings.Index(s[from:], key)
		if i < 0 {
			return false
		}
		if !noreEscaped(s, from+i) {
			return true
		}
		from += i + 1
	}
}

// noreHasLineTerminator reports whether s contains \n, \r, U+0085, U+2028
// or U+2029.
func noreHasLineTerminator(s string) bool {
	return strings.ContainsAny(s, "\n\r\u0085  ")
}

// noreCountDiff returns count(a) - count(b) (non-overlapping occurrences).
func noreCountDiff(s, a, b string) int { return strings.Count(s, a) - strings.Count(s, b) }
