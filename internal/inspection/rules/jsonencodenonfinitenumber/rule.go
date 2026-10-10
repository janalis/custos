// Package jsonencodenonfinitenumber implements the native JsonEncodeNonFiniteNumber inspection.
package jsonencodenonfinitenumber

import (
	"custos/internal/inspection/analysis"
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
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
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
	if nonfinite(ctx, semanticquery.CallArgument(c.Args, 0, "value"), 0) {
		ctx.ReportNode(c, message)
	}
}

func nonfinite(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if depth > 32 {
		return false
	}
	e = semanticquery.NativeValue(ctx, e)
	if c, ok := e.(*syntax.ConstFetch); ok {
		name := semanticquery.GlobalConstName(ctx, c)
		return name == "INF" || name == "NAN"
	}
	if u, ok := e.(*syntax.Unary); ok && (u.Op.Kind == syntax.TMinus || u.Op.Kind == syntax.TPlus) {
		return nonfinite(ctx, u.Expr, depth+1)
	}
	if values, k := semanticquery.NativeArrayEntries(ctx, e); k {
		for _, v := range values {
			if nonfinite(ctx, v, depth+1) {
				return true
			}
		}
	}
	return false
}
