// Package loopclosurecapturesreference implements the native LoopClosureCapturesReference inspection.
package loopclosurecapturesreference

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

const message = "Capture the iteration counter by value for deferred callbacks."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "LoopClosureCapturesReference" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KClosure} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	closure := n.(*syntax.Closure)
	assignment, ok := closure.Parent().(*syntax.Assign)
	if !ok || assignment.ByRef {
		return
	}
	appendTo, ok := assignment.Var.(*syntax.ArrayDimFetch)
	if !ok || appendTo.Dim != nil || astquery.VariableName(appendTo.Var) == "" {
		return
	}
	stmt, ok := assignment.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	block, ok := stmt.Parent().(*syntax.Block)
	if !ok || len(block.Stmts) != 1 {
		return
	}
	loop, ok := block.Parent().(*syntax.For)
	if !ok || loop.Body != block || len(loop.Init) != 1 {
		return
	}
	init, ok := loop.Init[0].(*syntax.Assign)
	if !ok || init.ByRef || astquery.VariableName(init.Var) == "" {
		return
	}
	counter := astquery.VariableName(init.Var)
	captured := false
	for _, use := range closure.Uses {
		if use.ByRef && use.Var.Name == counter {
			captured = true
		}
	}
	if !captured || !readsCounter(closure.Body, counter) {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, loop)
	if !ok {
		return
	}
	consume, ok := next.(*syntax.Foreach)
	if !ok || consume.ByRef || astquery.VariableName(consume.Expr) != astquery.VariableName(appendTo.Var) || astquery.VariableName(consume.Value) == "" {
		return
	}
	called := false
	syntax.Inspect(consume.Body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if c, ok := n.(*syntax.FuncCall); ok && astquery.VariableName(c.Name) == astquery.VariableName(consume.Value) {
			called = true
		}
		return true
	})
	if called {
		ctx.ReportNode(closure, message)
	}
}

func readsCounter(body *syntax.Block, counter string) bool {
	read := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if v, ok := n.(*syntax.Variable); ok && v.Name == counter && !writeOnly(v) {
			read = true
		}
		return true
	})
	return read
}

// writeOnly reports whether the occurrence only stores into or unbinds the
// variable: a plain assignment target, a foreach key or value target, a value
// slot of a destructuring pattern in either position, or an unset() operand.
func writeOnly(v *syntax.Variable) bool {
	var target syntax.Node = v
	for {
		switch parent := target.Parent().(type) {
		case *syntax.ArrayItem:
			if parent.Value != target || parent.Unpack {
				return false
			}
			// The enclosing pattern must itself be a write target.
			target = parent.Parent()
		case *syntax.Assign:
			return parent.Var == target && parent.Op.Kind == syntax.TEqual
		case *syntax.Foreach:
			return parent.Key == target || parent.Value == target
		case *syntax.Unset:
			return true
		default:
			return false
		}
	}
}
