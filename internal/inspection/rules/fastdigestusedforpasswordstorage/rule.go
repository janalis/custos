// Package fastdigestusedforpasswordstorage implements the native FastDigestUsedForPasswordStorage inspection.
package fastdigestusedforpasswordstorage

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Store passwords with a password hashing API."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FastDigestUsedForPasswordStorage" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFastDigestUsedForPasswordStorage(ctx, n, message)
}
