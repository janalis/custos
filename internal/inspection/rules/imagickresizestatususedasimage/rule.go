// Package imagickresizestatususedasimage implements the native ImagickResizeStatusUsedAsImage inspection.
package imagickresizestatususedasimage

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the mutated Imagick object after resizing."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "ImagickResizeStatusUsedAsImage" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KPropertyFetch}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	// Only a file spelling it can reach resizeImage (see astquery.Mentions).
	if !astquery.Mentions(ctx.File, "resizeImage") {
		return
	}
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	var e syntax.Expr
	switch c := n.(type) {
	case *syntax.MethodCall:
		e = c.Var
	case *syntax.PropertyFetch:
		e = c.Var
	}
	c, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, c, "Imagick", "resizeImage") {
		return
	}
	if v, ok := syntax.UnwrapParens(e).(*syntax.Variable); ok && rebound(ctx, v) {
		return
	}
	ctx.ReportNode(n, message)
}

// rebound reports whether v, or a variable it was copied from, may have been
// given another value by something other than a plain assignment: a foreach
// target, a destructuring target, a catch binding, an argument of a call that
// binds by reference, or any binding that shares
// the variable's storage (reference alias, global, static, by-reference
// closure capture). Value rebinds only count after the last plain assignment
// read at v; shared storage counts wherever it is established.
func rebound(ctx *analysis.Context, v *syntax.Variable) bool {
	scope := syntax.EnclosingVariableScope(v)
	// A write later in a loop that encloses v reaches it on the next iteration.
	limit := v.Span().Start
	for p := v.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = p.Span().End
		}
	}
	named := func(e syntax.Expr) bool {
		x, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		return ok && x.Name == v.Name
	}
	mentions := func(e syntax.Expr) bool {
		found := false
		if e != nil {
			syntax.Inspect(e, func(n syntax.Node) bool {
				x, ok := n.(*syntax.Variable)
				found = found || ok && x.Name == v.Name
				return !found
			})
		}
		return found
	}
	var source *syntax.Variable
	shared, replaced := false, false
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			if c, ok := n.(*syntax.Closure); ok && c.Span().Start < limit {
				for _, u := range c.Uses {
					shared = shared || u.ByRef && u.Var.Name == v.Name
				}
			}
			return false
		}
		if n.Span().Start >= limit {
			return false
		}
		switch n := n.(type) {
		case *syntax.Foreach:
			replaced = replaced || mentions(n.Key) || mentions(n.Value)
		case *syntax.Catch:
			replaced = replaced || n.Var != nil && n.Var.Name == v.Name
		case *syntax.Global:
			for _, g := range n.Vars {
				shared = shared || named(g)
			}
		case *syntax.StaticVar:
			shared = shared || n.Var.Name == v.Name
		case *syntax.Arg:
			// A resolved callee that binds an argument by reference may write it.
			if list, ok := n.Parent().(*syntax.ArgList); ok && named(n.Value) {
				callee, _ := semanticquery.ResolveCallee(ctx, list.Parent())
				replaced = replaced || semanticquery.ArgBindsByRef(list, callee.Params, true)
			}
		case *syntax.Assign:
			switch target := n.Var.(type) {
			case *syntax.Array, *syntax.List:
				replaced = replaced || mentions(target)
			case *syntax.Variable:
				if n.ByRef {
					shared = shared || named(n.Value)
				} else if target.Name == v.Name && n.Value.Span().End < v.Span().Start {
					// NativeLocalValue reads this assignment: earlier rebinds are overwritten.
					replaced = false
					source, _ = syntax.UnwrapParens(n.Value).(*syntax.Variable)
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
	return shared || replaced || source != nil && rebound(ctx, source)
}
