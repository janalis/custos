// Package xpathqueryfailuredereferenced implements XPathQueryFailureDereferenced.
package xpathqueryfailuredereferenced

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check the XPath query result before accessing its nodes."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "XPathQueryFailureDereferenced" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KMethodCall}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var receiver syntax.Expr
	switch c := n.(type) {
	case *syntax.PropertyFetch:
		receiver = c.Var
	case *syntax.MethodCall:
		receiver = c.Var
	}
	q, ok := semanticquery.NativeValue(ctx, receiver).(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, q, "DOMXPath", "query") && !ctx.Flow().Excludes(receiver, "false") && (semanticquery.NativeLabelled(ctx, receiver) || !guarded(ctx, receiver)) {
		ctx.ReportNode(n, message)
	}
}

// guarded proves that a condition already evaluated on every path to the
// dereference excludes false. The result is either false or a DOMNodeList, so
// a truthiness or instanceof test separates the two as well as a comparison.
func guarded(ctx *analysis.Context, receiver syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(receiver).(*syntax.Variable)
	if !ok {
		return false
	}
	g := guard{ctx: ctx, variable: variable}
	inside := func(part syntax.Node) bool { return part != nil && part.Span().Contains(receiver.Span()) }
	for p := receiver.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch c := p.(type) {
		case *syntax.Binary:
			if inside(c.Right) && ((c.Op.Kind == syntax.TBooleanAnd && g.proves(c.Left, true)) || (c.Op.Kind == syntax.TBooleanOr && g.proves(c.Left, false))) {
				return true
			}
		case *syntax.Ternary:
			if (inside(c.Then) && g.proves(c.Cond, true)) || (inside(c.Else) && g.proves(c.Cond, false)) {
				return true
			}
		case *syntax.If:
			if g.branch(c, inside) {
				return true
			}
		}
		if st, ok := p.(syntax.Stmt); ok && g.exited(st) {
			return true
		}
	}
	return false
}

type guard struct {
	ctx      *analysis.Context
	variable *syntax.Variable
}

// branch reports whether reaching the dereference inside c required a
// condition outcome that excludes false: the taken condition held, or one of
// the conditions tested before it did not.
func (g guard) branch(c *syntax.If, inside func(syntax.Node) bool) bool {
	if inside(c.Cond) {
		return false
	}
	if inside(c.Body) {
		return g.proves(c.Cond, true)
	}
	if g.proves(c.Cond, false) {
		return true
	}
	for _, alternative := range c.ElseIfs {
		if inside(alternative) {
			return inside(alternative.Body) && g.proves(alternative.Cond, true)
		}
		if g.proves(alternative.Cond, false) {
			return true
		}
	}
	return false
}

// exited reports whether an earlier statement of the same list leaves the
// path whenever the result may be false.
func (g guard) exited(st syntax.Stmt) bool {
	list, i, _ := astquery.StmtList(g.ctx.File, st)
	for _, previous := range list[:max(i, 0)] {
		exit, ok := previous.(*syntax.If)
		if ok && exit.Else == nil && len(exit.ElseIfs) == 0 && syntax.Terminates(exit.Body) && g.proves(exit.Cond, false) {
			return true
		}
	}
	return false
}

// proves reports whether cond evaluating to truth implies a nonfalse result.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && g.proves(c.Expr, !truth)
	case *syntax.Instanceof:
		return truth && g.same(c.Expr)
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr:
			left, right := g.proves(c.Left, truth), g.proves(c.Right, truth)
			if truth == (c.Op.Kind == syntax.TBooleanAnd) {
				return left || right
			}
			return left && right
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.sentinel(c)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.sentinel(c)
		}
		return false
	}
	return truth && g.same(cond)
}

// sentinel matches a comparison of the result with the false constant.
func (g guard) sentinel(c *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{c.Left, c.Right}, {c.Right, c.Left}} {
		constant, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && constant.Name != nil && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "false") && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same ties a tested operand to the producer reaching the dereference: the
// variable with an identical flow value, or the assignment that binds it.
func (g guard) same(e syntax.Expr) bool {
	e = syntax.UnwrapParens(e)
	flow := g.ctx.Flow()
	reaching := flow.Value(g.variable)
	if assign, ok := e.(*syntax.Assign); ok {
		target, ok := assign.Var.(*syntax.Variable)
		return ok && target.Name == g.variable.Name && assign.Op.Kind == syntax.TEqual && !assign.ByRef &&
			syntax.UnwrapParens(assign.Value) == syntax.UnwrapParens(reaching.Expr)
	}
	tested, ok := e.(*syntax.Variable)
	if !ok || tested.Name != g.variable.Name {
		return false
	}
	value := flow.Value(tested)
	return value.Expr == reaching.Expr && value.Invalidated == reaching.Invalidated
}
