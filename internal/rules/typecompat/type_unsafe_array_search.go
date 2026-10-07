package typecompat

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// typeUnsafeArraySearch asks for the strict argument of in_array() and
// array_search().
type typeUnsafeArraySearch struct{}

func init() { register(typeUnsafeArraySearch{}) }

func (typeUnsafeArraySearch) ID() string { return "TypeUnsafeArraySearch" }

func (typeUnsafeArraySearch) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (typeUnsafeArraySearch) Semantic() {}

func (typeUnsafeArraySearch) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	qual, name, ok := util.CallName(call)
	if lname := strings.ToLower(name); !ok || (lname != "in_array" && lname != "array_search") || util.ArgCount(call) != 2 { // D1, D2, E3
		return
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D1
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
	span := call.Span()
	newText := qual + name + "(" + ctx.Text(needle.Value) + ", " + ctx.Text(haystack.Value) + ", true)"
	ctx.Report(span, "Pass a third argument to say whether this search must be type-strict.", analysis.Fix{
		Title: "Make the search type-strict",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: newText}} },
	})
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
		if content == "" || tucNumeric(content) { // numeric strings compare numerically
			return false
		}
	}
	return true
}

// stringLiteralContent returns the raw text between the quotes of a quoted
// string literal (interpolated ones included).
func stringLiteralContent(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	if raw, _, ok := util.QuotedStringRaw(e); ok {
		return raw, true
	}
	if is, ok := e.(*syntax.InterpolatedString); ok && !is.Heredoc && !is.Backtick {
		if t := ctx.Text(is); len(t) >= 2 && t[0] == '"' && t[len(t)-1] == '"' {
			return t[1 : len(t)-1], true
		}
	}
	return "", false
}
