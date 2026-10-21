// Package sqlitenondeterministicfunctiondeclareddeterministic implements SqliteNondeterministicFunctionDeclaredDeterministic.
package sqlitenondeterministicfunctiondeclareddeterministic

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Remove the deterministic flag from this callback."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteNondeterministicFunctionDeclaredDeterministic" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if ctx.PHP < phpversion.PHP71 || !semanticquery.NativeMethod(ctx, c, "SQLite3", "createFunction") {
		return
	}
	flag, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 3, "flags"))
	if !known || flag&2048 == 0 {
		return
	}
	_, body := semanticquery.NativeCallback(ctx, semanticquery.CallArgument(c.Args, 1, "callback"))
	returns, complete := semanticquery.NativeReturns(body)
	if !complete {
		return
	}
	for _, ret := range returns {
		found := false
		if p, ok := syntax.UnwrapParens(ret).(*syntax.FuncCall); ok {
			switch semanticquery.NativeBuiltinName(ctx, p) {
			case "random_int":
				found = !semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 0, "min"), semanticquery.CallArgument(p.Args, 1, "max"))
			case "random_bytes", "time", "microtime":
				found = true
			}
		}
		if found {
			ctx.ReportNode(c, message)
			return
		}
	}
}
