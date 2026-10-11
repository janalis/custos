// Package sodiumsecretboxkeylengthmismatch implements SodiumSecretboxKeyLengthMismatch.
package sodiumsecretboxkeylengthmismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a correctly sized secretbox key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumSecretboxKeyLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "sodium_crypto_secretbox", "sodium_crypto_secretbox_open":
	default:
		return
	}
	if key, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 2, "key")).(*syntax.Variable); ok && indirectlyWritten(ctx, key, c) {
		return
	}
	semanticquery.CheckSodiumSecretboxKeyLengthMismatch(ctx, n, message)
}

// indirectlyWritten reports whether the key variable may have been replaced
// before the call by something other than a plain assignment: a reference
// alias, a destructuring or foreach target, a by-reference closure import, a
// `global` import, a call able to write its argument, extract() or a variable
// variable. The shared local-value lookup only follows plain assignments.
func indirectlyWritten(ctx *analysis.Context, key *syntax.Variable, call *syntax.FuncCall) bool {
	scope := syntax.EnclosingVariableScope(call)
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= call.Span().Start {
			return false
		}
		if n != scope && syntax.IsVariableScope(n) {
			if closure, ok := n.(*syntax.Closure); ok {
				for _, use := range closure.Uses {
					found = found || (use.ByRef && use.Var.Name == key.Name)
				}
			}
			return false
		}
		switch x := n.(type) {
		case *syntax.FuncCall:
			found = semanticquery.NativeBuiltin(ctx, x, "extract")
		case *syntax.Variable:
			found = x.NameExpr != nil || (x.Name == key.Name && writtenThrough(ctx, x))
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// writtenThrough inspects one occurrence of the key variable.
func writtenThrough(ctx *analysis.Context, v *syntax.Variable) bool {
	switch p := v.Parent().(type) {
	case *syntax.Assign:
		return p.ByRef && p.Value == syntax.Expr(v)
	case *syntax.Global:
		return true
	case *syntax.Foreach:
		return p.Expr != syntax.Expr(v)
	case *syntax.Arg:
		return p.ByRef || !readOnlyCall(ctx, p)
	case *syntax.ArrayItem:
		return p.ByRef || (p.Value == syntax.Expr(v) && destructured(p))
	}
	return false
}

// readOnlyCall accepts only an available builtin function without any
// by-reference parameter as a call that cannot replace its argument.
func readOnlyCall(ctx *analysis.Context, arg *syntax.Arg) bool {
	call, ok := arg.Parent().Parent().(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) == "" {
		return false
	}
	for _, param := range ctx.Types().ResolveFunction(call).Params {
		if param.ByRef {
			return false
		}
	}
	return true
}

// destructured reports whether the array item belongs to a list()/[...]
// pattern that is the target of an assignment or a foreach.
func destructured(item *syntax.ArrayItem) bool {
	var child syntax.Node = item
	for {
		switch p := child.Parent().(type) {
		case *syntax.Array, *syntax.List, *syntax.ArrayItem:
			child = p
		case *syntax.Assign:
			return syntax.Node(p.Var) == child
		case *syntax.Foreach:
			return syntax.Node(p.Expr) != child
		default:
			return false
		}
	}
}
