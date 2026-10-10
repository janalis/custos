// Package xpathqueryfailuredereferenced implements XPathQueryFailureDereferenced.
package xpathqueryfailuredereferenced

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check the XPath query result before accessing its nodes."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "XPathQueryFailureDereferenced" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KMethodCall}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var receiver syntax.Expr
	switch c := n.(type) {
	case *syntax.PropertyFetch:
		receiver = c.Var
	case *syntax.MethodCall:
		receiver = c.Var
	}
	q, ok := semanticquery.NativeValue(ctx, receiver).(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, q, "DOMXPath", "query") && !semanticquery.NativeSentinelGuard(ctx, receiver, "false") {
		ctx.ReportNode(n, message)
	}
}
