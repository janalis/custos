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
	lo, hi, two := boundArgs(call)
	if suggested == "random_int" && !two { // D4
		if name != "rand" {
			return
		}
		suggested = "mt_rand"
	}
	msg := "Prefer " + suggested + "() over " + name + "()."
	// random_int() ignores srand()/mt_srand(), so renaming to it would break a
	// seeded sequence; mt_rand() rejects the reversed bounds rand() accepts.
	if suggested == "random_int" || two && name == "rand" && reversedBounds(ctx, lo, hi) {
		ctx.Report(call.Span(), msg)
		return
	}
	span := astquery.NamePartSpan(nameNode)
	repl := suggested
	if !strings.Contains(nameNode.Value, `\`) { // a namespaced or imported function would capture a bare call
		repl = semanticquery.QualifiedBuiltin(ctx, suggested, call.Span().Start)
	}
	ctx.Report(call.Span(), msg, diagnostic.Fix{
		Title: "Use " + suggested + "()",
		Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: repl}} },
	})
}

// boundArgs returns the two arguments of a call passing exactly two real
// ones; an unpacking or a first-class-callable placeholder has no known
// arity.
func boundArgs(call *syntax.FuncCall) (first, second *syntax.Arg, ok bool) {
	if astquery.ArgCount(call) != 2 {
		return nil, nil, false
	}
	first, _ = call.Args.Args[0].(*syntax.Arg)
	second, _ = call.Args.Args[1].(*syntax.Arg)
	if first == nil || second == nil || first.Unpack || second.Unpack {
		return nil, nil, false
	}
	return first, second, true
}

// reversedBounds reports whether both bounds are known integers with the
// minimum above the maximum. Named arguments may be written max first.
func reversedBounds(ctx *analysis.Context, lo, hi *syntax.Arg) bool {
	if lo.Name != nil && lo.Name.Value == "max" {
		lo, hi = hi, lo
	}
	a, aok := semanticquery.NativeInt(ctx, lo.Value)
	b, bok := semanticquery.NativeInt(ctx, hi.Value)
	return aok && bok && a > b
}
