// Package pdofetchbothleaksduplicatecolumns implements the native PdoFetchBothLeaksDuplicateColumns inspection.
package pdofetchbothleaksduplicatecolumns

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use associative fetch mode before serializing rows."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoFetchBothLeaksDuplicateColumns" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "json_encode") {
		return
	}
	value := semanticquery.CallArgument(c.Args, 0, "value")
	producer, ok := semanticquery.NativeValue(ctx, value).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "PDOStatement", "fetchAll") {
		return
	}
	if v, ok := syntax.UnwrapParens(value).(*syntax.Variable); ok && iteratedByReference(ctx, c, v.Name) {
		return
	}
	arg := semanticquery.CallArgument(producer.Args, 0, "mode")
	constant, ok := syntax.UnwrapParens(arg).(*syntax.ClassConstFetch)
	if !ok {
		return
	}
	name, ok := constant.Name.(*syntax.Identifier)
	class, okClass := constant.Class.(*syntax.Name)
	if !ok || !okClass || name.Value != "FETCH_BOTH" {
		return
	}
	resolved := ctx.Names().Class(class.Value, class.Span().Start)
	if strings.EqualFold(strings.TrimPrefix(resolved, `\`), "PDO") {
		ctx.ReportNode(c, message)
	}
}

// iteratedByReference reports whether the scope holding at has a foreach that
// binds elements of the named variable by reference. Such a loop can rewrite
// every row in place, which flow resolution does not see as a write.
func iteratedByReference(ctx *analysis.Context, at syntax.Node, name string) bool {
	scope := syntax.EnclosingVariableScope(at)
	found := false
	visit := func(n syntax.Node) bool {
		if found || n != scope && syntax.IsVariableScope(n) {
			return false
		}
		f, ok := n.(*syntax.Foreach)
		if !ok {
			return true
		}
		if v, ok := syntax.UnwrapParens(f.Expr).(*syntax.Variable); ok && v.Name == name {
			found = f.ByRef || bindsReference(f.Value)
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

// bindsReference reports whether a destructuring target takes a reference.
func bindsReference(target syntax.Expr) bool {
	found := false
	syntax.Inspect(target, func(n syntax.Node) bool {
		if item, ok := n.(*syntax.ArrayItem); ok && item.ByRef {
			found = true
		}
		return !found
	})
	return found
}
