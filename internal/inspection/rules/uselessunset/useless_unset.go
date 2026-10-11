package uselessunset

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// uselessUnset reports `unset($param)` on a function parameter: it only
// drops the local binding.
type uselessUnset struct{}

func (uselessUnset) ID() string { return "UselessUnset" }
func (uselessUnset) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

const uselessUnsetMsg = "Unsetting a parameter only drops the local variable; this unset() is pointless."

func (uselessUnset) Check(ctx *analysis.Context, n syntax.Node) {
	params := syntax.FuncLikeParams(n)
	body := syntax.FuncLikeBody(n)
	if len(params) == 0 || body == nil {
		return
	}
	// refs holds the names bound to a reference the unset would break: a
	// by-reference parameter, or a name aliased with `&` in the body.
	names, refs := map[string]bool{}, map[string]bool{}
	for _, p := range params { // D1
		if p.Var.NameExpr == nil && p.Var.Name != "" {
			names[p.Var.Name] = true
			if p.ByRef {
				refs[p.Var.Name] = true
			}
		}
	}
	if len(names) == 0 || uselessUnsetScopeExposed(ctx, body) {
		return
	}
	// E6: `global $p;` / `static $p;` rebinds the name; unsets after it are
	// no longer about the parameter. rebound maps a name to the earliest
	// rebinding offset.
	rebound := map[string]uint32{}
	rebind := func(v syntax.Expr, at uint32) {
		if vr, ok := v.(*syntax.Variable); ok && vr.NameExpr == nil && names[vr.Name] {
			if old, seen := rebound[vr.Name]; !seen || at < old {
				rebound[vr.Name] = at
			}
		}
	}
	alias := func(v syntax.Expr) {
		if vr, ok := v.(*syntax.Variable); ok && vr.NameExpr == nil && names[vr.Name] {
			refs[vr.Name] = true
		}
	}
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch g := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Closure:
			for _, use := range g.Uses {
				if use.ByRef {
					alias(use.Var)
				}
			}
			return false
		case *syntax.Assign:
			if g.ByRef {
				alias(g.Var)
				alias(g.Value)
			}
		case *syntax.Foreach:
			if g.ByRef {
				alias(g.Value)
			}
		case *syntax.Global:
			for _, v := range g.Vars {
				rebind(v, g.Span().Start)
			}
		case *syntax.StaticStmt:
			for _, v := range g.Vars {
				if v != nil && v.Var != nil {
					rebind(v.Var, g.Span().Start)
				}
			}
		}
		return true
	})
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch u := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false // D2: nested scopes are separate
		case *syntax.Unset:
			if u.Span().Len() == 0 || !flowquery.Reachable(u, n) {
				return false
			}
			for _, v := range u.Vars { // D3
				vr, ok := v.(*syntax.Variable)
				if !ok || vr.NameExpr != nil || !names[vr.Name] {
					continue
				}
				if at, ok := rebound[vr.Name]; ok && at < vr.Span().Start { // E6
					continue
				}
				if uselessUnsetObserved(ctx, n, u, vr.Name, refs[vr.Name]) {
					continue
				}
				if len(u.Vars) == 1 {
					ctx.Report(u.Span(), uselessUnsetMsg)
				} else {
					ctx.ReportNode(vr, uselessUnsetMsg)
				}
			}
			return false
		}
		return true
	})
}

// uselessUnsetScopeExposed reports whether the function body hands its
// local variables to code that reads them by name (custos): an included
// file (templates: `extract($data); unset($data); include $tpl;`), eval,
// get_defined_vars(), compact(), a variable variable (`$$n`) or, from PHP
// 7.0 where they return the current values, func_get_args() and
// func_get_arg(). Unsetting a parameter then keeps it out of that scope.
func uselessUnsetScopeExposed(ctx *analysis.Context, body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if found {
			return false
		}
		switch c := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Include, *syntax.Eval:
			found = true
		case *syntax.Variable:
			// `Foo::$$n` names a static property, not a local variable.
			sp, static := c.Parent().(*syntax.StaticPropertyFetch)
			found = c.NameExpr != nil && !(static && sp.Name == syntax.Expr(c))
		case *syntax.FuncCall:
			found = ctx.IsGlobalFunctionCall(c, "get_defined_vars") || ctx.IsGlobalFunctionCall(c, "compact") ||
				ctx.PHP.AtLeast(phpversion.PHP70) &&
					(ctx.IsGlobalFunctionCall(c, "func_get_args") || ctx.IsGlobalFunctionCall(c, "func_get_arg"))
		}
		return !found
	})
	return found
}

// uselessUnsetObserved reports whether the unset changes what later code of
// the scope sees (custos): a read of the name after the unset statement (or
// anywhere in a loop enclosing it) that the parameter's value or an earlier
// assignment may reach — `unset($userid); … if (!empty($userid))`, or an
// unset in a loop followed by `$config[$k] = …` that starts a fresh array.
// Rebinding writes (`=`, global, static, another unset) do not observe it,
// unless the name is bound to a reference (ref): the unset detaches it, so
// any later access but another detaching statement behaves differently.
func uselessUnsetObserved(ctx *analysis.Context, scope syntax.Node, u *syntax.Unset, name string, ref bool) bool {
	var loops []syntax.Node
	for p := u.Parent(); p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			loops = append(loops, p)
		}
	}
	end := u.Span().End
	for _, acc := range flowquery.VarAccesses(ctx.File, scope, name) {
		after := acc.Var.Span().Start >= end
		for _, l := range loops {
			after = after || l.Span().Contains(acc.Var.Span())
		}
		if !after {
			continue
		}
		if ref {
			switch acc.By.(type) {
			case *syntax.Unset, *syntax.Global, *syntax.StaticVar:
				continue // detach the name themselves
			}
			return true // e.g. `$p = …` would write through to the caller
		}
		if acc.Write && !acc.Compound {
			continue // rebinding: `=`, destructuring, foreach, global, static, unset, catch
		}
		defs, entry := flowquery.ReachingAssignmentsIn(ctx.File, scope, acc.Var, name)
		if entry {
			return true
		}
		for _, d := range defs {
			if d.Span().Start < end {
				return true
			}
		}
	}
	return false
}
