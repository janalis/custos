// Package sqliteresultusedafterfinalize implements SqliteResultUsedAfterFinalize.
package sqliteresultusedafterfinalize

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the SQLite result before finalizing it."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteResultUsedAfterFinalize" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	use := false
	for _, name := range []string{"fetchArray", "reset", "columnName", "columnType", "numColumns"} {
		use = use || semanticquery.NativeMethod(ctx, c, "SQLite3Result", name)
	}
	if !use || rebindable(c) {
		return
	}
	// finalize() cannot fail, so a call that always runs before the use is
	// enough: its result does not have to be tested.
	for _, prior := range semanticquery.NativePriorCalls(ctx, c, c.Var, "finalize") {
		finish, ok := prior.(*syntax.MethodCall)
		if ok && semanticquery.NativeMethod(ctx, finish, "SQLite3Result", "finalize") {
			ctx.ReportNode(c, message)
			return
		}
	}
	for p := c.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		finish, ok := guard(branch.Cond).(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, finish, "SQLite3Result", "finalize") {
			continue
		}
		a := ctx.Flow().Value(c.Var)
		b := ctx.Flow().Value(finish.Var)
		if a.Complete && b.Complete && a.Identity != 0 && a.Identity == b.Identity && a.Invalidated == b.Invalidated {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// rebindable reports whether the receiver is a by-reference parameter: any
// call can then replace the result behind the variable.
func rebindable(c *syntax.MethodCall) bool {
	v, ok := syntax.UnwrapParens(c.Var).(*syntax.Variable)
	if !ok {
		return false
	}
	for _, p := range syntax.VariableScopeParams(syntax.EnclosingVariableScope(c)) {
		if p.ByRef && p.Var.Name == v.Name {
			return true
		}
	}
	return false
}

// guard returns the expression whose truth an if condition tests: the
// condition itself, or the other operand of a comparison with `true`.
func guard(cond syntax.Expr) syntax.Expr {
	cond = syntax.UnwrapParens(cond)
	b, ok := cond.(*syntax.Binary)
	if !ok || (b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsEqual) {
		return cond
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if k, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch); ok && k.Name != nil && strings.EqualFold(strings.TrimPrefix(k.Name.Value, "\\"), "true") {
			return syntax.UnwrapParens(pair[0])
		}
	}
	return cond
}
