// Package xmlwriterattributeaftercontent implements the native XmlWriterAttributeAfterContent inspection.
package xmlwriterattributeaftercontent

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Write attributes before element content."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlWriterAttributeAfterContent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.ExpansionDUnaliased(ctx, c.Var, n) {
		return
	}
	if !semanticquery.NativeMethod(ctx, c, "XMLWriter", "writeAttribute") && !semanticquery.NativeMethod(ctx, c, "XMLWriter", "startAttribute") {
		return
	}
	stack := []bool{}
	attribute := false
	for _, p := range semanticquery.ExpansionDMethods(ctx, c, c.Var) {
		switch semanticquery.ExpansionDMethodName(p) {
		case "startelement", "startelementns":
			stack = append(stack, false)
		case "endelement", "fullendelement":
			if len(stack) == 0 {
				return
			}
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				stack[len(stack)-1] = true
			}
		case "text", "writeelement", "writeelementns", "writecdata", "writeraw":
			if len(stack) > 0 && !attribute {
				stack[len(stack)-1] = true
			}
		case "openmemory", "openuri", "startdocument", "enddocument":
			stack = nil
		case "startattribute", "startattributens":
			attribute = true
		case "endattribute":
			attribute = false
		case "writeattribute", "writeattributens":
		default:
			return
		}
	}
	if len(stack) > 0 && stack[len(stack)-1] {
		ctx.ReportNode(n, message)
	}
}
