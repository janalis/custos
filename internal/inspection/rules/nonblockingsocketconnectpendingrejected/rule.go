// Package nonblockingsocketconnectpendingrejected implements NonblockingSocketConnectPendingRejected.
package nonblockingsocketconnectpendingrejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Handle a pending nonblocking connection."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonblockingSocketConnectPendingRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckNonblockingSocketConnectPendingRejected(ctx, n, message)
}
