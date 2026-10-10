// Package finallythrowmasksexception implements the native FinallyThrowMasksException inspection.
package finallythrowmasksexception

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Preserve the original exception during cleanup."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FinallyThrowMasksException" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTry} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP55 {
		return
	}
	tr := n.(*syntax.Try)
	if tr.Finally == nil || tr.Body == nil || len(tr.Catches) != 0 {
		return
	}
	if explicitThrow(tr.Body) == nil {
		return
	}
	if throw := explicitThrow(tr.Finally.Body); throw != nil {
		ctx.ReportNode(throw, message)
	}
}

func explicitThrow(b *syntax.Block) *syntax.Throw {
	if b == nil || len(b.Stmts) == 0 {
		return nil
	}
	s, ok := b.Stmts[len(b.Stmts)-1].(*syntax.ExprStmt)
	if !ok {
		return nil
	}
	t, _ := s.Expr.(*syntax.Throw)
	return t
}
