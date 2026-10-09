package typeunsafearraysearch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// typeUnsafeArraySearch asks for the strict argument of in_array() and
// array_search().
type typeUnsafeArraySearch struct{}

func (typeUnsafeArraySearch) ID() string               { return "TypeUnsafeArraySearch" }
func (typeUnsafeArraySearch) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (typeUnsafeArraySearch) Semantic()                {}
func (typeUnsafeArraySearch) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, name, ok := astquery.CallName(call)
	if lname := strings.ToLower(name); !ok || (lname != "in_array" && lname != "array_search") || astquery.ArgCount(call) != 2 { // D1, D2, E3
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D1
		return
	}
	needle, ok1 := call.Args.Args[0].(*syntax.Arg)
	haystack, ok2 := call.Args.Args[1].(*syntax.Arg)
	if !ok1 || !ok2 { // `...` placeholder (rejected by PHP after another argument)
		return
	}
	if literalStringHaystack(ctx, haystack.Value) { // E1
		return
	}
	nt, ht := ctx.TypeOf(needle.Value), ctx.TypeOf(haystack.Value) // E2
	if !nt.IsUnknown() && !ht.IsUnknown() && len(nt.Atoms()) == 1 && len(ht.Atoms()) == 1 &&
		strings.TrimPrefix(ht.Atoms()[0], `\`) == strings.TrimPrefix(nt.Atoms()[0], `\`)+"[]" {
		return
	}
	// custos: no fixing. `true` changes the result whenever loose and strict
	// equality differ ('1' == 1, null == '', '1e1' == '10'): a config
	// string searched among int constants stops matching. The only cases
	// known to be equivalent (same scalar type on both sides) are already
	// exempt (E2).
	ctx.Report(call.Span(), "Pass a third argument to say whether this search must be type-strict.")
}

// literalStringHaystack implements E1.
func literalStringHaystack(ctx *analysis.Context, h syntax.Expr) bool {
	arr, ok := h.(*syntax.Array)
	if !ok || len(arr.Items) == 0 {
		return false
	}
	for _, it := range arr.Items {
		if it == nil || it.Key != nil || it.Unpack || it.Value == nil {
			return false
		}
		content, ok := stringLiteralContent(ctx, it.Value)
		if !ok {
			return false
		}
		content = strings.TrimSpace(content)
		if content == "" || astquery.IsComparisonNumericString(content) { // numeric strings compare numerically
			return false
		}
	}
	return true
}

// stringLiteralContent returns the raw text between the quotes of a quoted
// string literal (interpolated ones included).
func stringLiteralContent(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	if raw, _, ok := astquery.QuotedStringRaw(e); ok {
		return raw, true
	}
	if is, ok := e.(*syntax.InterpolatedString); ok && !is.Heredoc && !is.Backtick {
		if t := ctx.Text(is); len(t) >= 2 && t[0] == '"' && t[len(t)-1] == '"' {
			return t[1 : len(t)-1], true
		}
	}
	return "", false
}
