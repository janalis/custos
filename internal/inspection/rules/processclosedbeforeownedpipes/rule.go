// Package processclosedbeforeownedpipes implements the native ProcessClosedBeforeOwnedPipes inspection.
package processclosedbeforeownedpipes

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Close owned process pipes before waiting for process exit."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ProcessClosedBeforeOwnedPipes" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "proc_close") {
		return
	}
	target, ok := semanticquery.CallArgument(c.Args, 0, "process").(*syntax.Variable)
	if !ok {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) == 0 {
		return
	}
	st, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := st.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok || variable.Name != target.Name {
		return
	}
	open, ok := assignment.Value.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, open, "proc_open") {
		return
	}
	if _, ok := semanticquery.CallArgument(open.Args, 2, "pipes").(*syntax.Variable); !ok {
		return
	}
	descriptors, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(open.Args, 1, "descriptor_spec"))
	if !known {
		return
	}
	found := false
	for _, entry := range descriptors {
		tuple, known := semanticquery.NativeArrayEntries(ctx, entry)
		if !known {
			return
		}
		kind, known := semanticquery.NativeString(ctx, tuple["i:0"])
		if !known {
			return
		}
		if kind == "pipe" {
			mode, known := semanticquery.NativeString(ctx, tuple["i:1"])
			if !known || mode != "r" && mode != "w" {
				return
			}
			found = true
		}
	}
	if found {
		ctx.ReportNode(c, message)
	}
}
