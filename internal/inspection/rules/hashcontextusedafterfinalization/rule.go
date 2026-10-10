// Package hashcontextusedafterfinalization implements HashContextUsedAfterFinalization.
package hashcontextusedafterfinalization

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Create a new hash context after finalization."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HashContextUsedAfterFinalization" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckHashContextUsedAfterFinalization(ctx, n, message)
}
