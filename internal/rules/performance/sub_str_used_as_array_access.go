package performance

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// subStrUsedAsArrayAccess reports substr($s, $i, 1) on a string and suggests
// string offset access.
type subStrUsedAsArrayAccess struct{}

func init() { register(subStrUsedAsArrayAccess{}) }

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (subStrUsedAsArrayAccess) Semantic() {}

func (subStrUsedAsArrayAccess) ID() string { return "SubStrUsedAsArrayAccess" }

func (subStrUsedAsArrayAccess) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (subStrUsedAsArrayAccess) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := perfCall(ctx, n, "substr") // D1
	if call == nil {
		return
	}
	args, ok := perfPlainArgs(call.Args)
	if !ok || len(args) != 3 { // D2
		return
	}
	if lit, ok := args[2].(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt || lit.Raw != "1" { // D3
		return
	}
	switch args[0].(type) { // D4
	case *syntax.Variable, *syntax.ArrayDimFetch, *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
	default:
		return
	}
	if !ctx.TypeOf(args[0]).Has("string") { // D5
		return
	}
	src, off := ctx.Text(args[0]), ctx.Text(args[1])
	var access string
	if strings.HasPrefix(off, "-") { // R1: only `-1` keeps substr's result
		if !isMinusOne(args[1]) {
			return
		}
		access = src + "[" + util.QualifiedBuiltin(ctx, "strlen", call.Span().Start) + "(" + src + ") - 1]"
	} else { // R2: the offset must be known to be an int
		if !ctx.TypeOf(args[1]).OnlyOf("int") {
			return
		}
		access = src + "[" + off + "]"
	}
	span := call.Span()
	// R3: `??` does not parse below 7.0, no fix; custos: nor when the
	// source may be another scalar (an int's offset is null, not a digit).
	if ctx.PHP < phpver.PHP70 || !ctx.TypeOf(args[0]).OnlyOf("string", "null") {
		ctx.Report(span, "Use '"+access+"' (string offset access) instead.")
		return
	}
	repl := "(" + access + " ?? '')"
	ctx.Report(span, "Use '"+repl+"' (string offset access) instead.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

// isMinusOne reports whether e is the literal `-1` (unary minus on `1`).
func isMinusOne(e syntax.Expr) bool {
	u, ok := e.(*syntax.Unary)
	if !ok || u.Op.Kind != syntax.TMinus {
		return false
	}
	lit, ok := u.Expr.(*syntax.Literal)
	return ok && lit.LitKind == syntax.LitInt && lit.Raw == "1"
}
