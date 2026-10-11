// Package enuminstantiation implements the native EnumInstantiation inspection.
package enuminstantiation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use an enum case instead of instantiation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "EnumInstantiation" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	v := n.(*syntax.New)
	if c := semanticquery.NativeNewClass(ctx, v); c != nil && c.Kind == syntax.KindEnum || selfInEnum(v) {
		ctx.ReportNode(v, message)
	}
}

// selfInEnum reports whether n instantiates self or static lexically inside an
// enum body. A closure can be rebound to another class scope, so the walk gives
// up there; an anonymous class or a trait is its own scope and is never an enum.
func selfInEnum(n *syntax.New) bool {
	name, ok := n.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(name.Value, "self") && !strings.EqualFold(name.Value, "static") {
		return false
	}
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch c := p.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.ClassLike:
			return c.ClassKind == syntax.KindEnum
		}
	}
	return false
}
