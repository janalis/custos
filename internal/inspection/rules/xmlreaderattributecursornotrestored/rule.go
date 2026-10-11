// Package xmlreaderattributecursornotrestored implements XmlReaderAttributeCursorNotRestored.
package xmlreaderattributecursornotrestored

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Restore the XMLReader element cursor before element operations."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlReaderAttributeCursorNotRestored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical {
		return
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		property, ok := pair[0].(*syntax.PropertyFetch)
		if !ok {
			continue
		}
		name, ok := property.Name.(*syntax.Identifier)
		if !ok || name.Value != "nodeType" {
			continue
		}
		constant, ok := pair[1].(*syntax.ClassConstFetch)
		if !ok || !semanticquery.NamesGlobalClass(ctx, constant.Class, "XMLReader") {
			continue
		}
		field, ok := constant.Name.(*syntax.Identifier)
		if !ok || field.Value != "ELEMENT" {
			continue
		}
		for p := b.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
			branch, ok := p.(*syntax.If)
			if !ok || !branch.Body.Span().Contains(b.Span()) {
				continue
			}
			move, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
			if !ok || (!semanticquery.NativeMethod(ctx, move, "XMLReader", "moveToAttribute") && !semanticquery.NativeMethod(ctx, move, "XMLReader", "moveToFirstAttribute")) {
				continue
			}
			if !same(ctx, move.Var, property.Var) {
				continue
			}
			if !restored(ctx, b, property.Var, move) {
				ctx.ReportNode(b, message)
				return
			}
		}
	}
}

func same(ctx *analysis.Context, a, b syntax.Expr) bool {
	return semanticquery.ExpansionSameObject(ctx, a, b)
}

// restored reports whether a cursor move that leaves the attribute dominates
// the comparison and runs after the guarding attribute move. Calls before the
// guard cannot undo it.
func restored(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, move *syntax.MethodCall) bool {
	for _, call := range semanticquery.NativePriorCalls(ctx, at, receiver, "moveToElement", "read", "next") {
		if call.Span().Start >= move.Span().End {
			return true
		}
	}
	return false
}
