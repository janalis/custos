// Package sqlitebindingindexzero implements SqliteBindingIndexZero.
package sqlitebindingindexzero

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Start SQLite binding positions at one."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteBindingIndexZero" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "bindValue") && !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "bindParam") {
		return
	}
	param := semanticquery.CallArgument(c.Args, 0, "param")
	v, known := semanticquery.NativeInt(ctx, param)
	if !known || v != 0 {
		return
	}
	if variable, ok := syntax.UnwrapParens(param).(*syntax.Variable); ok && flowquery.SteppedBefore(ctx.File, variable) {
		return
	}
	ctx.ReportNode(c, message)
}
