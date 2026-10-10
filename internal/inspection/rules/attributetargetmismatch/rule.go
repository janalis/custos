// Package attributetargetmismatch implements the native AttributeTargetMismatch inspection.
package attributetargetmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use the attribute on a permitted target."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AttributeTargetMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAttribute} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Attribute)
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	_, flags, known := semanticquery.NativeAttributeFlags(ctx, a)
	if !known {
		return
	}
	var bit int64
	switch a.Parent().Parent().(type) {
	case *syntax.ClassLike:
		bit = 1
	case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction:
		bit = 2
	case *syntax.Method, *syntax.PropertyHook:
		bit = 4
	case *syntax.Property:
		bit = 8
	case *syntax.ClassConst, *syntax.EnumCase:
		bit = 16
	case *syntax.Param:
		bit = 32
		param := a.Parent().Parent().(*syntax.Param)
		if param.Modifiers.Has(syntax.TPublic) || param.Modifiers.Has(syntax.TProtected) || param.Modifiers.Has(syntax.TPrivate) {
			bit |= 8
		}
	default:
		return
	}
	if flags&bit == 0 {
		ctx.ReportNode(a.Parent(), message)
	}
}
