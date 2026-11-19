// Package renamefailureunchecked implements the native RenameFailureUnchecked inspection.
package renamefailureunchecked

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Check rename success before returning success."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RenameFailureUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "rename") && semanticquery.NativeSuccessFollower(ctx, c) {
		ctx.ReportNode(c, message)
	}
}
