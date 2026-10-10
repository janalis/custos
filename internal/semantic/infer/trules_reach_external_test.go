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

// The SpecOnly T-rules typer drops assignments followed by an exit too.
func TestSpecOnlyExits(t *testing.T) {
	src := `<?php
function f(string $sql, array $c) {
    if ($c) { $sql = array_shift($c); return $sql; }
    $y = $sql;
    t('y', str_replace('a', 'b', $y));
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	tr.SpecOnly = true
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != "string" {
					t.Errorf("SpecOnly: %s", got)
				}
			}
		}
		return true
	})
}

// The casting typer sees assignments later in a loop body on the next
// iteration.
func TestTRulesLoopBackEdge(t *testing.T) {
	src := `<?php
function f(array $rows, string $p) {
    $n = 0;
    foreach ($rows as $r) {
        t('n', $n);
        $n = (string) $r;
    }
    t('pathinfo', pathinfo($p, PATHINFO_FILENAME));
    t('safe', \Safe\preg_replace('/a/', 'b', $p));
}
namespace Safe;
/** @return string|array|null */
function preg_replace($p, $r, $s) {}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				want := map[string]string{"'n'": "int|string", "'pathinfo'": "string", "'safe'": "string"}[lit.Raw]
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != want {
					t.Errorf("%s: got %s want %s", lit.Raw, got, want)
				}
			}
		}
		return true
	})
}

// The casting typer gives up on variables also defined by constructs it
// does not type (destructuring, out arguments…).
func TestTRulesUnmodelledDefinitions(t *testing.T) {
	src := `<?php
function f(string $t, array $xs, $c) {
    list($h, $m) = explode(':', $t);
    if (!$h) { $h = 0; }
    t('list', $h);
    $p = 1;
    preg_match('/x/', $t, $p);
    t('out', $p);
    $q = 1;
    if ($c) { $q = 2; }
    t('plain', $q);
    /** @var int $d */
    $d = 1;
    t('doc', $d);
    foreach ($xs as $x) { t('binding', $x); }
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpversion.PHP84))
	want := map[string]string{"'list'": "?unknown", "'out'": "?unknown", "'plain'": "int", "'doc'": "int", "'binding'": "?unknown"}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				lit := c.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal)
				if got := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).String(); got != want[lit.Raw] {
					t.Errorf("%s: got %s want %s", lit.Raw, got, want[lit.Raw])
				}
			}
		}
		return true
	})
}
