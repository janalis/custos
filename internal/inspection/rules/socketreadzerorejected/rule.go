// Package socketreadzerorejected implements SocketReadZeroRejected.
package socketreadzerorejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare socket read failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SocketReadZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSocketReadZeroRejected(ctx, n, message)
}
