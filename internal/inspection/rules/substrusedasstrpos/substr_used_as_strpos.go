package substrusedasstrpos

import (
	"strconv"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// subStrUsedAsStrPos reports prefix checks written as
// `substr($s, 0, strlen($p)) == $p` and suggests `strpos($s, $p) === 0`.
type subStrUsedAsStrPos struct{}

func (subStrUsedAsStrPos) ID() string { return "SubStrUsedAsStrPos" }
func (subStrUsedAsStrPos) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func isCaseFolding(name string) bool {
	switch name {
	case "strtolower", "strtoupper", "mb_strtolower", "mb_strtoupper", "mb_convert_case":
		return true
	}
	return false
}

func (subStrUsedAsStrPos) Check(ctx *analysis.Context, n syntax.Node) {
	s := n.(*syntax.FuncCall)
	name := ctx.GlobalFunctionName(s)
	if name != "substr" && name != "mb_substr" { // D1
		return
	}
	args, ok := astquery.CallArgValues(s)
	if !ok || len(args) < 3 || len(args) > 4 { // D2
		return
	}
	if lit, ok := args[1].(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt || lit.Raw != "0" { // D3
		return
	}
	// D5
	var subject syntax.Expr = s
	insensitive, folding := false, ""
	if _, isArg := s.Parent().(*syntax.Arg); isArg {
		wc := astquery.ParentFuncCall(s)
		if wc == nil || !isCaseFolding(ctx.GlobalFunctionName(wc)) {
			return
		}
		wargs, ok := astquery.CallArgValues(wc)
		if !ok || len(wargs) != 1 {
			return
		}
		subject, insensitive, folding = wc, true, ctx.GlobalFunctionName(wc)
	}
	// D6
	b, ok := subject.Parent().(*syntax.Binary)
	if !ok {
		return
	}
	other := b.Right
	if b.Right == subject {
		other = b.Left
	}
	if !prefixLengthMatches(ctx, name, args[2], other) { // D7
		return
	}
	if insensitive && !foldedLiteral(folding, other) { // D8
		return
	}
	var op string
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsIdentical:
		op = "==="
	case syntax.TIsNotEqual, syntax.TIsNotIdentical:
		op = "!==" // `<>` too (spec Divergences)
	default:
		return
	}
	base := "strpos"
	if insensitive {
		base = "stripos"
	}
	if name == "mb_substr" {
		base = "mb_" + base
	}
	fn := semanticquery.QualifiedBuiltinFor(ctx, base, s) // a namespaced function of that name would capture a bare call
	call := fn + "(" + ctx.Text(args[0]) + ", " + ctx.Text(other)
	if name == "mb_substr" && len(args) == 4 {
		// The encoding is mb_strpos' 4th parameter, after the offset.
		call += ", 0, " + ctx.Text(args[3])
	}
	call += ")"
	repl := call + " " + op + " 0"
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = "0 " + op + " " + call
	}
	span := b.Span()
	msg := "Use '" + repl + "' instead."
	if loose := b.Op.Kind == syntax.TIsEqual || b.Op.Kind == syntax.TIsNotEqual; loose && !nonNumericLiteral(other) {
		// `==` compares numeric strings as numbers ("0 " == "00"), strpos()
		// matches bytes: no fix (custos diverges).
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Use '" + fn + "'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// nonNumericLiteral reports whether e is a string literal that is not a
// numeric string: comparing it loosely with a string is a byte comparison.
func nonNumericLiteral(e syntax.Expr) bool {
	v, ok := astquery.QuotedStringValue(e)
	return ok && !isNumericString(v)
}

// prefixLengthMatches reports whether the substr length is the length of the
// compared value, counted in the unit of the substr variant (D7): a
// strlen/mb_strlen call on an operand equivalent to other, or an integer
// literal equal to the length of the string literal other.
func prefixLengthMatches(ctx *analysis.Context, substr string, length, other syntax.Expr) bool {
	mb := substr == "mb_substr"
	if l, ok := length.(*syntax.FuncCall); ok {
		ln := ctx.GlobalFunctionName(l)
		if (mb && ln != "mb_strlen") || (!mb && ln != "strlen") {
			return false
		}
		largs, ok := astquery.CallArgValues(l)
		if !ok || len(largs) < 1 || len(largs) > 2 || (!mb && len(largs) != 1) {
			return false
		}
		return astquery.EquivalentFoldNames(ctx.File, largs[0], other)
	}
	lit, ok := length.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitInt {
		return false
	}
	n, err := strconv.Atoi(lit.Raw)
	if err != nil {
		return false
	}
	val, ok := astquery.QuotedStringValue(other)
	if !ok || (mb && !isASCII(val)) {
		return false
	}
	return len(val) == n
}

// foldedLiteral reports whether other is an ASCII string literal already in
// the case produced by the folding function (D8), so that a case-insensitive
// search is equivalent to comparing the folded prefix.
func foldedLiteral(folding string, other syntax.Expr) bool {
	val, ok := astquery.QuotedStringValue(other)
	if !ok || !isASCII(val) {
		return false
	}
	var lo, hi byte
	switch folding {
	case "strtoupper", "mb_strtoupper":
		lo, hi = 'a', 'z'
	case "strtolower", "mb_strtolower":
		lo, hi = 'A', 'Z'
	default:
		return false
	}
	for i := 0; i < len(val); i++ {
		if val[i] >= lo && val[i] <= hi {
			return false
		}
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
