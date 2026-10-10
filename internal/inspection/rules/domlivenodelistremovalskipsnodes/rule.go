// Package domlivenodelistremovalskipsnodes implements DomLiveNodeListRemovalSkipsNodes.
package domlivenodelistremovalskipsnodes

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Remove live DOM matches without skipping shifted nodes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomLiveNodeListRemovalSkipsNodes" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMNode", "removeChild") {
		return
	}
	child := semanticquery.CallArgument(c.Args, 0, "child")
	item, ok := semanticquery.NativeValue(ctx, child).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, item, "DOMNodeList", "item") {
		return
	}
	index := semanticquery.CallArgument(item.Args, 0, "index")
	variable, ok := index.(*syntax.Variable)
	if !ok {
		return
	}
	list, ok := semanticquery.NativeValue(ctx, item.Var).(*syntax.MethodCall)
	if !ok || (!semanticquery.NativeMethod(ctx, list, "DOMDocument", "getElementsByTagName") && !semanticquery.NativeMethod(ctx, list, "DOMElement", "getElementsByTagName") && !semanticquery.NativeMethod(ctx, list, "DOMDocument", "getElementsByTagNameNS") && !semanticquery.NativeMethod(ctx, list, "DOMElement", "getElementsByTagNameNS")) {
		return
	}
	parent, ok := c.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	property, ok := parent.Name.(*syntax.Identifier)
	if !ok || property.Value != "parentNode" || !same(ctx, parent.Var, child) {
		return
	}
	for p := c.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		loop, ok := p.(*syntax.For)
		if !ok {
			continue
		}
		if len(loop.Init) != 1 || len(loop.Cond) != 1 || len(loop.Loop) != 1 {
			return
		}
		init, ok := loop.Init[0].(*syntax.Assign)
		if !ok || !sameVariable(init.Var, variable) {
			return
		}
		zero, known := semanticquery.NativeInt(ctx, init.Value)
		if !known || zero != 0 {
			return
		}
		condition, ok := loop.Cond[0].(*syntax.Binary)
		if !ok || condition.Op.Kind != syntax.TLess || !sameVariable(condition.Left, variable) {
			return
		}
		length, ok := condition.Right.(*syntax.PropertyFetch)
		if !ok {
			return
		}
		field, ok := length.Name.(*syntax.Identifier)
		if !ok || field.Value != "length" || !same(ctx, length.Var, item.Var) {
			return
		}
		inc, ok := loop.Loop[0].(*syntax.IncDec)
		if !ok || inc.Op.Kind != syntax.TInc || !sameVariable(inc.Var, variable) {
			return
		}
		body, ok := loop.Body.(*syntax.Block)
		if !ok {
			return
		}
		for _, s := range body.Stmts {
			statement, ok := s.(*syntax.ExprStmt)
			if !ok {
				return
			}
			if statement.Expr == c {
				ctx.ReportNode(c, message)
				return
			}
			if assignment, ok := statement.Expr.(*syntax.Assign); !ok || assignment.Value != item {
				return
			}
		}
	}
}

func sameVariable(a syntax.Expr, b *syntax.Variable) bool {
	v, ok := a.(*syntax.Variable)
	return ok && v.Name == b.Name
}

func same(ctx *analysis.Context, a, b syntax.Expr) bool {
	return semanticquery.ExpansionSameObject(ctx, a, b)
}
