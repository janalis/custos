// Package xpathqueryfailuredereferenced implements the native
// XPathQueryFailureDereferenced inspection.
package xpathqueryfailuredereferenced

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
)

const message = "Check the XPath query result before accessing its nodes."

type rule struct{}

// New returns the XPathQueryFailureDereferenced inspection.
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
	if ok && semanticquery.NativeMethod(ctx, q, "DOMXPath", "query") && !ctx.Flow().Excludes(receiver, "false") && (semanticquery.NativeLabeled(ctx, receiver) || !semanticquery.NativeNarrowedExcludes(ctx, receiver, types.Of(`\DOMNodeList`, "false"), "false")) {
		ctx.ReportNode(n, message)
	}
}
