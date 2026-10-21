// Package curlheadercallbackmissingbytecount implements CurlHeaderCallbackMissingByteCount.
package curlheadercallbackmissingbytecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the consumed header byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderCallbackMissingByteCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	value := semanticquery.CurlSetting(ctx, c, "CURLOPT_HEADERFUNCTION")
	if value == nil {
		return
	}
	_, body := semanticquery.NativeCallback(ctx, value)
	block, ok := body.(*syntax.Block)
	if !ok || semanticquery.NativeGeneratorBody(block) {
		return
	}
	missing := !callbackTerminates(ctx, block)
	syntax.Inspect(block, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if r, ok := n.(*syntax.Return); ok && r.Expr == nil && semanticquery.NativeCallbackReachable(ctx, r) {
			missing = true
		}
		return true
	})
	if missing {
		ctx.ReportNode(c, message)
	}
}

// callbackTerminates proves termination through literal callback
// conditions as well as the syntax-level terminal statements. Depth is bounded.
func callbackTerminates(ctx *analysis.Context, body syntax.Stmt) bool {
	return expansionDCallbackTerminates(ctx, body, 64)
}

func expansionDCallbackTerminates(ctx *analysis.Context, body syntax.Stmt, budget int) bool {
	if budget == 0 {
		return false
	}
	if syntax.Terminates(body) {
		return true
	}
	switch b := body.(type) {
	case *syntax.Block:
		for _, stmt := range b.Stmts {
			if expansionDCallbackTerminates(ctx, stmt, budget-1) {
				return true
			}
		}
	case *syntax.If:
		truth, known := semanticquery.NativeTruth(ctx, b.Cond)
		if known && len(b.ElseIfs) == 0 {
			if truth {
				return expansionDCallbackTerminates(ctx, b.Body, budget-1)
			}
			if b.Else != nil {
				return expansionDCallbackTerminates(ctx, b.Else.Body, budget-1)
			}
		}
	}
	return false
}
