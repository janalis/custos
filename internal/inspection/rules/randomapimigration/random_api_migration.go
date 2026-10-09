package randomapimigration

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// randomAPIMigration suggests the mt_* / random_int() replacements of the
// libc-based rand() family.
type randomAPIMigration struct{}

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (randomAPIMigration) Semantic()                {}
func (randomAPIMigration) ID() string               { return "RandomApiMigration" }
func (randomAPIMigration) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

var (
	randomClassic = map[string]string{"srand": "mt_srand", "getrandmax": "mt_getrandmax", "rand": "mt_rand"}
	randomModern  = map[string]string{"srand": "mt_srand", "getrandmax": "mt_getrandmax", "rand": "random_int", "mt_rand": "random_int"}
)

func (randomAPIMigration) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	nameNode, name, ok := astquery.FuncNamePart(call)
	if !ok {
		return
	}
	table := randomClassic // D1
	if ctx.Bool("SUGGEST_USING_RANDOM_INT") && ctx.PHP >= phpversion.PHP70 {
		table = randomModern
	}
	name = strings.ToLower(name) // function names are case-insensitive
	suggested, ok := table[name] // D2
	if !ok {
		return
	}
	if !semanticquery.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D3
		return
	}
	if suggested == "random_int" && astquery.ArgCount(call) != 2 { // D4
		if name != "rand" {
			return
		}
		suggested = "mt_rand"
	}
	span := astquery.NamePartSpan(nameNode)
	repl := suggested
	if !strings.Contains(nameNode.Value, `\`) { // a namespaced or imported function would capture a bare call
		repl = semanticquery.QualifiedBuiltin(ctx, suggested, call.Span().Start)
	}
	ctx.Report(call.Span(), "Prefer "+suggested+"() over "+name+"().", diagnostic.Fix{
		Title: "Use " + suggested + "()",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}
