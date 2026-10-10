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
	if !ok || appendTo.Dim != nil || variable(appendTo.Var) == "" {
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
	if !ok || init.ByRef || variable(init.Var) == "" {
		return
	}
	counter := variable(init.Var)
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
	if !ok || consume.ByRef || variable(consume.Expr) != variable(appendTo.Var) || variable(consume.Value) == "" {
		return
	}
	called := false
	syntax.Inspect(consume.Body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if c, ok := n.(*syntax.FuncCall); ok && variable(c.Name) == variable(consume.Value) {
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
		v, ok := n.(*syntax.Variable)
		if ok && v.Name == counter {
			assignment, assigned := v.Parent().(*syntax.Assign)
			if !assigned || assignment.Var != v || assignment.Op.Kind != syntax.TEqual {
				read = true
			}
		}
		return true
	})
	return read
}

func variable(e syntax.Expr) string {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok {
		return ""
	}
	return v.Name
}
