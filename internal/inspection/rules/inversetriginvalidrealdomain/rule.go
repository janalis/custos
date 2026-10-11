// Package inversetriginvalidrealdomain implements the native InverseTrigInvalidRealDomain inspection.
package inversetriginvalidrealdomain

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep the inverse trigonometric argument within minus one and one."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "InverseTrigInvalidRealDomain" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "asin" && name != "acos" {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "num")
	value, known := number(ctx, arg)
	if known && (value < -1 || value > 1) && !steppedBeforeCall(ctx.File, c, arg) {
		ctx.ReportNode(c, message)
	}
}

func number(ctx *analysis.Context, e syntax.Expr) (float64, bool) {
	if n, ok := semanticquery.NativeInt(ctx, e); ok {
		return float64(n), true
	}
	value := semanticquery.NativeValue(ctx, e)
	sign := 1.0
	if u, ok := value.(*syntax.Unary); ok {
		if u.Op.Kind != syntax.TMinus && u.Op.Kind != syntax.TPlus {
			return 0, false
		}
		if u.Op.Kind == syntax.TMinus {
			sign = -1
		}
		value = syntax.UnwrapParens(u.Expr)
	}
	lit, ok := value.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitFloat {
		return 0, false
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(lit.Raw, "_", ""), 64)
	return sign * n, err == nil
}

// steppedBeforeCall reports whether a variable argument may have been changed
// by `++`/`--`, through a by-reference closure capture or by extract() before
// the call runs. The value resolution treats none as a write, so a counter that
// starts outside the domain would otherwise still look like a proven value.
func steppedBeforeCall(f *syntax.File, call *syntax.FuncCall, arg syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(arg).(*syntax.Variable)
	if !ok {
		return false
	}
	// A write placed after the call still precedes it on the next iteration
	// of any loop enclosing the call.
	looped := false
	var scope syntax.Node
	for p := call.Parent(); p != nil && scope == nil; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			looped = true
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			scope = p
		}
	}
	stepped := false
	visit := func(x syntax.Node) bool {
		writes := false
		switch w := x.(type) {
		case *syntax.IncDec:
			target, plain := syntax.UnwrapParens(w.Var).(*syntax.Variable)
			writes = plain && target.Name == v.Name
		case *syntax.Closure:
			for _, use := range w.Uses {
				writes = writes || use.ByRef && use.Var.Name == v.Name
			}
		case *syntax.FuncCall:
			_, writes = astquery.IsFuncNamedFold(w, "extract")
		}
		if writes && (looped || x.Span().Start < call.Span().Start) {
			stepped = true
		}
		// A nested function owns its variables; only a by-reference capture
		// shares one with the scope of the call.
		return !stepped && (x == scope || !syntax.IsVariableScope(x))
	}
	if scope == nil {
		syntax.InspectFile(f, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return stepped
}
