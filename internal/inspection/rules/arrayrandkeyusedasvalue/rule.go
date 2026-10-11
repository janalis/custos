// Package arrayrandkeyusedasvalue implements the native ArrayRandKeyUsedAsValue inspection.
package arrayrandkeyusedasvalue

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
	if len(seen) != len(values) || keysSatisfyCases(len(array.Items), seen, ctx.PHP.AtLeast(phpversion.PHP80)) {
		return
	}
	ctx.ReportNode(call, message)
}

// keysSatisfyCases reports whether each list key 0..n-1 loosely equals one of
// the case literals, so the switch already works with the returned key.
func keysSatisfyCases(n int, cases map[string]bool, php8 bool) bool {
	for key := range n {
		matched := false
		for literal := range cases {
			if value, ok := comparedNumber(literal, php8); ok && value == float64(key) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// comparedNumber returns the number PHP compares an integer against for the
// string s. A numeric string always compares by value; any other string
// compares as text from PHP 8.0 (never equal to an integer) and by its leading
// numeric prefix, zero when there is none, before that.
func comparedNumber(s string, php8 bool) (float64, bool) {
	if !astquery.IsComparisonNumericString(s) {
		if php8 {
			return 0, false
		}
		end := len(s)
		for end > 0 && !astquery.IsComparisonNumericString(s[:end]) {
			end--
		}
		if end == 0 {
			return 0, true
		}
		s = s[:end]
	}
	value, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return value, true
}
