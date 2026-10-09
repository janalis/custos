package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestNullsafeChains(t *testing.T) {
	check(t, `<?php
class Leaf { public int $value; public static int $shared; public const NUMBER = 1; public function value(): int { return 1; } public static function shared(): int { return 1; } }
class Node { public Leaf $leaf; public ?Leaf $maybe; public function leaf(): Leaf { return $this->leaf; } public function maybe(): ?Leaf { return $this->maybe; } /** @return Leaf[] */ public function leaves(): array { return []; } public function take($arg): int { return 1; } }
function demo(?Node $node, Node $sure, $unknown) {
 t('property', $node?->leaf->value);
 t('method', $node?->leaf()->value());
 t('offset', $node?->leaves()[0]->value);
 t('static', $node?->leaf()::shared());
 t('staticprop', $node?->leaf()::$shared);
 t('const', $node?->leaf()::NUMBER);
 t('paren', ($node?->leaf())->value);
 t('certain', $sure?->leaf->value);
 t('null', null?->leaf->value);
 t('nullmethod', null?->leaf()->value());
 t('nulloffset', null?->leaf()[0]);
 t('nullable', $node?->maybe()?->value());
 t('argument', $sure->take($node?->leaf));
 t('indexsubchain', $sure->leaves()[$node?->leaf->value]->value);
 t('namesubchain', $sure->{$node?->leaf->value});
 t('unknown', $unknown?->leaf);
 t('unresolved', $node?->missing);
 $assigned = $node?->leaf(); t('assignment', $assigned->value);
 if ($node !== null) { t('narrow', $node?->leaf->value); }
 if ($node?->maybe !== null) { t('membernarrow', $node?->maybe); t('membercontinuation', $node?->maybe->value); t('methodcontinuation', $node?->maybe->value()); }
}
`, map[string]string{
		"property": "int|null", "method": "int|null", "offset": "int|null", "static": "int|null", "staticprop": "int|null", "const": "int|null", "paren": "int|null", "certain": "int", "null": "null", "nullmethod": "null", "nulloffset": "null", "nullable": "int|null", "argument": "int", "unknown": "?unknown", "unresolved": "?unknown", "assignment": "int", "narrow": "int", "indexsubchain": "int", "namesubchain": "?unknown", "membernarrow": `\Leaf`, "membercontinuation": "int", "methodcontinuation": "int",
	})
}

func TestNullsafeChainFacts(t *testing.T) {
	src := `<?php class Value { public int $n; public ?Value $maybe; public function get(): Value { return $this; } public static function n(): int { return 1; } } function demo(?Value $v, Value $sure) { $v?->get()->n; $v?->maybe; null?->get()->n; $sure->n; $v?->get()::n(); null?->get()::n(); }`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	var expressions []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if es, ok := n.(*syntax.ExprStmt); ok {
			expressions = append(expressions, es.Expr)
		}
		return true
	})
	wants := []struct {
		ordinary, evaluated string
		executes            bool
	}{
		{"int|null", "int", true}, {`\Value|null`, `\Value|null`, true}, {"null", "?unknown", false}, {"int", "int", true}, {"int|null", "int", true}, {"null", "?unknown", false},
	}
	for i, x := range expressions {
		w := wants[i]
		if got := env.TypeOf(x).String(); got != w.ordinary {
			t.Errorf("ordinary %d: %s", i, got)
		}
		got, executes := env.ChainTypeOf(x)
		if got.String() != w.evaluated || executes != w.executes {
			t.Errorf("fact %d: %s, %v", i, got, executes)
		}
		native := env.Native()
		if got := native.TypeOf(x).String(); got != w.ordinary {
			t.Errorf("native %d: %s", i, got)
		}
	}
	for _, specOnly := range []bool{false, true} {
		tr := infer.NewTRules(env)
		tr.SpecOnly = specOnly
		for _, i := range []int{4, 5} {
			got := tr.TypeOf(expressions[i]).String()
			want := wants[i].ordinary
			if specOnly {
				want = "int"
				if i == 5 {
					want = "?unknown"
				}
			}
			if got != want {
				t.Errorf("T-rules specOnly=%v index=%d: %s want %s", specOnly, i, got, want)
			}
		}
	}
}

func BenchmarkNullsafeChains(b *testing.B) {
	for _, safe := range []bool{false, true} {
		op := "->"
		label := "ordinary"
		if safe {
			op = "?->"
			label = "nullsafe"
		}
		b.Run(label, func(b *testing.B) {
			src := "<?php class Node { public Node $next; public int $value; } function demo(?Node $node) { $node" + op + "next" + strings.Repeat("->next", 20) + "->value; }"
			f := syntax.Parse("b.php", []byte(src), syntax.Options{Version: phpver.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			resolver := names.New(f)
			var expr syntax.Expr
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if es, ok := n.(*syntax.ExprStmt); ok {
					expr = es.Expr
				}
				return true
			})
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := infer.NewEnv(f, resolver, ix, phpver.PHP84)
				env.TypeOf(expr)
			}
		})
	}
}

func TestOrdinaryOffsetDoesNotPrimeReceiver(t *testing.T) {
	check(t, `<?php
function inspect(callable $handler) {
 if ($handler instanceof Closure) {
  $handler = [random_int(0, 1) ? new stdClass() : 'Worker', 'run'];
  $handler[0] ?: $handler = 'free';
 }
 if (is_object($handler)) {}
 elseif (!is_array($handler)) {}
 elseif (!is_object($handler[0])) {
  t('offset', $handler[0]);
  t('parenthesized', ($handler)[0]);
 }
}
`, map[string]string{"offset": "string", "parenthesized": `\stdClass|string`})
}
