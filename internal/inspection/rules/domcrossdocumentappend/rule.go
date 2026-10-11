// Package domcrossdocumentappend implements DomCrossDocumentAppend.
package domcrossdocumentappend

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Import the node into the destination document."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomCrossDocumentAppend" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMNode", "appendChild") && !semanticquery.NativeMethod(ctx, c, "DOMNode", "insertBefore") && !semanticquery.NativeMethod(ctx, c, "DOMNode", "replaceChild") {
		return
	}
	a := owner(ctx, c, c.Var)
	b := owner(ctx, c, semanticquery.CallArgument(c.Args, 0, "node"))
	if a != nil && b != nil && a != b {
		ctx.ReportNode(c, message)
	}
}

// owner returns the document owning e when at runs. Creation provenance is
// final for documents and inline expressions; a node held in a variable may
// have been moved by an earlier adoptNode() statement, whose receiver then
// owns it. Any other earlier adoption leaves the owner unknown, because it
// may be conditional or may have moved an ancestor of the node.
func owner(ctx *analysis.Context, at *syntax.MethodCall, e syntax.Expr) *syntax.New {
	doc := semanticquery.ExpansionDOMOwner(ctx, e)
	if _, ok := syntax.UnwrapParens(e).(*syntax.Variable); !ok || doc == nil {
		return doc
	}
	value := semanticquery.NativeValue(ctx, e)
	if value == syntax.Expr(doc) {
		return doc
	}
	var last *syntax.MethodCall
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		p, ok := record.Node.(*syntax.MethodCall)
		if !ok || p.Span().End > at.Span().Start || !semanticquery.NativeMethod(ctx, p, "DOMDocument", "adoptNode") {
			continue
		}
		if last == nil || p.Span().Start > last.Span().Start {
			last = p
		}
	}
	if last == nil || value == syntax.Expr(last) {
		return doc
	}
	if !semanticquery.NativeDominates(last, at) || !semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(last.Args, 0, "node"), e) {
		return nil
	}
	return semanticquery.ExpansionDOMOwner(ctx, last.Var)
}
