// Package csvescapedefaultdependency implements the native CsvEscapeDefaultDependency inspection.
package csvescapedefaultdependency

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Pass the CSV escape character explicitly."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CsvEscapeDefaultDependency" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	if ctx.PHP < phpversion.PHP84 {
		return
	}
	args, pos, _, ok := semanticquery.NativeCSVSignature(ctx, n)
	if !ok {
		return
	}
	escape := pos + 2
	if semanticquery.CallArgument(args, escape, "escape") != nil {
		return
	}
	if c, ok := n.(*syntax.MethodCall); ok {
		if id := c.Name.(*syntax.Identifier); !strings.EqualFold(id.Value, "setCsvControl") {
			for _, p := range semanticquery.NativePriorCalls(ctx, c, c.Var, "setCsvControl") {
				control := p.(*syntax.MethodCall)
				if semanticquery.CallArgument(control.Args, 2, "escape") != nil && semanticquery.CallArgument(args, pos, "separator") == nil && semanticquery.CallArgument(args, pos+1, "enclosure") == nil {
					return
				}
			}
		}
	}
	for _, a := range args.Args {
		if arg, ok := a.(*syntax.Arg); !ok || arg.Unpack {
			ctx.ReportNode(n, message)
			return
		}
	}
	point := syntax.Span{Start: args.Span().End - 1, End: args.Span().End - 1}
	text := `escape: "\\"`
	if len(args.Args) > 0 {
		trailing := false
		last := args.Args[len(args.Args)-1].Span().End
		for _, tok := range ctx.File.Tokens {
			if tok.Start >= last && tok.End <= point.Start && tok.Kind == syntax.TComma {
				trailing = true
			}
		}
		if !trailing {
			text = ", " + text
		}
	}
	fix := diagnostic.Fix{Title: "Pass the CSV escape explicitly", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: point, NewText: text}} }}
	ctx.ReportNode(n, message, fix)
}
