// Package headersaftercommittedoutput implements the native HeadersAfterCommittedOutput inspection.
package headersaftercommittedoutput

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Set response headers before committing output."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "HeadersAfterCommittedOutput" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "header" && name != "setcookie" && name != "http_response_code" {
		return
	}
	if name == "http_response_code" && !headersStatusWrite(ctx, c) {
		return
	}
	commit, committed := ctx.Flow().GlobalStateBefore(c, "committed-output")
	if committed && !headersUnsentProven(ctx, c, commit.Span) {
		ctx.ReportNode(c, message)
	}
}

// headersStatusWrite reports whether http_response_code is called as a setter.
// Without a status argument, or with one known to be falsy, it only reads.
func headersStatusWrite(ctx *analysis.Context, c *syntax.FuncCall) bool {
	code := semanticquery.CallArgument(c.Args, 0, "response_code")
	if code == nil {
		return false
	}
	set, known := semanticquery.NativeTruth(ctx, code)
	return set || !known
}

// headersUnsentProven proves that a headers_sent() call evaluated before the
// call returned false on its path: the conditions guarding the call narrow
// its result to false. A commit recorded between that test and the call
// invalidates it.
func headersUnsentProven(ctx *analysis.Context, c *syntax.FuncCall, commit syntax.Span) bool {
	return semanticquery.NativeEarlierCallProves(ctx, c, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		test, ok := x.(*syntax.FuncCall)
		if !ok || commit.Start >= test.Span().End && commit.End <= c.Span().Start || !semanticquery.NativeBuiltin(ctx, test, "headers_sent") {
			return types.Unknown, nil
		}
		return types.Bool, semanticquery.OnlyFalse
	})
}
