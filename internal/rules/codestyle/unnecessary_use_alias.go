package codestyle

import (
	"custos/internal/analysis/util"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// unnecessaryUseAlias reports import aliases equal to the imported name's
// last segment (`use A\B as B;`).
type unnecessaryUseAlias struct{}

func init() { register(unnecessaryUseAlias{}) }

func (unnecessaryUseAlias) ID() string { return "UnnecessaryUseAlias" }

func (unnecessaryUseAlias) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KUseItem}
}

func (unnecessaryUseAlias) Check(ctx *analysis.Context, n syntax.Node) {
	it := n.(*syntax.UseItem)
	if it.Alias == nil || it.Alias.Value == "" || it.Name == nil || it.Alias.Span().Len() == 0 { // D1
		return
	}
	use, ok := it.Parent().(*syntax.Use)
	if !ok {
		return
	}
	kind := it.Type
	if kind == syntax.UseNormal {
		kind = use.Type
	}
	name := it.Name.Value
	last := util.LastNamePart(name)
	// D2 / E1: case-sensitive, except for function imports (function names
	// are case-insensitive in PHP).
	if last != it.Alias.Value && (kind != syntax.UseFunction || !strings.EqualFold(last, it.Alias.Value)) {
		return
	}
	del := syntax.Span{Start: it.Name.Span().End, End: it.Alias.Span().End}
	ctx.Report(it.Alias.Span(), "Alias "+it.Alias.Value+" repeats the imported name; remove it.", analysis.Fix{
		Title: "Remove the alias",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: del}} },
	})
}
