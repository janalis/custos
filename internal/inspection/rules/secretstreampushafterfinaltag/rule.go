// Package secretstreampushafterfinaltag implements SecretstreamPushAfterFinalTag.
package secretstreampushafterfinaltag

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Start a new stream after the final tag."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SecretstreamPushAfterFinalTag" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSecretstreamPushAfterFinalTag(ctx, n, message)
}
