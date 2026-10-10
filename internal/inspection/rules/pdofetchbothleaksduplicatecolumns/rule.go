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
	producer, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "value")).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "PDOStatement", "fetchAll") {
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
