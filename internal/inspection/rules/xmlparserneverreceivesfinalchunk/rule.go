// Package xmlparserneverreceivesfinalchunk implements the native XmlParserNeverReceivesFinalChunk inspection.
package xmlparserneverreceivesfinalchunk

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Finish parsing before freeing the XML parser."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlParserNeverReceivesFinalChunk" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "xml_parser_free") {
		return
	}
	e := semanticquery.CallArgument(c.Args, 0, "parser")
	if !semanticquery.ExpansionDUnaliased(ctx, e, n) {
		return
	}
	creation, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok || (!semanticquery.NativeBuiltin(ctx, creation, "xml_parser_create") && !semanticquery.NativeBuiltin(ctx, creation, "xml_parser_create_ns")) {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 {
		return
	}
	parsed := false
	for _, record := range calls {
		p, ok := record.Node.(*syntax.FuncCall)
		if !ok || p.Span().End >= n.Span().Start || !semanticquery.NativeSameValue(ctx, semanticquery.CallArgument(p.Args, 0, "parser"), e) {
			continue
		}
		if !semanticquery.NativeDominates(p, n) {
			return
		}
		switch semanticquery.NativeBuiltinName(ctx, p) {
		case "xml_parse":
			final := semanticquery.CallArgument(p.Args, 2, "is_final")
			if final != nil {
				v, k := semanticquery.NativeTruth(ctx, final)
				if !k || v {
					return
				}
			}
			parsed = true
		case "xml_parse_into_struct", "xml_parser_free":
			return
		}
	}
	if parsed {
		ctx.ReportNode(n, message)
	}
}
