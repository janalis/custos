package infer

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

const fixpointChain = `class Base { public function next(): Base { return $this; } } final class Q extends Base {} final class P extends Base { public function next(): Q { return new Q(); } } `

// A loop-carried read narrowed back to a fixed class is a cycle whose first
// type, assumed, is inferred again: the fixpoint is the type, with no gap.
func TestRecursionFixpoint(t *testing.T) {
	for _, tc := range []struct {
		src, want string
		gap       bool
	}{
		{`final class P { public function getParent(): ?P { return null; } } function g(P $start) { $a = $start; while (true) { $parent = $a->getParent(); if (!$parent instanceof P) { break; } $a = $parent; } probe($a); }`, `\P`, false},
		{`final class P { public function getParent(): ?P { return null; } } function g(P $start) { $a = $start; while (true) { $parent = $a->getParent(); if (!$parent instanceof P) { break; } $a = $parent; } probe($a->getParent()); }`, `\P|null`, false},
		{`final class P { public function getParent(): ?P { return null; } } function g(P $start) { $a = [$start]; while (true) { $p = $a[0]->getParent(); if (!$p instanceof P) { break; } $a = [$p]; } probe($a[0]); }`, `\P`, false},
		// Each assumed type yields a larger one: no fixpoint, reported as before.
		{fixpointChain + `function g(P $start) { $a = $start; while (true) { $p = $a->next(); if (!$p instanceof Base) { break; } $a = $p; } probe($a); }`, `\P|\Q`, true},
		// An unknown first type proves nothing.
		{`function g($x) { while (true) { $p = $x->getParent(); if (!$p) { break; } $x = $p; } probe($x); }`, `?unknown`, true},
	} {
		env, file := assertionEnv(t, tc.src, false, phpversion.PHP84)
		got := env.TypeOf(assertionRead(file))
		gap := false
		for _, g := range env.Gaps() {
			gap = gap || g.Reason == "inference-recursion"
		}
		if got.String() != tc.want || gap != tc.gap {
			t.Errorf("%s: type %s gap %v, want %s gap %v", tc.src, got, gap, tc.want, tc.gap)
		}
	}
}

// Re-inference is bounded per Env; past the bound cycles stay unknown and are
// reported, so hostile files cannot multiply inference work.
func TestRecursionFixpointBudget(t *testing.T) {
	var src strings.Builder
	src.WriteString(`final class P { public function getParent(): ?P { return null; } } `)
	for i := range maxFixpoints + 1 {
		fmt.Fprintf(&src, `function g%d(P $start) { $a = $start; while (true) { $parent = $a->getParent(); if (!$parent instanceof P) { break; } $a = $parent; } probe%d($a); } `, i, i)
	}
	env, file := assertionEnv(t, src.String(), false, phpversion.PHP84)
	var reads []syntax.Expr
	syntax.InspectFile(file, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			if name, ok := call.Name.(*syntax.Name); ok && strings.HasPrefix(name.Value, "probe") {
				reads = append(reads, call.Args.Args[0].(*syntax.Arg).Value)
			}
		}
		return true
	})
	for _, read := range reads {
		env.TypeOf(read)
	}
	if env.fixpoints != maxFixpoints || len(env.Gaps()) == 0 {
		t.Fatalf("fixpoints=%d gaps=%v", env.fixpoints, env.Gaps())
	}
}

// Every type-dependent cache written during a query can be retracted.
func TestRecursionRollback(t *testing.T) {
	env, file := assertionEnv(t, `probe(f());`, false, phpversion.PHP84)
	x := assertionRead(file)
	call := x.(*syntax.FuncCall)
	v := &syntax.Variable{Name: "v"}
	span := syntax.Span{Start: 1, End: 2}
	env.typing = 1
	env.cache[x] = types.Unknown
	env.bases = map[*syntax.Variable]types.Type{v: types.Unknown}
	env.chains = map[syntax.Expr]nullsafeFact{x: {}}
	env.bodies = map[syntax.Span]types.Type{span: types.Unknown}
	env.props = map[syntax.Span]types.Type{span: types.Unknown}
	env.captures = map[captureKey]types.Type{{x, "c"}: types.Unknown}
	env.asserts = map[syntax.Expr]*callAsserts{x: nil}
	env.generics = map[string]map[string]tplBindings{"g": nil}
	env.functions = map[*syntax.FuncCall]*index.Function{call: nil}
	env.builtinCalls = map[*syntax.FuncCall]bool{call: true}
	for _, u := range []undo{{kind: undoCache, expr: x}, {kind: undoBase, v: v}, {kind: undoChain, expr: x}, {kind: undoBody, span: span}, {kind: undoProp, span: span}, {kind: undoCapture, expr: x, text: "c"}, {kind: undoAssert, expr: x}, {kind: undoGeneric, text: "g"}, {kind: undoFunction, call: call}, {kind: undoBuiltin, call: call}} {
		env.remember(u)
	}
	env.rollback(0)
	if len(env.cache)+len(env.bases)+len(env.chains)+len(env.bodies)+len(env.props)+len(env.captures)+len(env.asserts)+len(env.generics)+len(env.functions)+len(env.builtinCalls)+len(env.journal) != 0 {
		t.Fatal("a cache entry survived rollback")
	}
	env.typing = 0
	env.remember(undo{kind: undoCache, expr: x})
	if len(env.journal) != 0 {
		t.Fatal("journal written outside a type query")
	}
}
