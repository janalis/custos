package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// incrementDecrementOperationEquivalent suggests ++/-- for `$n += 1`,
// `$n = $n - 1` and friends.
type incrementDecrementOperationEquivalent struct{}

func init() { register(incrementDecrementOperationEquivalent{}) }

func (incrementDecrementOperationEquivalent) ID() string {
	return "IncrementDecrementOperationEquivalent"
}

func (incrementDecrementOperationEquivalent) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign}
}

func incDecIsOne(e syntax.Expr) bool {
	l, ok := e.(*syntax.Literal)
	return ok && l.LitKind == syntax.LitInt && l.Raw == "1"
}

func (incrementDecrementOperationEquivalent) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Assign)
	op := ""
	switch a.Op.Kind {
	case syntax.TPlusEqual: // D1
		if incDecIsOne(a.Value) {
			op = "++"
		}
	case syntax.TMinusEqual: // D2
		if incDecIsOne(a.Value) {
			op = "--"
		}
	case syntax.TEqual:
		if a.ByRef {
			return
		}
		b, ok := a.Value.(*syntax.Binary)
		if !ok {
			return
		}
		switch b.Op.Kind {
		case syntax.TPlus: // D3
			if (incDecIsOne(b.Left) && util.EquivalentFoldNames(ctx.File, b.Right, a.Var)) ||
				(incDecIsOne(b.Right) && util.EquivalentFoldNames(ctx.File, b.Left, a.Var)) {
				op = "++"
			}
		case syntax.TMinus: // D4
			if incDecIsOne(b.Right) && util.EquivalentFoldNames(ctx.File, b.Left, a.Var) {
				op = "--"
			}
		}
	}
	if op == "" {
		return
	}
	// `+ 1` on a string or bool is arithmetic (or a TypeError), `++` on a
	// string is alphanumeric and on a bool a no-op (custos diverges).
	operand := a.Var
	if b, ok := a.Value.(*syntax.Binary); ok && a.Op.Kind == syntax.TEqual {
		operand = b.Left
		if incDecIsOne(b.Left) && b.Op.Kind == syntax.TPlus {
			operand = b.Right
		}
	}
	if t := ctx.TypeOf(operand); t.Has("string") || t.Has("bool") || t.Has("true") || t.Has("false") {
		return
	}
	if dim, ok := a.Var.(*syntax.ArrayDimFetch); ok { // D5
		t := ctx.TypeOf(dim.Var)
		if t.IsUnknown() || t.Has("string") {
			return
		}
		arr := false
		for _, at := range t.Atoms() {
			if at == "array" || strings.HasSuffix(at, "[]") {
				arr = true
			}
		}
		if !arr {
			return
		}
	}
	target := ctx.Text(a.Var)
	repl := op + target
	if !ctx.Bool("PREFER_PREFIX_STYLE") {
		repl = target + op
	}
	span := a.Span()
	ctx.Report(span, "Use '"+repl+"' instead.", analysis.Fix{
		Title: "Use the " + op + " operator",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
