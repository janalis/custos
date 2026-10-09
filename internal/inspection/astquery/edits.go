package astquery

import (
	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
)

func ReplaceFix(span syntax.Span, repl string) diagnostic.Fix {
	return diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	}
}

// MethodRemovalEdits deletes a method with its doc comment and trailing
// whitespace (F1).
func MethodRemovalEdits(f *syntax.File, m syntax.Node) []diagnostic.TextEdit {
	var out []diagnostic.TextEdit
	for _, s := range MemberRemovalSpans(f, m) {
		out = append(out, diagnostic.TextEdit{Span: s})
	}
	return out
}
