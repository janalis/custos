package isiterablecanbeused

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// isIterableCanBeUsed reports `is_array($v) || $v instanceof Traversable`
// pairs that is_iterable() covers.
type isIterableCanBeUsed struct{}

func (isIterableCanBeUsed) ID() string { return "IsIterableCanBeUsed" }
func (isIterableCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (isIterableCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	if ctx.GlobalFunctionName(call) != "is_array" { // D1 (any case, global function)
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 1 { // D2
		return
	}
	argVal := args[0]
	if unstableOperand(ctx, argVal) { // E5: one evaluation is not two
		return
	}
	parent, ok := call.Parent().(*syntax.Binary)
	if !ok || parent.Op.Kind != syntax.TBooleanOr { // D3
		return
	}
	top := parent // D4
	for {
		p, _ := astquery.ParentSkipParens(top)
		b, ok := p.(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TBooleanOr {
			break
		}
		top = b
	}
	var subject syntax.Expr
	var walk func(b *syntax.Binary) bool // D5, D6
	walk = func(b *syntax.Binary) bool {
		for _, op := range [2]syntax.Expr{b.Left, b.Right} {
			e := syntax.UnwrapParens(op)
			if inner, ok := e.(*syntax.Binary); ok && inner.Op.Kind == syntax.TBooleanOr {
				if walk(inner) {
					return true
				}
				continue
			}
			io, ok := e.(*syntax.Instanceof)
			if !ok || !semanticquery.NamesGlobalClass(ctx, io.Class, "Traversable") || !astquery.EquivalentFoldNames(ctx.File, io.Expr, argVal) {
				continue
			}
			subject = io.Expr
			return true
		}
		return false
	}
	if walk(top) { // D7
		ctx.ReportNode(call, "Use 'is_iterable("+ctx.Text(subject)+")' instead of the is_array()/instanceof Traversable pair.", pairFix(ctx, call, parent, argVal)...)
	}
}

// pairFix returns the rewrite of the pair into one is_iterable() call (F1),
// or nothing when the two halves are not neighbours in the `||` chain, the
// subject is not a plain variable or a comment sits inside the pair.
func pairFix(ctx *analysis.Context, call *syntax.FuncCall, parent *syntax.Binary, arg syntax.Expr) []diagnostic.Fix {
	v, ok := astquery.PlainVariableName(arg)
	if !ok {
		return nil
	}
	other := func(e syntax.Expr) bool {
		io, ok := syntax.UnwrapParens(e).(*syntax.Instanceof)
		if !ok || !semanticquery.NamesGlobalClass(ctx, io.Class, "Traversable") {
			return false
		}
		name, ok := astquery.PlainVariableName(io.Expr)
		return ok && name == v
	}
	span := call.Span()
	gp, _ := parent.Parent().(*syntax.Binary)
	left, _ := parent.Left.(*syntax.Binary)
	switch {
	case parent.Left == call: // is_array($v) || $v instanceof T
		if !other(parent.Right) {
			return nil
		}
		span.End = parent.Right.Span().End
	case other(parent.Left): // $v instanceof T || is_array($v)
		span.Start = parent.Left.Span().Start
	case gp != nil && gp.Op.Kind == syntax.TBooleanOr && gp.Left == parent && other(gp.Right): // $x || is_array($v) || $v instanceof T
		span.End = gp.Right.Span().End
	case left != nil && left.Op.Kind == syntax.TBooleanOr && other(left.Right): // $x || $v instanceof T || is_array($v)
		span.Start = left.Right.Span().Start
	default:
		return nil
	}
	if astquery.HasComment(ctx.File, span) {
		return nil
	}
	// Qualified unless the bare name provably is the built-in here: a
	// namespaced or imported is_iterable() would win over it.
	repl := `\is_iterable($` + v + ")"
	name := call.Name.(*syntax.Name)
	if fqn, fallback := ctx.Names().Function("is_iterable", name.Span().Start); !strings.HasPrefix(name.Value, `\`) && fqn == "is_iterable" && fallback == "" {
		repl = repl[1:]
	}
	return []diagnostic.Fix{astquery.ReplaceFix(span, repl)}
}

// unstableOperand reports whether evaluating e twice provably differs from
// evaluating it once (E5): it contains an assignment, an increment or
// decrement, an argument bound to a by-reference parameter of a resolved
// function, or a call to one of the built-ins listed above.
func unstableOperand(ctx *analysis.Context, e syntax.Expr) bool {
	found := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.Assign, *syntax.IncDec:
			found = true
		case *syntax.FuncCall:
			f := ctx.Types().ResolveFunction(c)
			found = found || astquery.ImpureBuiltin(ctx.GlobalFunctionName(c)) ||
				f != nil && semanticquery.ArgBindsByRef(c.Args, f.Params, false)
		}
		return !found
	})
	return found
}
