// Package fibersuspendoutsidefiber implements FiberSuspendOutsideFiber.
package fibersuspendoutsidefiber

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Suspend only inside a running fiber."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberSuspendOutsideFiber" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KStaticCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFiberSuspendOutsideFiber(ctx, n, message)
}
