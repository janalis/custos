// Package jsonforceobjectchangesnestedlists implements the native JsonForceObjectChangesNestedLists inspection.
package jsonforceobjectchangesnestedlists

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve nested list representation during JSON encoding."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonForceObjectChangesNestedLists" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_encode") {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 1, "flags")
	yes, known := semanticquery.NativeFlag(ctx, flags, 16)
	if !known || !yes {
		return
	}
	if nested(ctx, call, semanticquery.CallArgument(call.Args, 0, "value"), false, 0) {
		ctx.ReportNode(flags, message)
	}
}

func nested(ctx *analysis.Context, call *syntax.FuncCall, e syntax.Expr, child bool, depth int) bool {
	if depth > 16 || rewritten(ctx, call, e) {
		return false
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, e)
	if !known {
		return false
	}
	if child && isList(ctx, semanticquery.NativeArray(ctx, e)) {
		return true
	}
	for _, item := range entries {
		if nested(ctx, call, item, true, depth+1) {
			return true
		}
	}
	return false
}

// isList reports whether the literal builds a nonempty array whose keys are
// 0..n-1 in insertion order, which is what json_encode emits as a JSON array.
// An overwrite keeps the position of the first occurrence of its key, so
// every key must be either the next unused index or an already inserted one.
func isList(ctx *analysis.Context, a *syntax.Array) bool {
	next := int64(0)
	for _, item := range a.Items {
		if item.Key != nil {
			key, _ := semanticquery.NativeArrayKey(ctx, item.Key)
			if !strings.HasPrefix(key, "i:") {
				return false
			}
			n, _ := strconv.ParseInt(key[2:], 10, 64)
			if n < 0 || n > next {
				return false
			}
			if n < next {
				continue
			}
		}
		next++
	}
	return next > 0
}

// rewritten reports whether e is a plain variable that its scope changes in
// a way the tracked value does not follow: a foreach that iterates it by
// reference or writes into it, a by-reference closure capture, a catch
// binding, or extract(). A variable copied from such a variable inherits the
// doubt. The scan ignores statement order.
func rewritten(ctx *analysis.Context, call *syntax.FuncCall, e syntax.Expr) bool {
	name, ok := astquery.PlainVariableName(syntax.UnwrapParens(e))
	return ok && rewrittenName(ctx, call, name, map[string]bool{})
}

func rewrittenName(ctx *analysis.Context, call *syntax.FuncCall, name string, seen map[string]bool) bool {
	if seen[name] {
		return false
	}
	seen[name] = true
	found := false
	var copied []string
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Foreach:
			found = found || (n.ByRef || bindsReference(n.Value)) && astquery.MentionsVariable(n.Expr, name) ||
				astquery.MentionsVariable(n.Key, name) || astquery.MentionsVariable(n.Value, name)
		case *syntax.ClosureUse:
			found = found || n.ByRef && n.Var.Name == name
		case *syntax.Catch:
			found = found || n.Var != nil && n.Var.Name == name
		case *syntax.FuncCall:
			_, extract := astquery.IsFuncNamedFold(n, "extract")
			found = found || extract
		case *syntax.Assign:
			if target, ok := astquery.PlainVariableName(n.Var); ok && target == name {
				if source, ok := astquery.PlainVariableName(syntax.UnwrapParens(n.Value)); ok {
					copied = append(copied, source)
				}
			}
		}
		return !found
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	for _, source := range copied {
		found = found || rewrittenName(ctx, call, source, seen)
	}
	return found
}

// bindsReference reports whether a foreach destructuring target takes one of
// its elements by reference.
func bindsReference(target syntax.Expr) bool {
	found := false
	if _, ok := target.(*syntax.Variable); ok || target == nil {
		return false
	}
	syntax.Inspect(target, func(n syntax.Node) bool {
		item, ok := n.(*syntax.ArrayItem)
		found = found || ok && item.ByRef
		return !found
	})
	return found
}
