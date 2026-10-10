// Package formencodingusedforrfc3986signature implements the native FormEncodingUsedForRfc3986Signature inspection.
package formencodingusedforrfc3986signature

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Use RFC3986 query encoding for this signature protocol."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FormEncodingUsedForRfc3986Signature" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP54 || !semanticquery.NativeBuiltin(ctx, c, "hash_hmac") {
		return
	}
	scope, ok := syntax.EnclosingVariableScope(c).(*syntax.Function)
	if !ok {
		return
	}
	fqn, _ := ctx.Names().Function(scope.Name.Value, scope.Span().Start)
	configured := false
	for _, name := range ctx.List("signatureFunctions") {
		if strings.EqualFold(strings.TrimPrefix(name, "\\"), fqn) {
			configured = true
			break
		}
	}
	if !configured {
		return
	}
	query, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "data")).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, query, "http_build_query") {
		return
	}
	encoding := semanticquery.CallArgument(query.Args, 3, "encoding_type")
	if encoding != nil {
		value, known := semanticquery.NativeInt(ctx, encoding)
		if !known || value != 1 {
			return
		}
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(query.Args, 0, "data"))
	if !known {
		return
	}
	for _, value := range entries {
		text, known := semanticquery.NativeString(ctx, value)
		if known && strings.Contains(text, " ") {
			ctx.ReportNode(query, message)
			return
		}
	}
}
