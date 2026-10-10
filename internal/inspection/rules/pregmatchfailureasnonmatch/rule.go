// Package pregmatchfailureasnonmatch implements the native PregMatchFailureAsNonMatch inspection.
package pregmatchfailureasnonmatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish regex failure from a successful non-match."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PregMatchFailureAsNonMatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "preg_match", "preg_match_all":
	default:
		return
	}
	p, _ := astquery.ParentSkipParens(call)
	u, ok := p.(*syntax.Unary)
	if !ok || u.Op.Kind != syntax.TExclaim {
		return
	}
	parent, _ := astquery.ParentSkipParens(u)
	if _, ok := parent.(*syntax.If); !ok {
		return
	}
	ctx.ReportNode(call, message)
}
