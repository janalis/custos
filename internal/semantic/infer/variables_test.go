package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// extract(), one-argument parse_str() and `$$name =` may set any local:
// later reads of variables defined before them are unknown.
func TestDynamicWritesClobber(t *testing.T) {
	src := `<?php
function g($c, array $v, $q, $name) {
    $n = (int) $c;
    extract($v);
    t('extract', $n);
    $fresh = 1;
    t('after', $fresh);
}
function h($q) { $p = 1; parse_str($q); t('parse', $p); $o = 1; parse_str($q, $out); t('parse2', $o); }
function k($name) { $m = 1; $$name = 'x'; t('varvar', $m); }
function l(array $rows) {
    $x = 1;
    foreach ($rows as $r) { t('loop', $x); extract($r); }
}
`
	checkAnywhere(t, src, map[string]string{
		"extract": "?unknown", "after": "int", "parse": "?unknown", "parse2": "int", "varvar": "?unknown", "loop": "?unknown",
	})
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				if lit.Raw == "'extract'" {
					if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value); !got.IsUnknown() {
						t.Errorf("T-rules extract: %s", got)
					}
				}
			}
		}
		return true
	})
}

// include/require may reassign any local of the scope.
func TestIncludeClobbers(t *testing.T) {
	checkAnywhere(t, `<?php
function f() {
    $total = '0';
    include 'x.php';
    t('after', $total);
    $n = 1;
    t('later', $n);
}
`, map[string]string{"after": "?unknown", "later": "int"})
}

// Every assignment replaces the variable's value: `.=` yields a string
// whatever the variable held.
func TestCompoundAssignmentReplaces(t *testing.T) {
	checkAnywhere(t, `<?php
function f($f, int $i) {
    $h = file_get_contents($f);
    $h .= 'x';
    t('concat', $h);
    $n = null;
    $n .= 'x';
    t('concatNull', $n);
    $m = null;
    $m = $m . 'x';
    t('concatExpr', $m);
    $k = '1';
    $k += $i;
    t('plus', $k);
    $r = 'a';
    $s = 'b';
    $r = &$s;
    t('byRef', $r);
}
`, map[string]string{"concat": "string", "concatNull": "string", "concatExpr": "string", "plus": "?unknown", "byRef": "string"})
}
