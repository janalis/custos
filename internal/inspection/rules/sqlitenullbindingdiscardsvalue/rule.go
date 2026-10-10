// Package sqlitenullbindingdiscardsvalue implements SqliteNullBindingDiscardsValue.
package sqlitenullbindingdiscardsvalue

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Pass null when using the SQLite null binding type."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteNullBindingDiscardsValue" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SQLite3Stmt", "bindValue") {
		return
	}
	t := semanticquery.CallArgument(c.Args, 2, "type")
	if !semanticquery.ExpansionGlobalConstant(ctx, t, "SQLITE3_NULL") {
		return
	}
	v := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "value"))
	switch x := v.(type) {
	case *syntax.Literal:
		if x.LitKind == syntax.LitString || x.LitKind == syntax.LitInt || x.LitKind == syntax.LitFloat {
			ctx.ReportNode(c, message)
		}
	case *syntax.ConstFetch:
		if strings.EqualFold(x.Name.Value, "true") || strings.EqualFold(x.Name.Value, "false") {
			ctx.ReportNode(c, message)
		}
	}
}
