package strstrusedasstrpos

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// strStrUsedAsStrPos reports strstr()/stristr() whose result only serves as
// a boolean and suggests strpos()/stripos() compared with false.
type strStrUsedAsStrPos struct{}

func (strStrUsedAsStrPos) ID() string               { return "StrStrUsedAsStrPos" }
func (strStrUsedAsStrPos) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (strStrUsedAsStrPos) Check(ctx *analysis.Context, n syntax.Node) {
	call, name := semanticquery.GlobalCall(ctx, n, "strstr", "stristr") // D1
	if call == nil {
		return
	}
	// Named or unpacked arguments have no fixed position: no report.
	args, ok := astquery.ArgValues(call.Args)
	if !ok || len(args) < 2 || len(args) > 3 { // D2
		return
	}
	// D2a: a before_needle that is not a literal cannot be dropped; a literal
	// true makes the result a prefix, whose truthiness is not "needle found".
	prefix := false
	if len(args) == 3 {
		if prefix, ok = astquery.BoolConst(args[2]); !ok {
			return
		}
	}
	fn := "strpos"
	if name == "stristr" {
		fn = "stripos"
	}
	qual, _, _ := astquery.CallName(call)
	if qual == "" { // a namespaced or imported strpos/stripos would capture a bare call
		fn = semanticquery.QualifiedBuiltin(ctx, fn, call.Span().Start)
	}

	var target syntax.Node
	var op string
	parens := false
	if b, ok := call.Parent().(*syntax.Binary); ok && isEqualityOp(b.Op.Kind) { // D3
		other := b.Left
		if other == syntax.Expr(call) {
			other = b.Right
		}
		if !perfIsFalse(other) { // D4
			return // pattern B cannot match a comparison parent either
		}
		op = ctx.SpanText(b.Op.Span)
		if len(op) == 2 {
			if prefix { // '' == false although the needle was found
				return
			}
			if op == "<>" {
				op = "!==" // see spec Divergences
			} else {
				op += "="
			}
		}
		target = b
	} else {
		if prefix || !astquery.IsLogicalOperand(n) { // D5
			return
		}
		target, op = call, "!=="
		if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
			target, op = u, "==="
			// F2: '!' binds tighter than the comparison that replaces it.
			_, isClone := u.Parent().(*syntax.Clone)
			parens = isClone || astquery.NeedsParensAsComparison(u)
		}
	}

	repl := qual + fn + "(" + ctx.Text(args[0]) + ", " + ctx.Text(args[1]) + ")"
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = "false " + op + " " + repl
	} else {
		repl = repl + " " + op + " false"
	}
	span := target.Span()
	text := repl
	if parens {
		text = "(" + repl + ")"
	}
	ctx.Report(span, "Use '"+repl+"' instead; it avoids building a substring.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} },
	})
}

func isEqualityOp(k syntax.TokenKind) bool {
	switch k {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
		return true
	}
	return false
}
