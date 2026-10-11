// Package domappendmovesnodeunexpectedly implements DomAppendMovesNodeUnexpectedly.
package domappendmovesnodeunexpectedly

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clone the node when preserving its existing parent."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomAppendMovesNodeUnexpectedly" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMNode", "appendChild") {
		return
	}
	node := semanticquery.CallArgument(c.Args, 0, "node")
	var previous *syntax.MethodCall
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		p, ok := record.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeDominates(p, c) {
			continue
		}
		if detaches(ctx, p, node) || previous != nil && releases(ctx, p, previous.Var, node) {
			previous = nil
			continue
		}
		if semanticquery.NativeMethod(ctx, p, "DOMNode", "appendChild") && semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(p.Args, 0, "node"), node) {
			previous = p
		}

	}
	if previous != nil {
		old := semanticquery.ExpansionDOMOwner(ctx, previous.Var)
		current := semanticquery.ExpansionDOMOwner(ctx, c.Var)
		if old != nil && current != nil && !semanticquery.ExpansionSameObject(ctx, previous.Var, c.Var) {
			ctx.ReportNode(c, message)
		}

	}
}

// detaches reports whether p deliberately takes node out of its parent: the
// node removes or replaces itself, or it is named in a removal or replacement.
func detaches(ctx *analysis.Context, p *syntax.MethodCall, node syntax.Expr) bool {
	if semanticquery.NativeMethod(ctx, p, "DOMNode", "remove") || semanticquery.NativeMethod(ctx, p, "DOMNode", "replaceWith") {
		return semanticquery.ExpansionSameObject(ctx, p.Var, node)
	}
	if semanticquery.NativeMethod(ctx, p, "DOMNode", "removeChild") {
		return semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(p.Args, 0, "child"), node)
	}
	if semanticquery.NativeMethod(ctx, p, "DOMNode", "replaceChild") {
		return semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(p.Args, 0, "node"), node) ||
			semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(p.Args, 1, "child"), node)
	}
	return false
}

// releases reports whether p may drop node from the parent that received it.
// Removing a child keeps the pending append only when that child is proven to
// be another locally created node.
func releases(ctx *analysis.Context, p *syntax.MethodCall, parent, node syntax.Expr) bool {
	position := 0
	switch {
	case semanticquery.NativeMethod(ctx, p, "DOMNode", "replaceChildren"):
		return semanticquery.ExpansionSameObject(ctx, p.Var, parent)
	case semanticquery.NativeMethod(ctx, p, "DOMNode", "removeChild"):
	case semanticquery.NativeMethod(ctx, p, "DOMNode", "replaceChild"):
		position = 1
	default:
		return false
	}
	if !semanticquery.ExpansionSameObject(ctx, p.Var, parent) {
		return false
	}
	removed, kept := created(ctx, semanticquery.CallArgument(p.Args, position, "child")), created(ctx, node)
	return removed == nil || kept == nil || removed == kept
}

// created returns the call that allocates e as a fresh node, or nil when the
// value has another or an unknown origin.
func created(ctx *analysis.Context, e syntax.Expr) *syntax.MethodCall {
	call, ok := semanticquery.NativeValue(ctx, e).(*syntax.MethodCall)
	if !ok {
		return nil
	}
	if semanticquery.NativeMethod(ctx, call, "DOMNode", "cloneNode") {
		return call
	}
	for _, name := range []string{"createElement", "createElementNS", "createTextNode", "createComment", "createCDATASection", "createDocumentFragment"} {
		if semanticquery.NativeMethod(ctx, call, "DOMDocument", name) {
			return call
		}
	}
	return nil
}
