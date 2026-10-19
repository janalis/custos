// Package jsonencodenonfinitenumber implements the native JsonEncodeNonFiniteNumber inspection.
package jsonencodenonfinitenumber

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Encode finite numbers in JSON."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonEncodeNonFiniteNumber" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 1, "flags")
	if flags != nil {
		bits, k := semanticquery.NativeContractInt(ctx, flags)
		if !k || bits&512 != 0 {
			return
		}
	}
	if nonfinite(ctx, c, semanticquery.CallArgument(c.Args, 0, "value"), 0) {
		ctx.ReportNode(c, message)
	}
}

func nonfinite(ctx *analysis.Context, call *syntax.FuncCall, e syntax.Expr, depth int) bool {
	if depth > 32 || rewritten(ctx, call, e) {
		return false
	}
	e = semanticquery.NativeValue(ctx, e)
	if c, ok := e.(*syntax.ConstFetch); ok {
		name := semanticquery.GlobalConstName(ctx, c)
		return name == "INF" || name == "NAN"
	}
	if u, ok := e.(*syntax.Unary); ok && (u.Op.Kind == syntax.TMinus || u.Op.Kind == syntax.TPlus) {
		return nonfinite(ctx, call, u.Expr, depth+1)
	}
	if values, k := semanticquery.NativeArrayEntries(ctx, e); k {
		for _, v := range values {
			if nonfinite(ctx, call, v, depth+1) {
				return true
			}
		}
	}
	return false
}

// rewritten reports whether e is a plain variable that its scope changes in
// a way the tracked value does not follow: a foreach that iterates it by
// reference or writes into it, a by-reference closure capture, or extract().
// The scan ignores statement order.
func rewritten(ctx *analysis.Context, call *syntax.FuncCall, e syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return false
	}
	found := false
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Foreach:
			found = found || n.ByRef && astquery.MentionsVariable(n.Expr, v.Name) ||
				astquery.MentionsVariable(n.Key, v.Name) || astquery.MentionsVariable(n.Value, v.Name)
		case *syntax.ClosureUse:
			found = found || n.ByRef && n.Var.Name == v.Name
		case *syntax.FuncCall:
			_, extract := astquery.IsFuncNamedFold(n, "extract")
			found = found || extract
		}
		return !found
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}
