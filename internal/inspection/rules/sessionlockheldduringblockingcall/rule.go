// Package sessionlockheldduringblockingcall implements the native SessionLockHeldDuringBlockingCall inspection.
package sessionlockheldduringblockingcall

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Release the session lock before blocking when safe."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SessionLockHeldDuringBlockingCall" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckSessionLockHeldDuringBlockingCall(ctx, n, message)
}
