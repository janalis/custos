// Package pdoarrayboundtosingleplaceholder implements the native PdoArrayBoundToSinglePlaceholder inspection.
package pdoarrayboundtosingleplaceholder

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Expand list values into separate SQL placeholders."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoArrayBoundToSinglePlaceholder" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	producer, ok := semanticquery.NativeValue(ctx, c.Var).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "PDO", "prepare") {
		return
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(c.Args, 0, "params"))
	if !known {
		return
	}
	for _, value := range entries {
		if semanticquery.NativeArray(ctx, value) != nil {
			ctx.ReportNode(c, message)
			return
		}
	}
}
