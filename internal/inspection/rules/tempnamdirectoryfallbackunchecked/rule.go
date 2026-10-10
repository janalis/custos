// Package tempnamdirectoryfallbackunchecked implements the native TempnamDirectoryFallbackUnchecked inspection.
package tempnamdirectoryfallbackunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Verify the temporary file remains in the required directory."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TempnamDirectoryFallbackUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckTempnamDirectoryFallbackUnchecked(ctx, n, message)
}
