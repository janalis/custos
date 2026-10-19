package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// NativePropertyUnchanged reports two reads of one property of one plain
// receiver, written in that order, that give the same value: reading the
// property runs no code, and nothing that runs between the reads may change
// it. Between them there is no call that may run code of the project, no
// write, unset or reference of a property of that name through any
// receiver, and no write of the receiver variable. A loop around the second
// read and not the first runs all of its body before it reads again.
func NativePropertyUnchanged(ctx *analysis.Context, first, second syntax.Expr) bool {
	a, ok := syntax.UnwrapParens(first).(*syntax.PropertyFetch)
	b, same := syntax.UnwrapParens(second).(*syntax.PropertyFetch)
	if !ok || !same || a.NullSafe || b.NullSafe || !astquery.Equivalent(ctx.File, a, b) || a.Span().End > b.Span().Start {
		return false
	}
	receiver, plain := syntax.UnwrapParens(a.Var).(*syntax.Variable)
	name, written := a.Name.(*syntax.Identifier)
	scope := syntax.EnclosingVariableScope(a)
	if !plain || receiver.NameExpr != nil || !written || syntax.EnclosingVariableScope(b) != scope || !nativePropertyPlainRead(ctx, a, name.Value) {
		return false
	}
	from, to := a.Span().End, b.Span().Start
	for p := b.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			if !p.Span().Contains(a.Span()) {
				to = max(to, p.Span().End)
			}
		}
	}
	for _, access := range flowquery.VarAccesses(ctx.File, scope, receiver.Name) {
		if at := access.Var.Span().Start; access.Write && at > from && at < to {
			return false
		}
	}
	unchanged := true
	visit := func(n syntax.Node) bool {
		if !unchanged || n.Span().End <= from || n.Span().Start >= to {
			return unchanged && n.Span().Start < to
		}
		if n != scope && syntax.IsVariableScope(n) {
			// A closure or a function declared here runs where it is called.
			return false
		}
		if n.Span().Contains(a.Span()) || n.Span().Contains(b.Span()) {
			// What holds a read runs before the first or after the second.
			return true
		}
		switch x := n.(type) {
		case *syntax.FuncCall:
			unchanged = NativeBuiltinName(ctx, x) != "" && ctx.Flow().Spares(x)
			// A PHP function may write what it is given by reference.
			if callee, resolved := ResolveCallee(ctx, x); unchanged && resolved {
				for _, reference := range ReferenceArguments(x.Args, callee.Params, true) {
					unchanged = unchanged && !nativePropertyNamed(reference.Argument.Value, name.Value)
				}
			}
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Eval:
			unchanged = false
		case *syntax.Assign:
			unchanged = !nativePropertyNamed(x.Var, name.Value) && !(x.ByRef && nativePropertyNamed(x.Value, name.Value))
		case *syntax.IncDec:
			unchanged = !nativePropertyNamed(x.Var, name.Value)
		case *syntax.Unset:
			for _, v := range x.Vars {
				unchanged = unchanged && !nativePropertyNamed(v, name.Value)
			}
		case *syntax.Foreach:
			unchanged = !nativePropertyNamed(x.Key, name.Value) && !nativePropertyNamed(x.Value, name.Value)
		case *syntax.Arg:
			unchanged = !x.ByRef || !nativePropertyNamed(x.Value, name.Value)
		}
		return unchanged
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return unchanged
}

// nativePropertyNamed reports an access to the property name, or to one
// named at run time, through any receiver, at any depth of the target.
func nativePropertyNamed(target syntax.Expr, name string) bool {
	found := false
	syntax.Inspect(target, func(n syntax.Node) bool {
		if fetch, ok := n.(*syntax.PropertyFetch); ok {
			written, literal := fetch.Name.(*syntax.Identifier)
			found = found || !literal || strings.EqualFold(written.Value, name)
		}
		return !found
	})
	return found
}

// nativePropertyPlainRead reports a property read that runs no code: every
// class the receiver may be of declares the property without a hook, has no
// `__get` method, and either is final or keeps the property private, so no
// class extending it can declare it again with one.
func nativePropertyPlainRead(ctx *analysis.Context, fetch *syntax.PropertyFetch, name string) bool {
	classes := ctx.TypeOf(fetch.Var).Classes()
	if len(classes) == 0 {
		return false
	}
	ix := ctx.Index()
	for _, class := range classes {
		class = strings.TrimPrefix(class, `\`)
		declared := ix.Class(class, ctx.PHP)
		property := ix.FindProperty(class, name, ctx.PHP)
		if declared == nil || property == nil || property.Magic || property.Static || property.Hooked || property.ReadsRunCode || ix.FindMethod(class, "__get", ctx.PHP) != nil {
			return false
		}
		if !declared.Final && property.Visibility != index.Private {
			return false
		}
	}
	return true
}
