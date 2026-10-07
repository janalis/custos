package performance

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// strStrUsedAsStrPos reports strstr()/stristr() whose result only serves as
// a boolean and suggests strpos()/stripos() compared with false.
type strStrUsedAsStrPos struct{}

func init() { register(strStrUsedAsStrPos{}) }

func (strStrUsedAsStrPos) ID() string { return "StrStrUsedAsStrPos" }

func (strStrUsedAsStrPos) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (strStrUsedAsStrPos) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := perfCall(ctx, n, "strstr", "stristr") // D1
	if call == nil {
		return
	}
	args := perfArgs(call.Args)
	if len(args) < 2 || call.Args == nil || len(args) != len(call.Args.Args) { // D2
		return
	}
	fn := "strpos"
	if name == "stristr" {
		fn = "stripos"
	}
	qual, _, _ := util.CallName(call)
	if qual == "" { // a namespaced or imported strpos/stripos would capture a bare call
		fn = util.QualifiedBuiltin(ctx, fn, call.Span().Start)
	}

	var target syntax.Node
	var op string
	if b, ok := call.Parent().(*syntax.Binary); ok && isEqualityOp(b.Op.Kind) { // D3
		other := b.Left
		if other == syntax.Expr(call) {
			other = b.Right
		}
		if !perfIsFalse(other) { // D4
			return // pattern B cannot match a comparison parent either
		}
		op = ctx.SpanText(b.Op.Span)
		switch {
		case op == "<>":
			op = "!==" // see spec Divergences
		case len(op) == 2:
			op += "="
		}
		target = b
	} else {
		if !util.IsLogicalOperand(n) { // D5
			return
		}
		target, op = call, "!=="
		if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
			target, op = u, "==="
		}
	}

	repl := qual + fn + "(" + ctx.Text(args[0].Value) + ", " + ctx.Text(args[1].Value) + ")"
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = "false " + op + " " + repl
	} else {
		repl = repl + " " + op + " false"
	}
	span := target.Span()
	ctx.Report(span, "Use '"+repl+"' instead; it avoids building a substring.", analysis.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

func isEqualityOp(k syntax.TokenKind) bool {
	switch k {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
		return true
	}
	return false
}
