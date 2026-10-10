package infer_test

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func TestDynamicBuiltinResolution(t *testing.T) {
	cases := []struct {
		name, prefix, call, companion string
		clobber                       bool
	}{
		{"global", "", "extract($data)", "", true},
		{"case", "", "ExTrAcT($data)", "", true},
		{"explicit", "namespace App;", `\extract($data)`, "", true},
		{"fallback", "namespace App;", "extract($data)", "", true},
		{"alias", "namespace App; use function extract as unpack;", "unpack($data)", "", true},
		{"parse alias", "namespace App; use function parse_str as decode;", "decode($data)", "", true},
		{"parse outputs", "", "parse_str($data, $output)", "", false},
		{"local", "namespace App; function extract($data) {}", "extract($data)", "", false},
		{"companion", "namespace App;", "extract($data)", "<?php namespace App; function extract($data) {}", false},
		{"import", "namespace App; use function Other\\extract;", "extract($data)", "<?php namespace Other; function extract($data) {}", false},
		{"unresolved import", "namespace App; use function Other\\extract;", "extract($data)", "", false},
		{"qualified", "namespace App;", "Other\\extract($data)", "", false},
		{"extract callable", "", "extract(...)", "", false},
		{"parse callable", "", "parse_str(...)", "", false},
		{"dynamic callable", "", "$data()", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "<?php " + tc.prefix + " function run($data) { $value = 1; " + tc.call + "; t('read', $value); $value = 2; t('fresh', $value); }"
			f := syntax.Parse("main.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			if tc.companion != "" {
				other := syntax.Parse("other.php", []byte(tc.companion), syntax.Options{Version: phpversion.PHP84})
				ix.Add(index.Extract(other))
			}
			env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
			tr := infer.NewTRules(env)
			syntax.InspectFile(f, func(n syntax.Node) bool {
				c, ok := n.(*syntax.FuncCall)
				if !ok {
					return true
				}
				nm, ok := c.Name.(*syntax.Name)
				if !ok || nm.Value != "t" {
					return true
				}
				label := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal).Raw
				expr := c.Args.Args[1].(*syntax.Arg).Value
				wantUnknown := tc.clobber && label == "'read'"
				for _, got := range []bool{env.TypeOf(expr).IsUnknown(), tr.TypeOf(expr).IsUnknown()} {
					if got != wantUnknown {
						t.Errorf("%s: unknown=%v want %v", label, got, wantUnknown)
					}
				}
				return true
			})
		})
	}
}

func BenchmarkDynamicBuiltinResolution(b *testing.B) {
	src := "<?php namespace App; use function extract as unpack; function run($data) { $value = 1;" + strings.Repeat(" unpack($data); t($value);", 100) + "}"
	f := syntax.Parse("bench.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	ns := names.New(f)
	var reads []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				reads = append(reads, c.Args.Args[0].(*syntax.Arg).Value)
			}
		}
		return true
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := infer.NewEnv(f, ns, ix, phpversion.PHP84)
		for _, read := range reads {
			env.TypeOf(read)
		}
	}
}

// class_alias() names resolve to the original class.
func TestClassAlias(t *testing.T) {
	checkAnywhere(t, `<?php
namespace X;
class user { public function id(): int { return 1; } }
class_alias(user::class, \core_user::class);
class_alias('X\user', 'legacy_user');
class_alias('legacy_user', 'older_user');
class_alias('loop_a', 'loop_b');
class_alias('loop_b', 'loop_a');
class_alias($dyn, 'never');
class_alias(static::class, 'never2');
class_alias(1, 'never3');
class_alias(user::class);
class_alias(class: 'X\user', alias: 'named');
class_alias('', 'empty');
function f() {
    t('const', (new \core_user)->id());
    t('string', (new \legacy_user)->id());
    t('chain', (new \older_user)->id());
    t('cycle', (new \loop_a)->id());
}
`, map[string]string{"const": "int", "string": "int", "chain": "int", "cycle": "?unknown"})
	ix := index.New(nil)
	f := syntax.Parse("a.php", []byte("<?php class A {} class_alias('A', 'B');"), syntax.Options{Version: phpversion.PHP84})
	ix.Add(index.Extract(f))
	if ix.Class("B", 0) == nil {
		t.Fatal("alias not found")
	}
	ix.Remove("a.php")
	if ix.Class("B", 0) != nil {
		t.Fatal("alias kept after Remove")
	}
}
