// Package curluploaddeclaredsizemismatch implements CurlUploadDeclaredSizeMismatch.
package curluploaddeclaredsizemismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the declared upload size to remaining bytes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlUploadDeclaredSizeMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlUploadDeclaredSizeMismatch(ctx, n, message)
}
