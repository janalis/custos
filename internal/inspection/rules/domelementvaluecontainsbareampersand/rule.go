// Package domelementvaluecontainsbareampersand implements DomElementValueContainsBareAmpersand.
package domelementvaluecontainsbareampersand

import (
	"encoding/xml"
	"regexp"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Append a text node for literal ampersand text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomElementValueContainsBareAmpersand" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMDocument", "createElement") {
		return
	}
	s, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "value"))
	if !known {
		return
	}
	for i := 0; i < len(s); i++ {
		if s[i] != '&' {
			continue
		}
		end := strings.IndexByte(s[i:], ';')
		if end < 0 {
			ctx.ReportNode(c, message)
			return
		}
		entity := s[i+1 : i+end]
		if entity != "amp" && entity != "lt" && entity != "gt" && entity != "apos" && entity != "quot" && !numericEntity(entity) && !declaredEntity(ctx, c, entity) {
			ctx.ReportNode(c, message)
			return
		}
		i += end
	}
}

func numericEntity(s string) bool {
	if len(s) < 2 || s[0] != '#' {
		return false
	}
	base := 10
	digits := s[1:]
	if strings.HasPrefix(digits, "x") {
		base = 16
		digits = digits[1:]
	}
	v, err := strconv.ParseUint(digits, base, 32)
	return err == nil && (v == 9 || v == 10 || v == 13 || v >= 32 && v <= 0xD7FF || v >= 0xE000 && v <= 0xFFFD || v >= 0x10000 && v <= 0x10FFFF)
}

func declaredEntity(ctx *analysis.Context, c *syntax.MethodCall, name string) bool {
	var prior []*syntax.MethodCall
	for _, node := range semanticquery.NativePriorCalls(ctx, c, c.Var, "loadXML") {
		if load, ok := node.(*syntax.MethodCall); ok && semanticquery.NativeMethod(ctx, load, "DOMDocument", "loadXML") {
			prior = append(prior, load)
		}
	}
	if len(prior) == 0 {
		return false
	}
	load := prior[len(prior)-1]
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(load.Args, 0, "source"))
	if !known || len(text) > 65536 {
		return false
	}
	decoder := xml.NewDecoder(strings.NewReader(text))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		if _, ok := token.(xml.StartElement); ok {
			return false
		}
		if directive, ok := token.(xml.Directive); ok {
			d := string(directive)
			if !strings.HasPrefix(d, "DOCTYPE ") {
				continue
			}
			open := strings.IndexByte(d, '[')
			close := strings.LastIndexByte(d, ']')
			if open < 0 || close < open {
				return false
			}
			subset := strings.TrimSpace(d[open+1 : close])
			found := false
			for subset != "" {
				match := entityDeclaration.FindStringSubmatch(subset)
				if match == nil {
					return false
				}
				if match[1] == name {
					found = true
				}
				subset = strings.TrimSpace(subset[len(match[0]):])
			}
			return found
		}
	}
}

var entityDeclaration = regexp.MustCompile(`^<!ENTITY\s+([A-Za-z_][A-Za-z0-9_.:-]*)\s+(?:"[^"%]*"|'[^'%]*')\s*>`)
