package semanticquery

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// StrCallCompare is a named function call compared with ===/!== as a direct
// operand.
type StrCallCompare struct {
	Call  *syntax.FuncCall
	Qual  string        // namespace qualifier as written
	Args  []syntax.Expr // argument values
	Cmp   *syntax.Binary
	Other syntax.Expr // the other comparison operand
}

// MatchStrCallCompare matches a call to one of names (last segment,
// case-insensitive, names lower-case) that resolves to that global function, with exactly nargs
// plain arguments, whose direct parent is a strict (non-)identity comparison.
func MatchStrCallCompare(ctx *analysis.Context, call *syntax.FuncCall, nargs int, names ...string) (StrCallCompare, bool) {
	var m StrCallCompare
	qual, name, ok := astquery.CallName(call)
	if !ok {
		return m, false
	}
	found := false
	for _, want := range names {
		if strings.EqualFold(name, want) {
			found = true
			break
		}
	}
	if !found || !ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) {
		return m, false
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != nargs {
		return m, false
	}
	b, ok := call.Parent().(*syntax.Binary)
	if !ok || (b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical) {
		return m, false
	}
	other := b.Left // call is b.Left or b.Right, its direct parent
	if other == syntax.Expr(call) {
		other = b.Right
	}
	return StrCallCompare{Call: call, Qual: qual, Args: args, Cmp: b, Other: other}, true
}

// ReportStrCallReplacement reports the comparison and offers to replace it
// with `[!]qual fn(haystack, needle)`, where qual is `\` when the original
// call was fully qualified or a bare fn would not reach the global function.
func ReportStrCallReplacement(ctx *analysis.Context, m StrCallCompare, fn string, needle syntax.Expr, negate bool) {
	ReportStrCallReplacementFix(ctx, m, fn, needle, negate, true)
}

// ReportStrCallReplacementFix is reportStrCallReplacement with the fix
// optional.
func ReportStrCallReplacementFix(ctx *analysis.Context, m StrCallCompare, fn string, needle syntax.Expr, negate, fixable bool) {
	qual := ""
	if m.Qual != "" || !BareReachesGlobal(ctx, fn, m.Call.Span().Start) {
		qual = `\`
	}
	r := qual + fn + "(" + ctx.Text(m.Args[0]) + ", " + ctx.Text(needle) + ")"
	if negate {
		r = "!" + r
	}
	span := m.Cmp.Span()
	if !fixable {
		ctx.Report(span, "Replace with '"+r+"'.")
		return
	}
	ctx.Report(span, "Replace with '"+r+"'.", diagnostic.Fix{
		Title: "Use " + fn + "()",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: r}}
		},
	})
}
