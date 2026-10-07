package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// strContainsCanBeUsed reports `strpos($h, $n) !== false` checks that
// str_contains() expresses directly.
type strContainsCanBeUsed struct{}

func init() { register(strContainsCanBeUsed{}) }

// Semantic marks the rule as needing the project index (user functions with
// the matched names may be declared in other files).
func (strContainsCanBeUsed) Semantic() {}

func (strContainsCanBeUsed) ID() string { return "StrContainsCanBeUsed" }

func (strContainsCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strContainsCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP80 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	m, ok := matchStrCallCompare(ctx, call, 2, "strpos", "mb_strpos") // D1-D3
	if !ok {
		return
	}
	if c, ok := m.other.(*syntax.ConstFetch); !ok || c.Name == nil || !strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") { // D4
		return
	}
	reportStrCallReplacement(ctx, m, "str_contains", m.args[1], m.cmp.Op.Kind == syntax.TIsIdentical) // D5
}

// strCallCompare is a named function call compared with ===/!== as a direct
// operand.
type strCallCompare struct {
	call  *syntax.FuncCall
	qual  string        // namespace qualifier as written
	args  []syntax.Expr // argument values
	cmp   *syntax.Binary
	other syntax.Expr // the other comparison operand
}

// matchStrCallCompare matches a call to one of names (last segment,
// case-insensitive, names lower-case) that resolves to that global function, with exactly nargs
// plain arguments, whose direct parent is a strict (non-)identity comparison.
func matchStrCallCompare(ctx *analysis.Context, call *syntax.FuncCall, nargs int, names ...string) (strCallCompare, bool) {
	var m strCallCompare
	qual, name, ok := util.CallName(call)
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
	if !found || !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) {
		return m, false
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) != nargs {
		return m, false
	}
	b, ok := call.Parent().(*syntax.Binary)
	if !ok || (b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical) {
		return m, false
	}
	other := b.Left
	if other == syntax.Expr(call) {
		other = b.Right
	} else if b.Right != syntax.Expr(call) {
		return m, false
	}
	return strCallCompare{call: call, qual: qual, args: args, cmp: b, other: other}, true
}

// reportStrCallReplacement reports the comparison and offers to replace it
// with `[!]qual fn(haystack, needle)`, where qual is `\` when the original
// call was fully qualified or a bare fn would not reach the global function.
func reportStrCallReplacement(ctx *analysis.Context, m strCallCompare, fn string, needle syntax.Expr, negate bool) {
	qual := ""
	if m.qual != "" || !util.BareReachesGlobal(ctx, fn, m.call.Span().Start) {
		qual = `\`
	}
	r := qual + fn + "(" + ctx.Text(m.args[0]) + ", " + ctx.Text(needle) + ")"
	if negate {
		r = "!" + r
	}
	span := m.cmp.Span()
	ctx.Report(span, "Replace with '"+r+"'.", analysis.Fix{
		Title: "Use " + fn + "()",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: r}}
		},
	})
}
