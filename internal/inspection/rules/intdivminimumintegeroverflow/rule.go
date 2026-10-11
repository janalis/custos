// Package intdivminimumintegeroverflow implements the native IntdivMinimumIntegerOverflow inspection.
package intdivminimumintegeroverflow

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

const message = "Avoid dividing PHP_INT_MIN by negative one."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IntdivMinimumIntegerOverflow" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "intdiv") {
		return
	}
	dividend := semanticquery.CallArgument(c.Args, 0, "num1")
	divisor := semanticquery.CallArgument(c.Args, 1, "num2")
	v, k := semanticquery.NativeInt(ctx, divisor)
	constant, ok := semanticquery.NativeLocalValue(ctx, dividend).(*syntax.ConstFetch)
	if !k || v != -1 || !ok || semanticquery.GlobalConstName(ctx, constant) != "PHP_INT_MIN" {
		return
	}
	looped := inLoop(c)
	if unchanged(ctx, c, looped, dividend, map[string]bool{}) && unchanged(ctx, c, looped, divisor, map[string]bool{}) {
		ctx.ReportNode(c, message)
	}
}

// inLoop reports whether n sits in a loop of its own scope, where code that
// follows n in the source may still run before it.
func inLoop(n syntax.Node) bool {
	for p := n.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			return true
		}
	}
	return false
}

// unchanged reports whether the variable operand e, and every variable its
// value is copied from, is only ever bound by plain assignments before call
// and never handed out by reference. The shared value resolvers follow plain
// assignments only, so any other write (`++`, compound or destructuring
// assignment, foreach target, unset, global, static, catch) or reference
// leaves the discovered value unproven. Non-variable operands are stable.
func unchanged(ctx *analysis.Context, call *syntax.FuncCall, looped bool, e syntax.Expr, seen map[string]bool) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	if !ok || seen[v.Name] {
		return true
	}
	seen[v.Name] = true
	for _, acc := range flowquery.VarAccesses(ctx.File, syntax.EnclosingVariableScope(v), v.Name) {
		if !looped && acc.Var.Span().Start >= call.Span().End {
			break
		}
		if !acc.Write {
			if referenced(ctx, acc) {
				return false
			}
			continue
		}
		a, ok := acc.By.(*syntax.Assign)
		if !ok || a.Var != syntax.Expr(acc.Var) || a.ByRef || a.Op.Kind != syntax.TEqual || !unchanged(ctx, call, looped, a.Value, seen) {
			return false
		}
	}
	return true
}

// referenced reports whether a read of a variable may hand out a reference
// to it: `$r = &$v`, `[&$v]`, `use (&$v)`, a call-time `&$v`, or an argument
// of a call that is unresolved or binds it to a by-reference parameter.
func referenced(ctx *analysis.Context, acc flowquery.VarAccess) bool {
	if u, ok := acc.By.(*syntax.ClosureUse); ok {
		return u.ByRef
	}
	switch p := acc.Var.Parent().(type) {
	case *syntax.Assign:
		return p.ByRef
	case *syntax.ArrayItem:
		return p.ByRef
	case *syntax.Arg:
		list := p.Parent().(*syntax.ArgList)
		callee, resolved := semanticquery.ResolveCallee(ctx, list.Parent())
		if p.ByRef || p.Unpack || !resolved {
			return true
		}
		params := callee.Params
		if p.Name != nil {
			return slices.ContainsFunc(params, func(param index.Param) bool {
				return param.ByRef && strings.EqualFold(strings.TrimPrefix(param.Name, "$"), p.Name.Value)
			})
		}
		i := slices.Index(list.Args, syntax.Expr(p))
		if n := len(params); i >= n && n > 0 && params[n-1].Variadic {
			i = n - 1
		}
		return i < len(params) && params[i].ByRef
	}
	return false
}
