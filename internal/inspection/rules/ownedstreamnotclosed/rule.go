// Package ownedstreamnotclosed implements the native OwnedStreamNotClosed inspection.
package ownedstreamnotclosed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Close the owned stream before leaving its scope."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OwnedStreamNotClosed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckOwnedStreamNotClosed(ctx, n, message)
}
