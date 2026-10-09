package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestArrayUnionAndSpread(t *testing.T) {
	checkShape(t, `<?php
function run(array $plain, $unknown, string $key) {
 $a = ['keep'=>1, 7=>'old']; $b = ['keep'=>'wrong', 'new'=>true, 8=>2.5];
 t('union', $a + $b);
 $a += $b; t('assigned', $a); t('key', $a['keep']);
 t('empty', [] + []); t('leftEmpty', [] + [1]); t('rightEmpty', [1] + []);
 t('spread', [9, ...[5=>'a', 8=>true], 4.5]);
 t('strings', ['x'=>1, ...['x'=>'later','y'=>false], 'y'=>2]);
 t('nested', [...[...[1,2]], 3]); t('emptySpread', [...[]]);
 t('unknown', [...$unknown]); t('unknownExplicit', [...$unknown, 1]);
 t('plain', [...$plain]); t('computed', [$key=>1, ...[2]]);
 t('arrayUnknown', [1] + $plain);
 /** @var int[] $ints */ $ints = source();
 t('typed', [...$ints]); t('typedUnion', $ints + [1]);
 /** @var array{x?: int} $optional */ $optional = source();
 t('optional', [...$optional]); t('optionalUnion', $optional + []);
 /** @var array{x: int, ...} $open */ $open = source();
 t('open', [...$open]); t('openUnion', $open + []);
 t('negative', [-4=>1, ...[3=>2], 3]);
 t('numeric', 1 + 2); $n=1; $n+=2; t('numericAssign',$n);
 t('nestedValue', [...[[1]], [2]]);
}
`, map[string]string{
		"union":    "array{keep: int, 7: string, new: true, 8: float}",
		"assigned": "array{keep: int, 7: string, new: true, 8: float}", "key": "int",
		"empty": "array{}", "leftEmpty": "int[]{0: int}", "rightEmpty": "int[]{0: int}",
		"spread":  "array{0: int, 1: string, 2: true, 3: float}",
		"strings": "array{x: string, y: int}", "nested": "int[]{0: int, 1: int, 2: int}", "emptySpread": "array{}",
		"unknown": "array", "unknownExplicit": "non-empty array", "plain": "array", "computed": "non-empty int[]",
		"arrayUnknown": "non-empty array", "typed": "int[]", "typedUnion": "non-empty int[]",
		"optional": "int[]", "optionalUnion": "int[]", "open": "non-empty array", "openUnion": "non-empty array",
		"negative": "int[]{-4: int, -3: int, -2: int}", "numeric": "int", "numericAssign": "int",
		"nestedValue": "array{0: int[]{0: int}, 1: int[]{0: int}}",
	})
}

func TestArraySpreadVersions(t *testing.T) {
	for _, tc := range []struct {
		version   phpver.Version
		src, want string
	}{
		{phpver.PHP73, `return [...[1]];`, "array"},
		{phpver.PHP74, `return [...[9=>1], 2];`, "int[]{0: int, 1: int}"},
		{phpver.PHP80, `return [...['x'=>1]];`, "non-empty array"},
		{phpver.PHP81, `return [...['x'=>1], 'x'=>'a'];`, "array{x: string}"},
		{phpver.PHP82, `return [-4=>1, ...[2]];`, "int[]{-4: int, 0: int}"},
	} {
		t.Run(fmt.Sprintf("%s_%s", tc.version, tc.src), func(t *testing.T) {
			// Parse permissively at a modern version, then infer at the target.
			f := syntax.Parse("t.php", []byte("<?php function f() {"+tc.src+"}"), syntax.Options{Version: phpver.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			env := infer.NewEnv(f, names.New(f), ix, tc.version)
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if r, ok := n.(*syntax.Return); ok {
					if got := env.TypeOf(r.Expr).ShapeString(); got != tc.want {
						t.Errorf("got %s want %s", got, tc.want)
					}
				}
				return true
			})
		})
	}
}

func TestArrayOperatorChangesPreserveNumericCompounds(t *testing.T) {
	check(t, `<?php
function run() {
 $n = 8; $n -= 2; t('subtract', $n);
 $m = 3; $m *= 2; t('multiply', $m);
 $p = 2; $p **= 3; t('power', $p);
}
`, map[string]string{"subtract": "int", "multiply": "int", "power": "int"})
}

func TestArrayOperationsShapeLimit(t *testing.T) {
	var entries strings.Builder
	for i := 0; i < 32; i++ {
		fmt.Fprintf(&entries, "'k%d'=>1,", i)
	}
	src := fmt.Sprintf("<?php $a=[%s]; t('cap',$a+['extra'=>2]); t('spread',[...$a, 'extra'=>2]); t('duplicates',$a+$a);", entries.String())
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name, ok := c.Name.(*syntax.Name); ok && name.Value == "t" {
				got := env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value)
				label := string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End])
				if label == "'duplicates'" {
					if len(got.ShapeKeys()) != 32 {
						t.Errorf("duplicate union lost shape: %s", got.ShapeString())
					}
				} else if got.HasShape() || got.ShapeString() != "non-empty int[]" {
					t.Errorf("%s: %s", label, got.ShapeString())
				}
			}
		}
		return true
	})
}

func BenchmarkArrayOperations(b *testing.B) {
	for _, tc := range []struct{ name, expression string }{
		{"union", "$a + $a"},
		{"union_assign", "$a += $a"},
		{"spread", "[...$a, ...$a]"},
		{"unknown_spread", "[...$unknown, ...$a]"},
		{"spread_chain", "[...$a, ...$a, ...$a, ...$a, ...$a, ...$a, ...$a, ...$a]"},
		{"spread_chain_over_cap", "[...$a, ...[1], ...$a, ...$a, ...$a, ...$a, ...$a, ...$a, ...$a]"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			var src strings.Builder
			src.WriteString("<?php function f($unknown) { $a=[")
			for i := 0; i < 32; i++ {
				fmt.Fprintf(&src, "'k%d'=>1,", i)
			}
			src.WriteString("]; return " + tc.expression + "; }")
			f := syntax.Parse("t.php", []byte(src.String()), syntax.Options{Version: phpver.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			var expr syntax.Expr
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if r, ok := n.(*syntax.Return); ok {
					expr = r.Expr
				}
				return true
			})
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
				env.TypeOf(expr)
			}
		})
	}
}

func TestArrayInferenceIncompleteItems(t *testing.T) {
	f := syntax.Parse("t.php", []byte("<?php"), syntax.Options{Version: phpver.PHP84})
	env := infer.NewEnv(f, names.New(f), index.New(stubs.Index()), phpver.PHP84)
	for _, items := range [][]*syntax.ArrayItem{{nil}, {{}}} {
		got := env.TypeOf(&syntax.Array{Items: items})
		if got.HasShape() || got.IsNonEmptyArray() || !got.OnlyOf("array") {
			t.Errorf("incomplete array: %s", got.ShapeString())
		}
	}
}
