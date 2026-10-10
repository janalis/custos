// Package collatorcomparisontruthinessreversed implements the native CollatorComparisonTruthinessReversed inspection.
package collatorcomparisontruthinessreversed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare collation results explicitly."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CollatorComparisonTruthinessReversed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "Collator", "compare") {
		return
	}
	if semanticquery.ExpansionCTruthy(c) {
		ctx.ReportNode(c, message)
		return
	}
	p, _ := astquery.ParentSkipParens(c)
	if b, ok := p.(*syntax.Binary); ok && (b.Op.Kind == syntax.TIsEqual || b.Op.Kind == syntax.TIsIdentical) {
		if truth, known := astquery.BoolConst(b.Right); known && truth {
			ctx.ReportNode(b, message)
		}
	}
}
