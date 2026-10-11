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
		// Only an XML Name can be declared, so any other run is never a reference.
		if xmlName(s[1:end]) {
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

// xmlName reports whether s matches the XML 1.0 Name production. Bytes that
// are not valid UTF-8 decode to U+FFFD, a name character, and stay unreported.
func xmlName(s string) bool {
	for i, r := range s {
		if !nameRune(r, i == 0) {
			return false
		}
	}
	return s != ""
}

func nameRune(r rune, start bool) bool {
	switch {
	case r == ':' || r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z':
		return true
	case r == '-' || r == '.' || r >= '0' && r <= '9' || r == 0xB7 || r >= 0x300 && r <= 0x36F || r == 0x203F || r == 0x2040:
		return !start
	}
	return r >= 0xC0 && r <= 0x2FF && r != 0xD7 && r != 0xF7 ||
		r >= 0x370 && r <= 0x1FFF && r != 0x37E ||
		r == 0x200C || r == 0x200D ||
		r >= 0x2070 && r <= 0x218F ||
		r >= 0x2C00 && r <= 0x2FEF ||
		r >= 0x3001 && r <= 0xD7FF ||
		r >= 0xF900 && r <= 0xFDCF ||
		r >= 0xFDF0 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0xEFFFF
}
