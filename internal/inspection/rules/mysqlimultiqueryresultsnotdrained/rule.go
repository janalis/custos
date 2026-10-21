// Package mysqlimultiqueryresultsnotdrained implements MysqliMultiQueryResultsNotDrained.
package mysqlimultiqueryresultsnotdrained

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Drain all multi-query results before another query."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliMultiQueryResultsNotDrained" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli", "query") {
		return
	}
	prior := semanticquery.EarlierMethodCalls(ctx, c, c.Var, "mysqli", "multi_query")
	if len(prior) == 0 {
		return
	}
	p := prior[len(prior)-1]
	if !semanticquery.NativeCallSucceeded(ctx, p, c) {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(p.Args, 0, "query"))
	if !known {
		return
	}
	tokens, known := semanticquery.SQLTokens(sql)
	if !known {
		return
	}
	count := 0
	hasTokens := false
	for _, token := range tokens {
		if token == ";" {
			if hasTokens {
				count++
				hasTokens = false
			}
		} else {
			hasTokens = true
		}
	}
	if hasTokens {
		count++
	}
	if count < 2 {
		return
	}
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		check, ok := event.(*syntax.MethodCall)
		if !ok || check.Span().Start < p.Span().End || check.Span().End >= c.Span().Start || (!semanticquery.NativeMethod(ctx, check, "mysqli", "more_results") && !semanticquery.NativeMethod(ctx, check, "mysqli", "next_result")) || !semanticquery.SameStableLocal(ctx, check.Var, c.Var) {
			continue
		}
		// The probe must decide the loop: alone or as an operand of `&&`.
		cond := syntax.Node(check)
		for {
			parent := cond.Parent()
			if b, ok := parent.(*syntax.Binary); ok && b.Op.Kind != syntax.TBooleanAnd {
				break
			}
			if _, ok := parent.(*syntax.Binary); !ok {
				if _, ok := parent.(*syntax.Paren); !ok {
					break
				}
			}
			cond = parent
		}
		var loop syntax.Node
		switch l := cond.Parent().(type) {
		case *syntax.While:
			if l.Cond == cond {
				loop = l
			}
		case *syntax.DoWhile:
			if l.Cond == cond {
				loop = l
			}
		}
		// A query inside the loop runs before the drain completes.
		if loop == nil || loop.Span().End > c.Span().Start {
			continue
		}
		drained := false
		syntax.Inspect(loop, func(n syntax.Node) bool {
			if next, ok := n.(*syntax.MethodCall); ok && semanticquery.NativeMethod(ctx, next, "mysqli", "next_result") && semanticquery.SameStableLocal(ctx, next.Var, c.Var) {
				drained = true
			}
			return true
		})
		if drained {
			return
		}
	}
	ctx.ReportNode(c, message)
}
