// Package abstractclassinstantiation implements the native AbstractClassInstantiation inspection.
package abstractclassinstantiation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Instantiate a concrete class."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AbstractClassInstantiation" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	v := n.(*syntax.New)
	if selfInAbstractClass(v) {
		ctx.ReportNode(v, message)
		return
	}
	c := semanticquery.NativeNewClass(ctx, v)
	if c != nil && c.Abstract {
		ctx.ReportNode(v, message)
	}
}

// selfInAbstractClass reports whether v is `new self` evaluated by a method
// body of a named abstract class. A closure or arrow function can be rebound
// to another class scope, a trait's self is the using class, and static and
// parent depend on the runtime class or an ancestor, so those stay unknown.
func selfInAbstractClass(v *syntax.New) bool {
	name, ok := v.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(name.Value, "self") {
		return false
	}
	var child syntax.Node = v
	for p := v.Parent(); p != nil; child, p = p, p.Parent() {
		switch p := p.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function:
			return false
		case *syntax.Method:
			if child != syntax.Node(p.Body) {
				return false
			}
			c, ok := p.Parent().(*syntax.ClassLike)
			return ok && c.ClassKind == syntax.KindClass && c.Name != nil && c.Modifiers.Has(syntax.TAbstract)
		}
	}
	return false
}
