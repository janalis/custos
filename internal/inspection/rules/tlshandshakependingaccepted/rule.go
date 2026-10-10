// Package tlshandshakependingaccepted implements TlsHandshakePendingAccepted.
package tlshandshakependingaccepted

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Require a completed TLS handshake before writing."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TlsHandshakePendingAccepted" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckTLSHandshakePendingAccepted(ctx, n, message)
}
