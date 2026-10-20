// Package reflectioncompositetypeassumednamed implements the native ReflectionCompositeTypeAssumedNamed inspection.
package reflectioncompositetypeassumednamed

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
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
	if !composite(typ, ctx.PHP) || guarded(ctx, call, typ) {
		return
	}
	ctx.ReportNode(call, message)
}

// composite reports whether a declared type reflects as a union or
// intersection type. A leading `?` and a `null` member only make a single
// named type nullable, which still reflects as ReflectionNamedType.
func composite(typ string, php phpversion.Version) bool {
	if strings.Contains(typ, "&") {
		return php >= phpversion.PHP81
	}
	members := 0
	for m := range strings.SplitSeq(strings.TrimPrefix(typ, "?"), "|") {
		if !strings.EqualFold(strings.TrimSpace(m), "null") {
			members++
		}
	}
	return members > 1
}

// guarded reports whether the conditions guarding the call prove the
// receiver a named type. The value is the composite class the declared type
// reflects as; its parent rides along so that an `instanceof
// ReflectionNamedType` test narrows to that class rather than to nothing.
func guarded(ctx *analysis.Context, call *syntax.MethodCall, typ string) bool {
	class := `\ReflectionUnionType`
	if !strings.Contains(typ, "|") {
		class = `\ReflectionIntersectionType`
	}
	if semanticquery.NativeLabelled(ctx, call.Var) {
		return false // a jump may land between a guard and the call
	}
	t := ctx.Types().Native().NarrowedFrom(call.Var, types.Of(class, `\ReflectionType`))
	return !t.IsUnknown() && !t.Has(class)
}
