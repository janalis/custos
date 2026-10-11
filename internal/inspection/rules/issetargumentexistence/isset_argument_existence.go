package issetargumentexistence

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// issetArgumentExistence reports isset/empty/?? checks on a variable whose
// first mention in the function is that very check.
type issetArgumentExistence struct{}

func (issetArgumentExistence) ID() string { return "IssetArgumentExistence" }
func (issetArgumentExistence) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary, syntax.KIsset, syntax.KEmpty}
}

var issetSpecialNames = map[string]bool{
	"this": true, "_GET": true, "_POST": true, "_SESSION": true, "_REQUEST": true,
	"_FILES": true, "_COOKIE": true, "_ENV": true, "_SERVER": true, "GLOBALS": true,
	"HTTP_RAW_POST_DATA": true, "php_errormsg": true, "http_response_header": true,
}

func (r issetArgumentExistence) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Binary: // D1
		if x.Op.Kind == syntax.TCoalesce {
			r.candidate(ctx, x.Left)
		}
	case *syntax.Isset: // D2
		for _, v := range x.Vars {
			r.candidate(ctx, v)
		}
	case *syntax.Empty: // D3
		r.candidate(ctx, x.Expr)
	}
}

func (issetArgumentExistence) candidate(ctx *analysis.Context, e syntax.Expr) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" || v.Span().Len() == 0 {
		return
	}
	name := v.Name
	if issetSpecialNames[name] { // D4
		return
	}
	fn := syntax.EnclosingFuncLike(v)
	body := syntax.FuncLikeBody(fn) // D6
	if fn == nil || body == nil {
		return
	}
	for _, p := range syntax.FuncLikeParams(fn) { // D5
		if p.Var != nil && p.Var.Name == name {
			return
		}
	}
	if c, ok := fn.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if u.Var != nil && u.Var.Name == name {
				return
			}
		}
	}

	// D7: first mention of name in the function's own scope.
	first := issetFirstMention(body, name)
	hasGoto, hasInclude, dynamic := false, false, false
	syntax.Inspect(body, func(c syntax.Node) bool {
		switch c := c.(type) {
		case *syntax.Goto:
			hasGoto = true
		case *syntax.Include:
			hasInclude = true
			if c.Span().Start < v.Span().Start { // D11: may define the variable
				dynamic = true
			}
		case *syntax.Eval: // D11: the evaluated code runs in this scope
			dynamic = true
		case *syntax.FuncCall: // D11: extract(), parse_str($s) define variables
			if ctx.IsGlobalFunctionCall(c, "extract") || (ctx.IsGlobalFunctionCall(c, "parse_str") && astquery.ArgCount(c) == 1) {
				dynamic = true
			}
		case *syntax.Assign: // D11: `$$name = …`
			if t, ok := c.Var.(*syntax.Variable); ok && t.NameExpr != nil {
				dynamic = true
			}
		}
		return true
	})
	if first == nil { // only in broken code (error recovery), e.g. `isset$x`
		return
	}
	if first != v {
		a, ok := first.Parent().(*syntax.Assign)
		if !ok || !isAncestor(a, v) {
			return
		}
	}
	if hasGoto || dynamic { // D9, D11
		return
	}
	if hasInclude && !ctx.Bool("IGNORE_INCLUDES") { // D10
		return
	}
	// D8 (custos: every enclosing loop, not only the innermost — an outer
	// loop's later statement sets the variable for the next iteration — and
	// every construct that defines the variable, not only `$n = …`).
	for p := first.Parent(); p != nil && p != fn; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if definesVar(ctx, p, name) {
				return
			}
		}
	}
	ctx.ReportSeverity(v.Span(), diagnostic.SeverityError, "Variable '$"+name+"' is not defined in this scope.")
}

func isAncestor(anc, n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if p == anc {
			return true
		}
	}
	return false
}

