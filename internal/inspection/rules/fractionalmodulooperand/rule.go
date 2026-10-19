// Package fractionalmodulooperand implements the native FractionalModuloOperand inspection.
package fractionalmodulooperand

import (
	"math"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use fmod for fractional operands."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FractionalModuloOperand" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TMod {
		return
	}
	for _, e := range []syntax.Expr{b.Left, b.Right} {
		v := semanticquery.NativeValue(ctx, e)
		if u, ok := v.(*syntax.Unary); ok && (u.Op.Kind == syntax.TMinus || u.Op.Kind == syntax.TPlus) {
			v = syntax.UnwrapParens(u.Expr)
		}
		if l, ok := v.(*syntax.Literal); ok && l.LitKind == syntax.LitFloat {
			f, err := strconv.ParseFloat(strings.ReplaceAll(l.Raw, "_", ""), 64)
			if err == nil && !math.IsInf(f, 0) && !math.IsNaN(f) && math.Trunc(f) != f && !rewritable(ctx, b, e) {
				ctx.ReportNode(b, message)
				return
			}
		}
	}
}

// rewritable reports whether a variable operand can change behind the local
// value facts: through a closure holding a reference, a builtin that writes
// the caller's symbol table, or, for a global, application code. The local
// facts follow `goto` jumps, and give up on a scope that also has a `try`.
// Local facts may chain through other variables, so names are not compared.
func rewritable(ctx *analysis.Context, at *syntax.Binary, operand syntax.Expr) bool {
	if _, ok := syntax.UnwrapParens(operand).(*syntax.Variable); !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(at)
	// Code after the operation only matters when a loop runs it again.
	limit := at.Span().End
	for p := at.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = math.MaxUint32
		}
	}
	found := false
	visit := func(n syntax.Node) bool {
		if found {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure:
			for _, use := range n.Uses {
				found = found || (use.ByRef && n.Span().Start < limit)
			}
		case *syntax.FuncCall:
			if n.Span().Start < limit {
				name := semanticquery.NativeBuiltinName(ctx, n)
				found = name == "extract" || (name == "parse_str" && semanticquery.CallArgument(n.Args, 1, "result") == nil) ||
					(scope == nil && name == "")
			}
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			found = scope == nil && n.Span().Start < limit
		}
		return !syntax.IsVariableScope(n)
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	}
	return found
}
