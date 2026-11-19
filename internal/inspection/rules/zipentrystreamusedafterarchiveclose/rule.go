// Package zipentrystreamusedafterarchiveclose implements the native ZipEntryStreamUsedAfterArchiveClose inspection.
package zipentrystreamusedafterarchiveclose

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Read the entry stream before closing its archive."

type rule struct{}

// New returns the ZipEntryStreamUsedAfterArchiveClose inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipEntryStreamUsedAfterArchiveClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "stream_get_contents", "fread", "fgets":
	default:
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "stream")
	variable, ok := input.(*syntax.Variable)
	if !ok {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	previous, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return
	}
	close, ok := previous.Expr.(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, close, "ZipArchive", "close") {
		return
	}
	earlier, ok := prior[len(prior)-2].(*syntax.ExprStmt)
	if !ok {
		return
	}
	assignment, ok := earlier.Expr.(*syntax.Assign)
	if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
		return
	}
	target, ok := assignment.Var.(*syntax.Variable)
	if !ok || target.Name != variable.Name {
		return
	}
	origin, ok := assignment.Value.(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, origin, "ZipArchive", "getStream") {
		return
	}
	if astquery.Equivalent(ctx.File, origin.Var, close.Var) {
		ctx.ReportNode(c, message)
	}
}
