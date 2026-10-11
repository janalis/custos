// Package collatorsortdiscardsrequiredkeys implements the native CollatorSortDiscardsRequiredKeys inspection.
package collatorsortdiscardsrequiredkeys

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve keys during collation sorting."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CollatorSortDiscardsRequiredKeys" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "Collator", "sort") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "array")
	entries, ok := semanticquery.NativeArrayEntries(ctx, arg)
	if !ok || rebindable(ctx, c, arg) {
		return
	}
	for key := range entries {
		if strings.HasPrefix(key, "s:") {
			ctx.ReportNode(c, message, astquery.ReplaceFix(c.Name.Span(), "asort"))
			return
		}
	}
}

// rebindable reports whether the sorted variable, or a local it was copied
// from, can be replaced without a visible assignment in the same scope: a
// closure capturing it by reference or a catch clause binding it. The flow
// resolver does not model either, so the literal it resolves proves nothing.
func rebindable(ctx *analysis.Context, call syntax.Node, arg syntax.Expr) bool {
	v, ok := syntax.UnwrapParens(arg).(*syntax.Variable)
	if !ok {
		return false
	}
	unproven := map[string]bool{}
	copies := map[string][]string{}
	root := syntax.EnclosingFuncLike(call)
	visit := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Closure:
			for _, u := range x.Uses {
				if u.ByRef {
					unproven[u.Var.Name] = true
				}
			}
			return n == root
		case *syntax.Catch:
			if x.Var != nil {
				unproven[x.Var.Name] = true
			}
		case *syntax.Assign:
			to, toOK := syntax.UnwrapParens(x.Var).(*syntax.Variable)
			from, fromOK := syntax.UnwrapParens(x.Value).(*syntax.Variable)
			if toOK && fromOK {
				copies[to.Name] = append(copies[to.Name], from.Name)
			}
		}
		// Nested functions and methods own separate variable scopes.
		return n == root || !syntax.IsFuncLike(n)
	}
	if root != nil {
		syntax.Inspect(root, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	seen := map[string]bool{}
	for queue := []string{v.Name}; len(queue) > 0; queue = queue[1:] {
		name := queue[0]
		if unproven[name] {
			return true
		}
		if !seen[name] {
			seen[name] = true
			queue = append(queue, copies[name]...)
		}
	}
	return false
}
