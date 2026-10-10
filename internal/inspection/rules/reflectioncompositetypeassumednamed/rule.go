// Package reflectioncompositetypeassumednamed implements the native ReflectionCompositeTypeAssumedNamed inspection.
package reflectioncompositetypeassumednamed

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Inspect the composite reflection type members."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReflectionCompositeTypeAssumedNamed" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	name, ok := call.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(name.Value, "getName") || ctx.PHP < phpversion.PHP80 {
		return
	}
	source, ok := semanticquery.NativeValue(ctx, call.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	method, ok := source.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	typ := ""
	if strings.EqualFold(method.Value, "getReturnType") {
		_, typ, _ = semanticquery.NativeReflectionTarget(ctx, source.Var)
	} else if strings.EqualFold(method.Value, "getType") {
		if param, ok := semanticquery.NativeValue(ctx, source.Var).(*syntax.ArrayDimFetch); ok {
			list, ok := semanticquery.NativeValue(ctx, param.Var).(*syntax.MethodCall)
			if !ok {
				return
			}
			id, ok := list.Name.(*syntax.Identifier)
			if !ok || !strings.EqualFold(id.Value, "getParameters") {
				return
			}
			params, _, known := semanticquery.NativeReflectionTarget(ctx, list.Var)
			i, ik := semanticquery.NativeInt(ctx, param.Dim)
			if !known || !ik || i < 0 || i >= int64(len(params)) {
				return
			}
			typ = params[i].Type
		} else if prop := semanticquery.NativeConstruction(ctx, source.Var, "ReflectionProperty"); prop != nil {
			cl, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(prop.Args, 0, "class"))
			pn, ok2 := semanticquery.NativeString(ctx, semanticquery.CallArgument(prop.Args, 1, "property"))
			if !ok || !ok2 {
				return
			}
			p := ctx.Index().FindProperty(cl, pn, ctx.PHP)
			if p != nil {
				typ = p.Type
			}
		}
	}
	if !strings.Contains(typ, "|") && !(ctx.PHP >= phpversion.PHP81 && strings.Contains(typ, "&")) {
		return
	}
	for p := call.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if b, ok := p.(*syntax.If); ok && b.Body.Span().Contains(call.Span()) {
			inst, ok := syntax.UnwrapParens(b.Cond).(*syntax.Instanceof)
			if ok {
				cl, ok := inst.Class.(*syntax.Name)
				if ok && strings.EqualFold(ctx.Names().Class(cl.Value, cl.Span().Start), "ReflectionNamedType") && semanticquery.NativeSameValue(ctx, inst.Expr, call.Var) {
					return
				}
			}
		}
	}
	ctx.ReportNode(call, message)
}
