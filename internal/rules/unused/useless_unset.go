package unused

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// uselessUnset reports `unset($param)` on a function parameter: it only
// drops the local binding.
type uselessUnset struct{}

func init() { register(uselessUnset{}) }

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
	names := map[string]bool{}
	for _, p := range params { // D1
		if p.Var.NameExpr == nil && p.Var.Name != "" {
			names[p.Var.Name] = true
		}
	}
	if len(names) == 0 {
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
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch g := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
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
			if u.Span().Len() == 0 || !util.Reachable(u, n) {
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
