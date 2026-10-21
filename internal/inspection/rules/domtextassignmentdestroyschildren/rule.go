// Package domtextassignmentdestroyschildren implements DomTextAssignmentDestroysChildren.
package domtextassignmentdestroyschildren

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve child nodes when adding element text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomTextAssignmentDestroysChildren" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	a := n.(*syntax.Assign)
	p, ok := a.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	id, ok := p.Name.(*syntax.Identifier)
	if !ok || (id.Value != "nodeValue" && id.Value != "textContent") {
		return
	}
	if !semanticquery.ExpansionDOMElementChildren(ctx, p.Var, a) && !documentElementHasChildren(ctx, p.Var, a) && !createdBeforeWrite(ctx, p.Var, a) {
		return
	}
	if childMayBeDetached(ctx, p.Var, a) {
		return
	}
	ctx.ReportNode(a, message)
}

// keepsChildAttached lists element methods that never move or remove their
// receiver, so calling them on an appended child keeps the child proof.
var keepsChildAttached = map[string]bool{
	"appendchild": true, "getattribute": true, "getattributens": true, "hasattribute": true,
	"hasattributens": true, "hasattributes": true, "haschildnodes": true, "removeattribute": true,
	"removeattributens": true, "setattribute": true, "setattributens": true,
}

// childMayBeDetached reports whether a variable appended to the receiver is
// later used in a way that could move or remove that node: the shared proofs
// only watch statements naming the parent. Rebinding the variable is harmless.
func childMayBeDetached(ctx *analysis.Context, receiver syntax.Expr, at syntax.Node) bool {
	parent, ok := receiver.(*syntax.Variable)
	if !ok {
		return false
	}
	appended := map[string]bool{}
	detached := false
	scan := func(root syntax.Node) {
		syntax.Inspect(root, func(n syntax.Node) bool {
			v, ok := n.(*syntax.Variable)
			if !ok || !appended[v.Name] {
				return true
			}
			switch use := v.Parent().(type) {
			case *syntax.PropertyFetch:
				if use.Var == syntax.Node(v) {
					return true
				}
			case *syntax.MethodCall:
				if id, ok := use.Name.(*syntax.Identifier); ok && use.Var == syntax.Node(v) && keepsChildAttached[strings.ToLower(id.Value)] {
					return true
				}
			}
			detached = true
			return false
		})
	}
	for _, statement := range flowquery.NativePriorStatements(ctx.File, at) {
		if st, ok := statement.(*syntax.ExprStmt); ok {
			switch x := st.Expr.(type) {
			case *syntax.Assign:
				if target, ok := x.Var.(*syntax.Variable); ok && !x.ByRef && x.Op.Kind == syntax.TEqual && appended[target.Name] {
					scan(x.Value)
					delete(appended, target.Name)
					continue
				}
			case *syntax.MethodCall:
				target, onParent := x.Var.(*syntax.Variable)
				child, named := semanticquery.CallArgument(x.Args, 0, "node").(*syntax.Variable)
				if onParent && named && target.Name == parent.Name && semanticquery.NativeMethod(ctx, x, "DOMNode", "appendChild") {
					appended[child.Name] = true
					continue
				}
			}
		}
		scan(statement)
	}
	return detached
}

// createdBeforeWrite proves children of a directly constructed
// local element before a write to that element's properties.
func createdBeforeWrite(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	variable, ok := e.(*syntax.Variable)
	if !ok {
		return false
	}
	statements := flowquery.NativePriorStatements(ctx.File, at)
	if len(statements) > 256 {
		return false
	}
	element, children := false, false
	for _, statement := range statements {
		st, ok := statement.(*syntax.ExprStmt)
		if !ok {
			element, children = false, false
			continue
		}
		if a, ok := st.Expr.(*syntax.Assign); ok {
			if target, ok := a.Var.(*syntax.Variable); ok && target.Name == variable.Name {
				c, ok := a.Value.(*syntax.MethodCall)
				element = ok && !a.ByRef && a.Op.Kind == syntax.TEqual && (semanticquery.NativeMethod(ctx, c, "DOMDocument", "createElement") || semanticquery.NativeMethod(ctx, c, "DOMDocument", "createElementNS"))
				children = false
				continue
			}
		}
		if c, ok := st.Expr.(*syntax.MethodCall); ok {
			if target, ok := c.Var.(*syntax.Variable); ok && target.Name == variable.Name && element && semanticquery.NativeMethod(ctx, c, "DOMNode", "appendChild") {
				child, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "node")).(*syntax.MethodCall)
				children = children || ok && (semanticquery.NativeMethod(ctx, child, "DOMDocument", "createElement") || semanticquery.NativeMethod(ctx, child, "DOMDocument", "createElementNS"))
				continue
			}
		}
		syntax.Inspect(statement, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == variable.Name {
				element, children = false, false
			}
			return true
		})
	}
	return element && children
}

// documentElementHasChildren handles a DOM assignment's left-hand receiver, which
// has no read value in the flow model. Only direct, nonescaping statements prove state.
func documentElementHasChildren(ctx *analysis.Context, e syntax.Expr, at syntax.Node) bool {
	p, ok := e.(*syntax.PropertyFetch)
	if !ok {
		return false
	}
	field, ok := p.Name.(*syntax.Identifier)
	if !ok || field.Value != "documentElement" {
		return false
	}
	v, ok := p.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	statements := flowquery.NativePriorStatements(ctx.File, at)
	if len(statements) > 256 {
		return false
	}
	document, children := false, false
	for _, statement := range statements {
		st, ok := statement.(*syntax.ExprStmt)
		if !ok {
			document, children = false, false
			continue
		}
		if a, ok := st.Expr.(*syntax.Assign); ok {
			if target, ok := a.Var.(*syntax.Variable); ok && target.Name == v.Name {
				creation, ok := a.Value.(*syntax.New)
				document = ok && !a.ByRef && a.Op.Kind == syntax.TEqual && semanticquery.NamesGlobalClass(ctx, creation.Class, "DOMDocument")
				children = false
				continue
			}
		}
		if c, ok := st.Expr.(*syntax.MethodCall); ok {
			if target, ok := c.Var.(*syntax.Variable); ok && target.Name == v.Name && document && semanticquery.NativeMethod(ctx, c, "DOMDocument", "loadXML") {
				s, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "source"))
				children = known && semanticquery.XMLChildren(s)
				continue
			}
		}
		syntax.Inspect(statement, func(n syntax.Node) bool {
			if x, ok := n.(*syntax.Variable); ok && x.Name == v.Name {
				document, children = false, false
			}
			return true
		})
	}
	return document && children
}
