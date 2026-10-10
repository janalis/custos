// Package arrayrandkeyusedasvalue implements the native ArrayRandKeyUsedAsValue inspection.
package arrayrandkeyusedasvalue

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the array element at the returned random key."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayRandKeyUsedAsValue" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_rand") {
		return
	}
	count := semanticquery.CallArgument(call.Args, 1, "num")
	if count != nil {
		n, known := semanticquery.NativeInt(ctx, count)
		if !known || n != 1 {
			return
		}
	}
	array := semanticquery.NativeArray(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if array == nil || len(array.Items) == 0 {
		return
	}
	values := map[string]bool{}
	for _, item := range array.Items {
		if item.Key != nil {
			return
		}
		s, known := semanticquery.NativeString(ctx, item.Value)
		if !known {
			return
		}
		values[s] = true
	}
	arg, ok := call.Parent().(*syntax.Arg)
	if !ok || arg.Unpack {
		return
	}
	list := arg.Parent().(*syntax.ArgList)
	consumer, ok := list.Parent().(*syntax.FuncCall)
	if !ok {
		return
	}
	body := semanticquery.NativeFunctionBody(ctx, consumer)
	if body == nil || len(body.Stmts) != 1 {
		return
	}
	sw, ok := body.Stmts[0].(*syntax.Switch)
	if !ok {
		return
	}
	parameter, ok := sw.Cond.(*syntax.Variable)
	if !ok {
		return
	}
	fn := ctx.Types().ResolveFunction(consumer)
	position := -1
	for i, p := range fn.Params {
		if p.Name == parameter.Name {
			position = i
		}
	}
	if position < 0 || semanticquery.CallArgument(consumer.Args, position, parameter.Name) != call {
		return
	}
	seen := map[string]bool{}
	for _, c := range sw.Cases {
		if c.Cond == nil {
			return
		}
		s, known := semanticquery.NativeString(ctx, c.Cond)
		if !known || !values[s] {
			return
		}
		seen[s] = true
	}
	if len(seen) == len(values) {
		ctx.ReportNode(call, message)
	}
}
