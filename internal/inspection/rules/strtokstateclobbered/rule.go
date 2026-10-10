// Package strtokstateclobbered implements the StrtokStateClobbered inspection.
package strtokstateclobbered

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StrtokStateClobbered" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Keep strtok sequences separate."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "strtok")
	if call == nil || call.Args == nil || len(call.Args.Args) != 1 {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, call)
	if len(prior) < 2 {
		return
	}
	var starts []*syntax.FuncCall
	for i := len(prior) - 1; i >= len(prior)-2; i-- {
		statement, ok := prior[i].(*syntax.ExprStmt)
		if !ok {
			return
		}
		expression := statement.Expr
		if assign, ok := expression.(*syntax.Assign); ok {
			expression = assign.Value
		}
		start, _ := semanticquery.GlobalCall(ctx, expression, "strtok")
		if start == nil || start.Args == nil || len(start.Args.Args) != 2 {
			return
		}
		starts = append(starts, start)
	}
	if astquery.Equivalent(ctx.File, semanticquery.CallArgument(starts[0].Args, 0, "string"), semanticquery.CallArgument(starts[1].Args, 0, "string")) {
		return
	}
	ctx.ReportNode(call, message)
}
func (rule) Semantic() {}
