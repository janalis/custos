package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// isNullFunctionUsage rewrites is_null() calls into identity comparisons
// with null.
type isNullFunctionUsage struct{}

func init() { register(isNullFunctionUsage{}) }

func (isNullFunctionUsage) ID() string { return "IsNullFunctionUsage" }

func (isNullFunctionUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (isNullFunctionUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, ok := call.Name.(*syntax.Name)
	if !ok || call.Args == nil || len(call.Args.Args) != 1 { // E1
		return
	}
	if !ctx.IsGlobalFunctionCall(call, "is_null") { // D1, E2
		return
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || arg.Value == nil || arg.Unpack {
		return
	}
	// D2
	var target syntax.Node = call
	positive := true
	switch p := call.Parent().(type) {
	case *syntax.Unary:
		if p.Op.Kind == syntax.TExclaim {
			target, positive = p, false
		}
	case *syntax.Binary:
		other := p.Right
		if syntax.Node(p.Right) == syntax.Node(call) {
			other = p.Left
		}
		val, isBool := inBoolConst(other)
		if !isBool {
			break
		}
		switch p.Op.Kind {
		case syntax.TIsEqual, syntax.TIsIdentical:
			target, positive = p, val
		case syntax.TIsNotEqual, syntax.TIsNotIdentical:
			target, positive = p, !val
		}
	}
	a := ctx.Text(arg.Value)
	switch arg.Value.(type) {
	case *syntax.Binary, *syntax.Instanceof:
		a = "(" + a + ")"
	default:
		if util.NeedsParensAsEqualityOperand(arg.Value) { // assignments, ternaries, include, print…
			a = "(" + a + ")"
		}
	}
	op := "==="
	if !positive {
		op = "!=="
	}
	repl := a + " " + op + " null"
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = "null " + op + " " + a
	}
	span := target.Span()
	msg := "Replace with '" + repl + "'."
	if isEmptyNeedsParens(target) { // `'v=' . $x === null` would compare the concatenation
		repl = "(" + repl + ")"
	}
	ctx.Report(span, msg, analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// inBoolConst reports whether e is the constant true/false (any case,
// optionally fully qualified) and its value. Kept apart from util.BoolConst,
// which folds case with Unicode rules (`falſe` matches there, not here).
func inBoolConst(e syntax.Expr) (val, ok bool) {
	c, isConst := e.(*syntax.ConstFetch)
	if !isConst || c.Name == nil {
		return false, false
	}
	switch strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)) {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}
