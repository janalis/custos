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

func TestArrayFilterCallbackBinding(t *testing.T) {
	checkShape(t, `<?php
function run($unknown, array $args) {
 $xs = [1, null, false];
 t('omitted', array_filter($xs));
 t('null', array_filter($xs, null));
 t('named', array_filter(callback: null, array: $xs));
 t('mode', array_filter(array: $xs, mode: ARRAY_FILTER_USE_KEY));
 t('unknown', array_filter($xs, $unknown));
 t('unpacked', array_filter($xs, ...$args));
 t('empty', array_filter([]));
 t('falseOnly', array_filter([false, null]));
 t('callback', array_filter($xs, fn($x) => true));
}
`, map[string]string{
		"omitted": "int[]", "null": "int[]", "named": "int[]", "mode": "int[]",
		"unknown": "false[]|int[]|null[]", "unpacked": "false[]|int[]|null[]",
		"empty": "array{}", "falseOnly": "array{}", "callback": "false[]|int[]|null[]",
	})
	src := `<?php t('null', array_filter([1, null], null)); t('omitted', array_filter([1, null]));`
	checkVer(t, phpver.PHP74, false, src, map[string]string{"null": "int[]|null[]", "omitted": "int[]"})
	checkVer(t, phpver.PHP80, false, src, map[string]string{"null": "int[]", "omitted": "int[]"})
}

func TestArrayMapInputFacts(t *testing.T) {
	checkShape(t, `<?php
/**
 * @param int[] $xs
 * @param non-empty-array<int, int> $full
 * @param array{a?: int} $optional
 */
function run($xs, $full, $optional, $unknown, array $args) {
 t('identity', array_map(null, ['a' => 1, 4 => 'x']));
 t('identityEmpty', array_map(null, []));
 t('identityNamed', array_map(array: ['a' => 1], callback: null));
 t('mapped', array_map(fn(): string => '', ['a' => 1, 4 => 2]));
 t('mappedEmpty', array_map('intval', []));
 t('mappedOptional', array_map('intval', $optional));
 t('mappedTyped', array_map('intval', $xs));
 t('mappedFull', array_map('intval', $full));
 t('multiNonEmpty', array_map(fn(): int => 1, [], ['a']));
 t('multiMaybeEmpty', array_map(fn(): int => 1, $xs, $xs));
 t('multiEmpty', array_map(fn(): int => 1, [], []));
 t('multiInvalid', array_map(fn(): int => 1, [1], $unknown));
 t('unknownCallback', array_map($unknown, [1]));
 t('unknownSpread', array_map(null, [1], ...$args));
 t('mappedSpread', array_map('intval', [1], ...$args));
 t('unknownNamed', array_map(null, [1], other: [2]));
 t('missing', array_map(null));
 t('invalid', array_map(null, $unknown));
 t('invalidMulti', array_map(null, [1], $unknown));
 t('zipEmpty', array_map(null, [], []));
 t('zip', array_map(null, ['a' => 1, 'b' => 2], [true]));
 t('zipRows', array_map(null, ['a' => 1], ['b' => 'x']));
 t('zipGeneric', array_map(null, $xs, $xs));
 t('zipOptional', array_map(null, [1], $optional));
}
`, map[string]string{
		"identity": "array{a: int, 4: string}", "identityEmpty": "array{}", "identityNamed": "int[]{a: int}",
		"mapped": "string[]{a: string, 4: string}", "mappedEmpty": "int[]{}", "mappedOptional": "int[]{a?: int}",
		"mappedTyped": "int[]", "mappedFull": "non-empty int[]", "multiNonEmpty": "non-empty int[]",
		"multiMaybeEmpty": "int[]", "multiEmpty": "int[]", "multiInvalid": "int[]", "unknownCallback": "array",
		"unknownSpread": "array", "mappedSpread": "int[]", "unknownNamed": "array", "missing": "array",
		"invalid": "array", "invalidMulti": "array", "zipEmpty": "array{}",
		"zip":     "array{0: array{0: int, 1: true}, 1: array{0: int, 1: null}}",
		"zipRows": "array{0: array{0: int, 1: string}}", "zipGeneric": "array[]", "zipOptional": "non-empty array[]",
	})
}

func TestArrayMapShapeCap(t *testing.T) {
	arrays := strings.Repeat(", [1]", 33)
	checkShape(t, "<?php t('zip', array_map(null"+arrays+")); t('mapped', array_map('intval'"+arrays+"));", map[string]string{"zip": "array", "mapped": "int[]"})
}

func TestArrayBuiltinRecovery(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`array_map('intval', [1], ...)`, "int[]"},
		{`array_map(null, [1], ...)`, "array"},
		{`array_filter([1, null], ...)`, "array"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := syntax.Parse("arrays.php", []byte("<?php "+tc.src+";"), syntax.Options{Version: phpver.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
			if got := env.TypeOf(f.Stmts[0].(*syntax.ExprStmt).Expr).ShapeString(); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func BenchmarkArrayBuiltinInference(b *testing.B) {
	f := syntax.Parse("arrays.php", []byte(`<?php array_map(null, ['a' => 1, 'b' => 2], ['x']); array_map('intval', ['a' => 1]); array_filter([1, null], null);`), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	resolver := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, resolver, ix, phpver.PHP84)
		for _, stmt := range f.Stmts {
			env.TypeOf(stmt.(*syntax.ExprStmt).Expr)
		}
	}
}
