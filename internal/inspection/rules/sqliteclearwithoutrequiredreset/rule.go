// Package sqliteclearwithoutrequiredreset implements SqliteClearWithoutRequiredReset.
package sqliteclearwithoutrequiredreset

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Reset the SQLite statement before rebinding cleared parameters."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteClearWithoutRequiredReset" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP >= phpversion.PHP72 {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "execute") {
		return
	}
	executed, cleared, bound := false, false, false
	var result *syntax.MethodCall
	for _, node := range semanticquery.NativePriorCalls(ctx, c, c.Var, "execute", "clear", "bindValue", "bindParam", "reset") {
		if p, ok := node.(*syntax.MethodCall); ok {
			if semanticquery.NativeMethod(ctx, p, "SQLite3Stmt", "reset") {
				executed, cleared, bound = false, false, false
			} else if semanticquery.NativeMethod(ctx, p, "SQLite3Stmt", "execute") {
				executed = true
				result = p
				cleared, bound = false, false
			} else if semanticquery.NativeMethod(ctx, p, "SQLite3Stmt", "clear") {
				cleared = executed
				bound = false
			} else if cleared && (semanticquery.NativeMethod(ctx, p, "SQLite3Stmt", "bindValue") || semanticquery.NativeMethod(ctx, p, "SQLite3Stmt", "bindParam")) {
				bound = true
			}
		}
	}
	if bound && busyResult(ctx, c, result) {
		ctx.ReportNode(c, message)
	}
}

// A successful row fetch proves that the old statement is still active;
// execute alone resets its cursor even on older PHP versions.
func busyResult(ctx *analysis.Context, at, producer *syntax.MethodCall) bool {
	for p := at.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(at.Span()) {
			continue
		}
		read, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
		if ok && semanticquery.NativeMethod(ctx, read, "SQLite3Result", "fetchArray") && semanticquery.NativeValue(ctx, read.Var) == producer {
			return true
		}
	}
	return false
}
