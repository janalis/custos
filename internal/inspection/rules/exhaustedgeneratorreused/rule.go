// Package exhaustedgeneratorreused implements the native ExhaustedGeneratorReused inspection.
package exhaustedgeneratorreused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Create a fresh generator before traversing again."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExhaustedGeneratorReused" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach, syntax.KFuncCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	var input syntax.Expr
	switch x := n.(type) {
	case *syntax.Foreach:
		input = x.Expr
	case *syntax.FuncCall:
		if !semanticquery.NativeBuiltin(ctx, x, "iterator_to_array") {
			return
		}
		input = semanticquery.CallArgument(x.Args, 0, "iterator")
	}
	producer, ok := semanticquery.NativeValue(ctx, input).(*syntax.FuncCall)
	if !ok {
		return
	}
	if !semanticquery.NativeGeneratorBody(semanticquery.NativeFunctionBody(ctx, producer)) {
		return
	}
	// Only a plain local variable carries one generator identity between
	// two traversals: every other expression may yield a fresh generator.
	v, ok := input.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return
	}
	if _, known := ctx.Flow().StateBefore(input, input, "iterator_to_array"); known {
		if !reboundElsewhere(ctx.File, v) {
			ctx.ReportNode(n, message)
		}
		return
	}
	f, ok := n.(*syntax.Foreach)
	if !ok {
		return
	}
	previous, ok := astquery.PrevStmt(ctx.File, f)
	if !ok {
		return
	}
	first, ok := previous.(*syntax.Foreach)
	if !ok || first.ByRef || !sameVariable(first.Expr, v) || sameVariable(first.Key, v) || sameVariable(first.Value, v) {
		return
	}
	body, ok := first.Body.(*syntax.Block)
	if !ok || len(body.Stmts) != 0 {
		return
	}
	gbody := semanticquery.NativeFunctionBody(ctx, producer)
	for _, st := range gbody.Stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			return
		}
		y, ok := es.Expr.(*syntax.Yield)
		if !ok || flowquery.MayHaveSideEffects(y.Value) {
			return
		}
	}
	ctx.ReportNode(n, message)
}

// sameVariable reports whether e is the plain variable v names.
func sameVariable(e syntax.Expr, v *syntax.Variable) bool {
	o, ok := e.(*syntax.Variable)
	return ok && o.NameExpr == nil && o.Name == v.Name
}

// reboundElsewhere reports whether the scope of v lets code the flow facts do
// not follow bind v to another generator: a closure capturing it by
// reference, a `global` or `static` declaration, a symbol-table import
// (extract, parse_str) or an included file. Top-level variables are also
// reachable through a `global` declaration of any function of the file.
func reboundElsewhere(f *syntax.File, v *syntax.Variable) bool {
	scope := syntax.EnclosingFuncLike(v)
	local := func(x syntax.Node) bool { return syntax.EnclosingFuncLike(x) == scope }
	found := false
	visit := func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.ClosureUse:
			found = found || x.ByRef && sameVariable(x.Var, v)
		case *syntax.Global:
			if scope == nil || local(x) {
				for _, g := range x.Vars {
					found = found || sameVariable(g, v)
				}
			}
		case *syntax.StaticVar:
			found = found || local(x) && sameVariable(x.Var, v)
		case *syntax.Include:
			found = found || local(x)
		case *syntax.FuncCall:
			_, extract := astquery.IsFuncNamedFold(x, "extract")
			_, parse := astquery.IsFuncNamedFold(x, "parse_str")
			found = found || (extract || parse) && local(x)
		}
		return !found
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
		return found
	}
	for _, st := range f.Stmts {
		syntax.Inspect(st, visit)
	}
	return found
}
