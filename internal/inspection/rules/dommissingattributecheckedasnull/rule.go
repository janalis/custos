// Package dommissingattributecheckedasnull implements DomMissingAttributeCheckedAsNull.
package dommissingattributecheckedasnull

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use hasAttribute to test whether an attribute exists."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomMissingAttributeCheckedAsNull" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		null, ok := pair[1].(*syntax.ConstFetch)
		if !ok || !strings.EqualFold(null.Name.Value, "null") {
			continue
		}
		c, ok := semanticquery.NativeValue(ctx, pair[0]).(*syntax.MethodCall)
		if ok && !shortCircuits(c) && semanticquery.NativeMethod(ctx, c, "DOMElement", "getAttribute") {
			ctx.ReportNode(b, message)
			return
		}
	}
}

// shortCircuits reports whether a nullsafe step in the member chain of e can
// make the whole chain evaluate to null without calling its last member.
func shortCircuits(e syntax.Expr) bool {
	for {
		switch x := e.(type) {
		case *syntax.MethodCall:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.PropertyFetch:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.ArrayDimFetch:
			e = x.Var
		default:
			return false
		}
	}
}
