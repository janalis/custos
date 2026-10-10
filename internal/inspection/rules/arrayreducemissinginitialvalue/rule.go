// Package arrayreducemissinginitialvalue implements the native ArrayReduceMissingInitialValue inspection.
package arrayreducemissinginitialvalue

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply an array as the initial accumulator."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayReduceMissingInitialValue" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_reduce") || semanticquery.CallArgument(call.Args, 2, "initial") != nil {
		return
	}
	callback := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 1, "callback"))
	params, body := astquery.ScopeParts(callback)
	if len(params) < 1 || params[0].Var == nil {
		return
	}
	var result syntax.Expr
	switch b := body.(type) {
	case *syntax.Block:
		if len(b.Stmts) != 1 {
			return
		}
		r, ok := b.Stmts[0].(*syntax.Return)
		if !ok {
			return
		}
		result = r.Expr
	case syntax.Expr:
		result = b
	}
	inner, ok := syntax.UnwrapParens(result).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, inner, "array_merge") {
		return
	}
	first := semanticquery.CallArgument(inner.Args, 0, "arrays")
	v, ok := first.(*syntax.Variable)
	if !ok || v.Name != params[0].Var.Name {
		return
	}
	args := astquery.WrittenArguments(call.Args)
	if len(args) < 2 {
		return
	}
	end := args[len(args)-1].Span().End
	text := ", []"
	for _, arg := range args {
		if arg.Name != nil {
			text = ", initial: []"
		}
	}
	ctx.ReportNode(call, message, diagnostic.Fix{Title: "Initialize the accumulator as an array", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: syntax.Span{Start: end, End: end}, NewText: text}}
	}})
}
