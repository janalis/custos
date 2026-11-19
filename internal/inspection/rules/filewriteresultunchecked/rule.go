// Package filewriteresultunchecked implements the native FileWriteResultUnchecked inspection.
package filewriteresultunchecked

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Verify the write before reporting success."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileWriteResultUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, call, "file_put_contents") && semanticquery.NativeSuccessFollower(ctx, call) {
		ctx.ReportNode(call, message)
	}
}
