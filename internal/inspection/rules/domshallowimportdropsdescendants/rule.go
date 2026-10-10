// Package domshallowimportdropsdescendants implements DomShallowImportDropsDescendants.
package domshallowimportdropsdescendants

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Import descendants when copying an element subtree."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomShallowImportDropsDescendants" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMDocument", "importNode") {
		return
	}
	deep := semanticquery.CallArgument(c.Args, 1, "deep")
	if deep != nil {
		value, known := semanticquery.NativeTruth(ctx, deep)
		if !known || value {
			return
		}
	}
	if !semanticquery.ExpansionDOMElementChildren(ctx, semanticquery.CallArgument(c.Args, 0, "node"), c) {
		return
	}
	var fixes []diagnostic.Fix
	if deep == nil && len(c.Args.Args) == 1 {
		a := c.Args.Args[0].(*syntax.Arg)
		suffix := ", true"
		if a.Name != nil {
			suffix = ", deep: true"
		}
		span := syntax.Span{Start: a.Span().End, End: a.Span().End}
		fixes = append(fixes, edit(span, suffix))
	} else if deep != nil {
		if constant, ok := deep.(*syntax.ConstFetch); ok && strings.EqualFold(constant.Name.Value, "false") {
			fixes = append(fixes, edit(deep.Span(), "true"))
		}
	}
	ctx.ReportNode(c, message, fixes...)
}

func edit(span syntax.Span, text string) diagnostic.Fix {
	return diagnostic.Fix{Title: message, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}
