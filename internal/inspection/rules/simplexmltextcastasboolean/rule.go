// Package simplexmltextcastasboolean implements the native SimpleXmlTextCastAsBoolean inspection.
package simplexmltextcastasboolean

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Parse XML boolean text explicitly."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SimpleXmlTextCastAsBoolean" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	u := n.(*syntax.Unary)
	if u.Op.Kind != syntax.TBoolCast {
		return
	}
	creation := semanticquery.NativeConstruction(ctx, u.Expr, "SimpleXMLElement")
	if creation == nil {
		return
	}
	if methods := semanticquery.ExpansionDMethods(ctx, u, u.Expr); methods == nil || len(methods) > 0 || !semanticquery.ExpansionDPristine(ctx, u.Expr, u) {
		return
	}
	s, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(creation.Args, 0, "data"))
	if !known {
		return
	}
	_, text, plain, known := semanticquery.ExpansionDXML(s)
	if known && plain && (text == "true" || text == "false") {
		ctx.ReportNode(n, message)
	}
}
