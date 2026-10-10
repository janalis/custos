// Package simplexmlchildvaluecontainsbareampersand implements the native SimpleXmlChildValueContainsBareAmpersand inspection.
package simplexmlchildvaluecontainsbareampersand

import (
	"strings"

	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape bare ampersands in XML child values."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SimpleXmlChildValueContainsBareAmpersand" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "SimpleXMLElement", "addChild") {
		return
	}
	s, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "value"))
	if !known || !strings.Contains(s, "&") {
		return
	}
	for {
		i := strings.IndexByte(s, '&')
		if i < 0 {
			return
		}
		s = s[i:]
		end := strings.IndexByte(s, ';')
		if end < 0 {
			ctx.ReportNode(n, message)
			return
		}
		// Named entities can be defined by the document DTD. Their absence is unknown.
		if reference := s[1:end]; reference != "" && reference[0] != '#' && !strings.ContainsAny(reference, "&< \t\r\n") {
			s = s[end+1:]
			continue
		}
		_, _, _, valid := semanticquery.ExpansionDXML("<r>" + s[:end+1] + "</r>")
		if !valid {
			ctx.ReportNode(n, message)
			return
		}
		s = s[end+1:]
	}
}
