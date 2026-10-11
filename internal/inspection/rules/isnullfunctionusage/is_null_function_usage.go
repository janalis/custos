package isnullfunctionusage

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// isNullFunctionUsage rewrites is_null() calls into identity comparisons
// with null.
type isNullFunctionUsage struct{}

func (isNullFunctionUsage) ID() string               { return "IsNullFunctionUsage" }
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
		if astquery.NeedsParensAsEqualityOperand(arg.Value) { // assignments, ternaries, include, print…
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
	if astquery.IsEmptyNeedsParens(target) { // `'v=' . $x === null` would compare the concatenation
		repl = "(" + repl + ")"
	}
	title := "Replace with '" + repl + "'"
	// `return!is_null(foo())` must not become `returnfoo() !== null`, nor
	// `is_null($x)and $y` become `$x === nulland $y`.
	if span.Start > 0 && isWordByte(ctx.Src[span.Start-1]) && (isWordByte(repl[0]) || repl[0] == '\\') {
		repl = " " + repl
	}
	if int(span.End) < len(ctx.Src) && isWordByte(ctx.Src[span.End]) && isWordByte(repl[len(repl)-1]) {
		repl += " "
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: title,
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// isWordByte reports whether b can be part of a PHP identifier or number, so
// that two such bytes written side by side lex as a single token.
func isWordByte(b byte) bool {
	return b == '_' || b >= 0x80 || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// inBoolConst reports whether e is the constant true/false (any case,
// optionally fully qualified) and its value. Kept apart from astquery.BoolConst,
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
