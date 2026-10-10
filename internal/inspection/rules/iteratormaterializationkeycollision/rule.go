// Package iteratormaterializationkeycollision implements the native IteratorMaterializationKeyCollision inspection.
package iteratormaterializationkeycollision

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Preserve all yielded values when materializing the iterator."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IteratorMaterializationKeyCollision" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "iterator_to_array") || ctx.PHP < phpversion.PHP55 {
		return
	}
	if p := semanticquery.CallArgument(call.Args, 1, "preserve_keys"); p != nil {
		truth, known := semanticquery.NativeTruth(ctx, p)
		if !known || !truth {
			return
		}
	}
	producer, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 0, "iterator")).(*syntax.FuncCall)
	if !ok {
		return
	}
	body := semanticquery.NativeFunctionBody(ctx, producer)
	if body == nil {
		return
	}
	seen := map[string]bool{}
	for _, st := range body.Stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			return
		}
		y, ok := es.Expr.(*syntax.Yield)
		if !ok {
			return
		}
		key, known := semanticquery.NativeArrayKey(ctx, y.Key)
		if !known {
			return
		}
		if seen[key] {
			ctx.ReportNode(call, message)
			return
		}
		seen[key] = true
	}
}
