// Package sqliteresultusedafterfinalize implements SqliteResultUsedAfterFinalize.
package sqliteresultusedafterfinalize

import (
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
	if !use {
		return
	}
	for p := c.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		finish, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
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
