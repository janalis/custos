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
	if space, _ := encodedSpace(ctx, semanticquery.CallArgument(query.Args, 0, "data"), 0); space {
		ctx.ReportNode(query, message)
	}
}

// maxDepth bounds the walk through nested literal query data.
const maxDepth = 8

// encodedSpace reports whether the literal query data provably serializes a
// space, and whether it provably serializes at least one pair. A key is only
// written when its value is: null and empty arrays drop the whole entry, so a
// spaced key counts only next to a value known to be emitted.
func encodedSpace(ctx *analysis.Context, data syntax.Expr, depth int) (space, emits bool) {
	if depth > maxDepth {
		return false, false
	}
	entries, known := semanticquery.NativeArrayEntries(ctx, data)
	if !known {
		return false, false
	}
	for key, value := range entries {
		valueSpace, valueEmits := false, false
		if text, ok := semanticquery.NativeString(ctx, value); ok {
			valueSpace, valueEmits = strings.Contains(text, " "), true
		} else if _, ok := semanticquery.NativeInt(ctx, value); ok {
			valueEmits = true
		} else {
			valueSpace, valueEmits = encodedSpace(ctx, value, depth+1)
		}
		if valueSpace || valueEmits && strings.Contains(key, " ") {
			return true, true
		}
		emits = emits || valueEmits
	}
	return false, emits
}
