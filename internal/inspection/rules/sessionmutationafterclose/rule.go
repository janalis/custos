// Package sessionmutationafterclose implements the native SessionMutationAfterClose inspection.
package sessionmutationafterclose

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Persist session changes before closing the session."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionMutationAfterClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign, syntax.KIncDec} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSessionMutationAfterClose(ctx, n, message)
}
