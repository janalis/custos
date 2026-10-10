// Package hashfilefailureusedasdigest implements HashFileFailureUsedAsDigest.
package hashfilefailureusedasdigest

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject file hashing failure before using the digest."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HashFileFailureUsedAsDigest" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckHashFileFailureUsedAsDigest(ctx, n, message)
}
