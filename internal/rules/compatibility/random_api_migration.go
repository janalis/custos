package compatibility

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// randomApiMigration suggests the mt_* / random_int() replacements of the
// libc-based rand() family.
type randomApiMigration struct{}

func init() { register(randomApiMigration{}) }

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (randomApiMigration) Semantic() {}

func (randomApiMigration) ID() string { return "RandomApiMigration" }

func (randomApiMigration) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

var (
	randomClassic = map[string]string{"srand": "mt_srand", "getrandmax": "mt_getrandmax", "rand": "mt_rand"}
	randomModern  = map[string]string{"srand": "mt_srand", "getrandmax": "mt_getrandmax", "rand": "random_int", "mt_rand": "random_int"}
)

func (randomApiMigration) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	nameNode, name, ok := util.FuncNamePart(call)
	if !ok {
		return
	}
	table := randomClassic // D1
	if ctx.Bool("SUGGEST_USING_RANDOM_INT") && ctx.PHP >= phpver.PHP70 {
		table = randomModern
	}
	name = strings.ToLower(name) // function names are case-insensitive
	suggested, ok := table[name] // D2
	if !ok {
		return
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D3
		return
	}
	if suggested == "random_int" && util.ArgCount(call) != 2 { // D4
		if name != "rand" {
			return
		}
		suggested = "mt_rand"
	}
	span := util.NamePartSpan(nameNode)
	repl := suggested
	if !strings.Contains(nameNode.Value, `\`) { // a namespaced or imported function would capture a bare call
		repl = util.QualifiedBuiltin(ctx, suggested, call.Span().Start)
	}
	ctx.Report(call.Span(), "Prefer "+suggested+"() over "+name+"().", analysis.Fix{
		Title: "Use " + suggested + "()",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