// definesVar reports whether root, staying in its function's scope, holds a
// construct that can leave $name defined for a later iteration: an
// assignment to it or to one of its elements (plain, compound, by
// reference, destructuring), `++`/`--`, a foreach key/value or catch
// binding, a `global`/`static` declaration, a closure importing it by
// reference, or a call argument that may bind by reference.
func definesVar(ctx *analysis.Context, root syntax.Node, name string) bool {
	found := false
	syntax.Inspect(root, func(c syntax.Node) bool {
		if found {
			return false
		}
		switch c := c.(type) {
		case *syntax.Assign:
			found = issetTargets(c.Var, name) || (c.ByRef && issetRoot(c.Value, name))
		case *syntax.IncDec:
			found = issetRoot(c.Var, name)
		case *syntax.Foreach:
			found = (c.Key != nil && issetTargets(c.Key, name)) || issetTargets(c.Value, name)
		case *syntax.Catch:
			found = c.Var != nil && c.Var.Name == name
		case *syntax.Global:
			for _, g := range c.Vars {
				found = found || issetRoot(g, name)
			}
		case *syntax.StaticVar:
			found = c.Var != nil && c.Var.Name == name
		case *syntax.FuncCall:
			found = passesByRef(ctx, c, c.Args, name)
		case *syntax.MethodCall:
			found = passesByRef(ctx, c, c.Args, name)
		case *syntax.StaticCall:
			found = passesByRef(ctx, c, c.Args, name)
		case *syntax.New:
			found = passesByRef(ctx, c, c.Args, name)
		case *syntax.Closure:
			for _, u := range c.Uses {
				found = found || (u.ByRef && u.Var != nil && u.Var.Name == name)
			}
			return false
		case *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		}
		return !found
	})
	return found
}

// passesByRef reports whether call hands $name (or one of its elements) to
// a by-reference parameter. A callee that cannot be resolved, and a
// position hidden behind an unpacked argument, may take it by reference.
func passesByRef(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList, name string) bool {
	if list == nil {
		return false
	}
	callee, resolved := semanticquery.ResolveCallee(ctx, call)
	unpacked := false
	for i, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok {
			continue
		}
		if arg.Unpack {
			unpacked = true
			continue
		}
		if !issetRoot(arg.Value, name) {
			continue
		}
		if arg.ByRef || !resolved || unpacked {
			return true
		}
		for j, p := range callee.Params {
			if !p.ByRef {
				continue
			}
			if arg.Name != nil {
				if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
					return true
				}
			} else if j == i || (p.Variadic && j < i) {
				return true
			}
		}
	}
	return false
}

// issetRoot reports whether e is $name or an element of it (`$name[…]`).
func issetRoot(e syntax.Expr, name string) bool {
	for {
		switch t := syntax.UnwrapParens(e).(type) {
		case *syntax.Variable:
			return t.NameExpr == nil && t.Name == name
		case *syntax.ArrayDimFetch:
			e = t.Var
		default:
			return false
		}
	}
}

// issetTargets reports whether the write target e is $name, one of its
// elements, or a destructuring pattern listing either.
func issetTargets(e syntax.Expr, name string) bool {
	switch t := syntax.UnwrapParens(e).(type) {
	case *syntax.List:
		for _, it := range t.Items {
			if it != nil && it.Value != nil && issetTargets(it.Value, name) {
				return true
			}
		}
	case *syntax.Array:
		for _, it := range t.Items {
			if it != nil && it.Value != nil && issetTargets(it.Value, name) {
				return true
			}
		}
	}
	return issetRoot(e, name)
}

// issetFirstMention returns the first plain variable named name in body, in
// source order, staying in body's scope: closures contribute only their
// `use` list (which reads or binds the outer variable), arrow functions
// their body unless a parameter shadows name, and nested named functions
// and classes nothing.
func issetFirstMention(body syntax.Node, name string) *syntax.Variable {
	var first *syntax.Variable
	syntax.Inspect(body, func(c syntax.Node) bool {
		if first != nil {
			return false
		}
		switch c := c.(type) {
		case *syntax.Variable:
			if c.NameExpr == nil && c.Name == name && !astquery.IsStaticPropName(c) {
				first = c
				return false
			}
		case *syntax.Closure:
			for _, u := range c.Uses {
				if u.Var != nil && u.Var.Name == name {
					first = u.Var
					break
				}
			}
			return false
		case *syntax.ArrowFunction:
			for _, p := range c.Params {
				if p.Var != nil && p.Var.Name == name {
					return false
				}
			}
		case *syntax.Function, *syntax.ClassLike:
			return false
		}
		return true
	})
	return first
}
