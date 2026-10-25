// Package pregmatchfailureasnonmatch implements the native PregMatchFailureAsNonMatch inspection.
package pregmatchfailureasnonmatch

import (
	"strings"

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
	if alwaysMatches(ctx, semanticquery.CallArgument(call.Args, 0, "pattern")) { // E3
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

// alwaysMatches reports a literal pattern whose body is empty or `^`
// (`//u`, `{^}u`): it matches every subject, so a falsy result can only be
// the failure, which is what the UTF-8 validity check `!preg_match('//u',
// $s)` tests.
func alwaysMatches(ctx *analysis.Context, pattern syntax.Expr) bool {
	v, ok := semanticquery.NativeString(ctx, pattern)
	if !ok || len(v) < 2 {
		return false
	}
	closing := map[byte]byte{'(': ')', '{': '}', '[': ']', '<': '>'}[v[0]]
	if closing == 0 {
		closing = v[0]
	}
	end := strings.LastIndexByte(v, closing)
	if end < 1 {
		return false
	}
	body := v[1:end]
	return body == "" || body == "^"
}
