// Package arraychunkinvalidsize implements the native ArrayChunkInvalidSize inspection.
package arraychunkinvalidsize

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a positive chunk size."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayChunkInvalidSize" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_chunk") {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 1, "length")
	size, known := semanticquery.NativeInt(ctx, arg)
	if known && size < 1 && !rewrittenOutOfBand(ctx.File, call, arg) {
		ctx.ReportNode(call, message)
	}
}

// rewrittenOutOfBand reports whether a variable of the size expression, or a
// variable whose value an assignment copies into one, can change without a
// plain assignment in the scope of the call: through `++` or `--`, through a
// closure or array item holding a reference to it, or through extract().
// The resolved literal is then not a proven size.
func rewrittenOutOfBand(f *syntax.File, call *syntax.FuncCall, size syntax.Expr) bool {
	scope := syntax.EnclosingFuncLike(call)
	seen := map[string]bool{}
	var queue []string
	collect := func(x syntax.Node) bool {
		if v, ok := x.(*syntax.Variable); ok && !seen[v.Name] {
			seen[v.Name] = true
			queue = append(queue, v.Name)
		}
		return true
	}
	syntax.Inspect(size, collect)
	if len(queue) == 0 {
		return false
	}
	for ; len(queue) > 0; queue = queue[1:] {
		for _, access := range flowquery.VarAccesses(f, scope, queue[0]) {
			switch by := access.By.(type) {
			case *syntax.IncDec:
				return true
			case *syntax.ClosureUse:
				if by.ByRef {
					return true
				}
			case *syntax.Assign:
				if access.Write {
					syntax.Inspect(by.Value, collect)
				}
			}
			if item, ok := access.Var.Parent().(*syntax.ArrayItem); ok && item.ByRef {
				return true
			}
		}
	}
	return astquery.CallsExtract(f, scope)
}
