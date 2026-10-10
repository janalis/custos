// Package readmodifywritelocktoolate implements the native ReadModifyWriteLockTooLate inspection.
package readmodifywritelocktoolate

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Hold the lock across the complete read-modify-write operation."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadModifyWriteLockTooLate" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckReadModifyWriteLockTooLate(ctx, n, message)
}
