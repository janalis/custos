package stringcasemanipulation

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// stringCaseManipulation reports position searches whose haystack/needle is
// case-converted only to make the search case-insensitive.
type stringCaseManipulation struct{}

func (stringCaseManipulation) ID() string { return "StringCaseManipulation" }
func (stringCaseManipulation) Semantic()  {}
func (stringCaseManipulation) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var caseInsensitiveSearch = map[string]string{
	"strpos":     "stripos",
	"mb_strpos":  "mb_stripos",
	"strrpos":    "strripos",
	"mb_strrpos": "mb_strripos",
}

func isCaseConversion(name string) bool {
	switch name {
	case "strtolower", "mb_strtolower", "strtoupper", "mb_strtoupper":
		return true
	}
	return false
}

func (stringCaseManipulation) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	variant, ok := caseInsensitiveSearch[ctx.GlobalFunctionName(call)] // D1
	if !ok {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok || len(args) != 2 { // D2
		return
	}
	var parts [2]string
	var dirs [2]byte // 'l' lower-cased, 'u' upper-cased, 0 kept
	found := false
	multibyte := strings.HasPrefix(variant, "mb_") // an mb_ search or conversion
	needle := args[1]                              // without its conversion
	for i, a := range args {                       // D3
		parts[i] = ctx.Text(a)
		inner, ok := a.(*syntax.FuncCall)
		if !ok {
			continue
		}
		conv := ctx.GlobalFunctionName(inner)
		if !isCaseConversion(conv) {
			continue
		}
		iargs, ok := astquery.CallArgValues(inner)
		if !ok || len(iargs) != 1 {
			continue
		}
		parts[i] = ctx.Text(iargs[0])
		multibyte = multibyte || strings.HasPrefix(conv, "mb_")
		if i == 1 {
			needle = iargs[0]
		}
		dirs[i] = 'l'
		if strings.HasSuffix(conv, "upper") {
			dirs[i] = 'u'
		}
		found = true
	}
	if !found { // D4
		return
	}
	// A namespaced function of that name would capture a bare call.
	repl := semanticquery.QualifiedBuiltin(ctx, variant, call.Span().Start) + "(" + parts[0] + ", " + parts[1] + ")"
	span := call.Span()
	msg := "Use '" + repl + "' instead of changing the case."
	// custos: mb_strtolower()/mb_strtoupper() map some letters to a different
	// number of characters ('ß' -> 'SS') and fold others than stripos() or
	// mb_stripos() do, so matches and offsets differ: only the byte family
	// is rewritten. Before PHP 8.0 stripos() reads a non-string needle as a
	// character code, where the conversion made it a string.
	if multibyte || !caseSearchEquivalent(args, dirs) ||
		(ctx.PHP < phpversion.PHP80 && !ctx.TypeOf(needle).OnlyOf("string")) {
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, diagnostic.Fix{
		Title: "Use '" + variant + "'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: repl}}
		},
	})
}

// caseSearchEquivalent reports whether the case-insensitive search finds
// exactly what the converted one does (custos): both sides converted the
// same way, or the kept side a literal without letters of the other case.
// `strpos(strtolower($name), $query)` never matches a query with capitals;
// stripos() would.
func caseSearchEquivalent(args []syntax.Expr, dirs [2]byte) bool {
	if dirs[0] != 0 && dirs[1] != 0 {
		return dirs[0] == dirs[1]
	}
	kept, dir := args[0], dirs[1]
	if dirs[0] != 0 {
		kept, dir = args[1], dirs[0]
	}
	v, ok := astquery.QuotedStringValue(kept)
	if !ok {
		return false
	}
	if dir == 'l' {
		return strings.ToLower(v) == v
	}
	return strings.ToUpper(v) == v
}
