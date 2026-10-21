// Package sodiumsecretboxnoncereused implements SodiumSecretboxNonceReused.
package sodiumsecretboxnoncereused

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const (
	message  = "Use a fresh nonce for each distinct message."
	function = "sodium_crypto_secretbox"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SodiumSecretboxNonceReused" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, function) || rebound(ctx, c) {
		return
	}
	check(ctx, n, message)
}

// rebound reports whether the nonce or key variable of c is written before c
// with no earlier encryption of another plaintext left after that write. The
// shared check keeps a flow identity across destructuring, loop variables,
// element writes and by-reference arguments, so those writes are found here.
func rebound(ctx *analysis.Context, c *syntax.FuncCall) bool {
	nonce := semanticquery.CallArgument(c.Args, 1, "nonce")
	key := semanticquery.CallArgument(c.Args, 2, "key")
	names := [2]string{localName(nonce), localName(key)}
	at := c.Span().Start
	var written uint32
	write := func(end uint32) {
		if end <= at && end > written {
			written = end
		}
	}
	mark := func(target syntax.Expr, end uint32) {
		if targets(target, names) {
			write(end)
		}
	}
	var priors []*syntax.FuncCall
	scope := syntax.EnclosingVariableScope(c)
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			mark(x.Var, x.Span().End)
			if x.ByRef {
				mark(x.Value, x.Span().End)
			}
		case *syntax.IncDec:
			mark(x.Var, x.Span().End)
		case *syntax.Foreach:
			for _, target := range []syntax.Expr{x.Key, x.Value} {
				if target != nil {
					mark(target, target.Span().End)
				}
			}
		case *syntax.FuncCall:
			switch semanticquery.NativeBuiltinName(ctx, x) {
			case function:
				priors = append(priors, x)
			case "extract", "parse_str":
				write(x.Span().End)
			}
		case *syntax.ArgList:
			for i, a := range x.Args {
				if arg, ok := a.(*syntax.Arg); ok && targets(arg.Value, names) && (arg.ByRef || bindsByRef(ctx, x, i)) {
					write(x.Span().End)
				}
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	if written == 0 {
		return false
	}
	plaintext, _ := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "message"))
	for _, p := range priors {
		if p.Span().Start < written || !semanticquery.NativeDominates(p, c) {
			continue
		}
		previous, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(p.Args, 0, "message"))
		if known && previous != plaintext &&
			astquery.Equivalent(ctx.File, semanticquery.CallArgument(p.Args, 1, "nonce"), nonce) &&
			astquery.Equivalent(ctx.File, semanticquery.CallArgument(p.Args, 2, "key"), key) {
			return false
		}
	}
	return true
}

func localName(e syntax.Expr) string {
	if v, ok := syntax.UnwrapParens(e).(*syntax.Variable); ok {
		return v.Name
	}
	return ""
}

// targets reports whether writing to e may rebind or alter one of names:
// the variable itself, one of its offsets, a destructuring slot holding
// either, or a variable variable.
func targets(e syntax.Expr, names [2]string) bool {
	var items []*syntax.ArrayItem
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Variable:
		return x.NameExpr != nil || x.Name == names[0] || x.Name == names[1]
	case *syntax.ArrayDimFetch:
		return targets(x.Var, names)
	case *syntax.Array:
		items = x.Items
	case *syntax.List:
		items = x.Items
	}
	for _, item := range items {
		if item != nil && targets(item.Value, names) {
			return true
		}
	}
	return false
}

// bindsByRef reports whether argument i of list binds to a by-reference
// parameter of a resolved callee (named, positional or variadic tail).
func bindsByRef(ctx *analysis.Context, list *syntax.ArgList, i int) bool {
	callee, _ := semanticquery.ResolveCallee(ctx, list.Parent())
	params := callee.Params
	if name := list.Args[i].(*syntax.Arg).Name; name != nil {
		for _, p := range params {
			if p.ByRef && strings.EqualFold(strings.TrimPrefix(p.Name, "$"), name.Value) {
				return true
			}
		}
		return false
	}
	if i >= len(params) {
		i = len(params) - 1
		if i < 0 || !params[i].Variadic {
			return false
		}
	}
	return params[i].ByRef
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	plaintext, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "message"))
	if !known {
		return
	}
	nonce := semanticquery.CallArgument(c.Args, 1, "nonce")
	key := semanticquery.CallArgument(c.Args, 2, "key")
	events := semanticquery.ScopeCallEvents(ctx, c)
	var latestMutation uint32
	for _, event := range events {
		m, ok := event.(*syntax.FuncCall)
		if !ok || m.Span().End >= c.Span().Start {
			continue
		}
		switch semanticquery.NativeBuiltinName(ctx, m) {
		case "sodium_increment", "sodium_add", "sodium_memzero":
			argument := "string"
			if semanticquery.NativeBuiltinName(ctx, m) == "sodium_add" {
				argument = "string1"
			}
			changed := semanticquery.CallArgument(m.Args, 0, argument)
			if m.Span().End > latestMutation && (astquery.Equivalent(ctx.File, nonce, changed) || astquery.Equivalent(ctx.File, key, changed)) {
				latestMutation = m.Span().End
			}
		}
	}
	for _, event := range events {
		p, ok := event.(*syntax.FuncCall)
		if !ok || p.Span().End < latestMutation || !semanticquery.NativeBuiltin(ctx, p, "sodium_crypto_secretbox") || !semanticquery.NativeDominates(p, c) {
			continue
		}
		previous, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(p.Args, 0, "message"))
		if known && previous != plaintext && semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 1, "nonce"), semanticquery.CallArgument(c.Args, 1, "nonce")) && semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(p.Args, 2, "key"), semanticquery.CallArgument(c.Args, 2, "key")) {
			ctx.ReportNode(c, message)
			return
		}
	}
}
