// Package simplexmlmissingnodecheckedasnull implements the native SimpleXmlMissingNodeCheckedAsNull inspection.
package simplexmlmissingnodecheckedasnull

import (
	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Check XML child existence with isset."

type rule struct{}

// New returns the SimpleXmlMissingNodeCheckedAsNull inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SimpleXmlMissingNodeCheckedAsNull" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical && b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		if !syntax.IsNullConst(syntax.UnwrapParens(pair[1])) {
			continue
		}
		p, ok := syntax.UnwrapParens(pair[0]).(*syntax.PropertyFetch)
		if !ok {
			continue
		}
		id, ok := p.Name.(*syntax.Identifier)
		if !ok {
			continue
		}
		creation := semanticquery.NativeConstruction(ctx, p.Var, "SimpleXMLElement")
		if creation == nil {
			continue
		}
		if methods := semanticquery.ObjectMethodHistory(ctx, b, p.Var); methods == nil || len(methods) > 0 || !semanticquery.MembersUnwrittenBefore(ctx, p.Var, b) {
			continue
		}
		s, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(creation.Args, 0, "data"))
		if !known {
			continue
		}
		children, _, _, known := semanticquery.ParseLiteralXML(s)
		if !known {
			continue
		}
		found := false
		for _, name := range children {
			if name == id.Value {
				found = true
			}
		}
		if !found {
			ctx.ReportNode(n, message)
			return
		}
	}
}
