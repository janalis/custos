// Package untrustedxmlexternalentityexpansion implements the native UntrustedXmlExternalEntityExpansion inspection.
package untrustedxmlexternalentityexpansion

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Disable external entity expansion for request-controlled XML."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedXmlExternalEntityExpansion" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMDocument", "loadXML") {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "source")
	fetch, ok := input.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	variable, ok := fetch.Var.(*syntax.Variable)
	if !ok {
		return
	}
	switch variable.Name {
	case "_GET", "_POST", "_REQUEST", "_COOKIE":
	default:
		return
	}
	// Keep the request source proof within an untouched straight-line prefix.
	// Unknown preceding operations may replace or sanitize superglobal fields.
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) > 64 {
		return
	}
	for _, statement := range prior {
		expression, ok := statement.(*syntax.ExprStmt)
		if !ok {
			return
		}
		assignment, ok := expression.Expr.(*syntax.Assign)
		if !ok || assignment.ByRef || assignment.Op.Kind != syntax.TEqual {
			return
		}
		target, ok := assignment.Var.(*syntax.Variable)
		if !ok || target.Name == variable.Name {
			return
		}
		creation, ok := assignment.Value.(*syntax.New)
		if !ok || !semanticquery.NamesGlobalClass(ctx, creation.Class, "DOMDocument") || creation.Args != nil && len(creation.Args.Args) != 0 {
			return
		}
	}
	options := semanticquery.CallArgument(c.Args, 1, "options")
	entity, known := semanticquery.NativeFlag(ctx, options, 2)
	if !known || !entity {
		return
	}
	dtd, known := semanticquery.NativeFlag(ctx, options, 4)
	if !known || !dtd {
		return
	}
	noxxe, known := semanticquery.NativeFlag(ctx, options, 8388608)
	if !known || noxxe && ctx.PHP.AtLeast(phpversion.PHP84) {
		return
	}
	ctx.ReportNode(c, message)
}
