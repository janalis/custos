package domshallowimportdropsdescendants

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// role says what a tracked local variable holds while the proof is checked.
type role uint8

const (
	document role = iota + 1 // owns or holds the imported element
	element                  // the imported element itself
	child                    // a node appended to the imported element
)

// Calls that leave the children of the imported element untouched.
var (
	documentReaders = []string{"createElement", "createElementNS", "createTextNode", "createComment", "createCDATASection", "createAttribute", "createDocumentFragment", "importNode", "saveXML", "saveHTML"}
	// The shared proof already orders these when they run on every path.
	documentWriters = []string{"loadXML", "load", "appendChild", "removeChild"}
	elementCalls    = []string{"setAttribute", "setAttributeNS", "getAttribute", "hasAttribute"}
	nodeCalls       = []string{"appendChild", "hasChildNodes"}
)

type proof struct {
	ctx     *analysis.Context
	at      *syntax.MethodCall
	tracked map[string]role
	stale   bool
}

// stale reports whether the children proven for the imported node may be gone
// at the import: every use of the element, of the document holding it and of
// its appended children must be one that provably keeps them in place.
func stale(ctx *analysis.Context, node syntax.Expr, at *syntax.MethodCall) bool {
	p := &proof{ctx: ctx, at: at, tracked: map[string]role{}}
	switch x := node.(type) {
	case *syntax.Variable:
		p.track(x, element)
	case *syntax.PropertyFetch:
		owner, ok := x.Var.(*syntax.Variable)
		if !ok {
			return true
		}
		p.track(owner, document)
	default:
		return true
	}
	scope := syntax.EnclosingVariableScope(at)
	limit := at.Span().End
	if inLoop(at, scope) {
		limit = ^uint32(0)
	}
	var uses []*syntax.Variable
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Variable:
			if x.Span().Start < limit {
				uses = append(uses, x)
			}
		case *syntax.Closure:
			// Captured objects stay shared with the closure body.
			for _, use := range x.Uses {
				uses = append(uses, use.Var)
			}
		}
		// Arrow functions read the enclosing variables directly.
		if _, arrow := n.(*syntax.ArrowFunction); !arrow && n != scope && syntax.IsVariableScope(n) {
			return false
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	// Nodes appended to the element, and a document the element is attached to,
	// can change its children as well.
	for _, v := range uses {
		if p.tracked[v.Name] != element {
			continue
		}
		if call, ok := v.Parent().(*syntax.MethodCall); ok && call.Var == v && semanticquery.NativeMethod(ctx, call, "DOMNode", "appendChild") {
			if appended, ok := semanticquery.CallArgument(call.Args, 0, "node").(*syntax.Variable); ok {
				p.track(appended, child)
			}
		}
		if call := argumentOf(v); call != nil && semanticquery.NativeMethod(ctx, call, "DOMDocument", "appendChild") {
			if holder, ok := call.Var.(*syntax.Variable); ok {
				p.track(holder, document)
			}
		}
	}
	for _, v := range uses {
		if r := p.tracked[v.Name]; r != 0 && !p.stale {
			p.stale = !p.benign(v, r)
		}
	}
	return p.stale
}

func (p *proof) track(v *syntax.Variable, r role) {
	if had := p.tracked[v.Name]; had != 0 && had != r {
		p.stale = true
	}
	p.tracked[v.Name] = r
}

// benign accepts only the uses known to keep the proven children in place.
func (p *proof) benign(v *syntax.Variable, r role) bool {
	switch parent := v.Parent().(type) {
	case *syntax.Assign:
		if parent.Var != v || parent.ByRef || parent.Op.Kind != syntax.TEqual {
			return false
		}
		// Binding another variable or a fetched node would alias an untracked one.
		switch parent.Value.(type) {
		case *syntax.New, *syntax.MethodCall:
			return true
		}
		return r == child
	case *syntax.MethodCall:
		if parent.Var != v {
			return false
		}
		switch r {
		case document:
			return p.calls(parent, "DOMDocument", documentReaders) || (p.calls(parent, "DOMDocument", documentWriters) && semanticquery.NativeDominates(parent, p.at))
		default:
			return p.calls(parent, "DOMElement", elementCalls) || p.calls(parent, "DOMNode", nodeCalls)
		}
	case *syntax.PropertyFetch:
		// A document's root element is accepted only as the imported node.
		if id, ok := parent.Name.(*syntax.Identifier); ok && parent.Var == v && r == document && id.Value == "documentElement" {
			call := argumentOf(parent)
			return call != nil && semanticquery.NativeMethod(p.ctx, call, "DOMDocument", "importNode")
		}
	case *syntax.Arg:
		call := argumentOf(v)
		if call == nil {
			return false
		}
		receiver, _ := call.Var.(*syntax.Variable)
		switch r {
		case element:
			if semanticquery.NativeMethod(p.ctx, call, "DOMDocument", "importNode") {
				return true
			}
			return receiver != nil && p.tracked[receiver.Name] == document && semanticquery.NativeMethod(p.ctx, call, "DOMDocument", "appendChild")
		case child:
			return receiver != nil && p.tracked[receiver.Name] == element && semanticquery.NativeMethod(p.ctx, call, "DOMNode", "appendChild")
		}
	}
	return false
}

func (p *proof) calls(c *syntax.MethodCall, class string, names []string) bool {
	for _, name := range names {
		if semanticquery.NativeMethod(p.ctx, c, class, name) {
			return true
		}
	}
	return false
}

// argumentOf returns the method call receiving e as its node argument.
func argumentOf(e syntax.Expr) *syntax.MethodCall {
	arg, ok := e.Parent().(*syntax.Arg)
	if !ok {
		return nil
	}
	call, ok := arg.Parent().Parent().(*syntax.MethodCall)
	if !ok || semanticquery.CallArgument(call.Args, 0, "node") != e {
		return nil
	}
	return call
}

func inLoop(at, scope syntax.Node) bool {
	for n := at.Parent(); n != nil && n != scope; n = n.Parent() {
		switch n.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			return true
		}
	}
	return false
}
