package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/runner"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// checkWith is check with other project files indexed as the runner does
// (symbols plus index-time inferred returns).
func checkWith(t *testing.T, others map[string]string, src string, want map[string]string) {
	t.Helper()
	opt := syntax.Options{Version: phpver.PHP84}
	ix := index.New(stubs.Index())
	var all []*index.FileSymbols
	for path, s := range others {
		fs := runner.ExtractSymbols(path, []byte(s), opt)
		ix.Add(fs)
		all = append(all, fs)
	}
	for _, fs := range all {
		ix.DropStaleInferred(fs)
	}
	f := syntax.Parse("t.php", []byte(src), opt)
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	local := index.New(ix)
	local.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), local, phpver.PHP84)
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if es, ok := n.(*syntax.ExprStmt); ok {
			if c, ok := es.Expr.(*syntax.FuncCall); ok {
				if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && len(c.Args.Args) == 2 {
					label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
					got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).ShapeString()
				}
			}
		}
		return true
	})
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %s want %s", k, got[k], w)
		}
	}
}

func TestCrossFileInferredReturns(t *testing.T) {
	lib := `<?php
namespace Lib;
function plain() { return 1; }
function maybe($x) { if ($x) { return 'a'; } }
function viaLocal() { return plain() + 1; }
function viaOther() { return \Other\thing(); }
function viaBuiltin(string $s) { return strlen($s); }
function viaShadowed(string $s) { return strlen($s); }
function newObj() { return new Thing(); }
function arr() { return ['a' => 1, 'b' => 'x']; }
class Thing {
    private int $n = 0;
    public $loose;
    public function count() { return $this->n; }
    public function loose() { return $this->loose; }
    public final function fin() { return 2.5; }
    public static function make() { return new self(); }
    public function inherited() { return $this->parentProp; }
    private function hidden() { return 1; }
}
final class Sealed { public function get() { return true; } }
`
	shadow := `<?php
namespace Shadow;
function viaShadowed2(string $s) { return strlen($s); }
`
	shadowDecl := `<?php
namespace Shadow;
function strlen($s) { return $s; }
`
	polyfill := `<?php
function mb_str_pad($s) { return 1; }
`
	checkWith(t, map[string]string{"lib.php": lib, "shadow.php": shadow, "shadowdecl.php": shadowDecl, "polyfill.php": polyfill}, `<?php
use Lib\Thing;
use Lib\Sealed;
function run(Thing $t, Sealed $s) {
    t('plain', \Lib\plain());
    t('maybe', \Lib\maybe(1));
    t('viaLocal', \Lib\viaLocal());
    t('viaOther', \Lib\viaOther());
    t('viaBuiltin', \Lib\viaBuiltin('x'));
    t('shadowed', \Shadow\viaShadowed2('x'));
    t('newObj', \Lib\newObj());
    t('arr', \Lib\arr());
    t('virtual', $t->count());
    t('final', $t->fin());
    t('static', Thing::make());
    t('named', Thing::count());
    t('loose', Thing::loose());
    t('inherited', Thing::inherited());
    t('sealed', $s->get());
    t('polyfill', mb_str_pad('x'));
}
`, map[string]string{
		"plain":      "int",
		"maybe":      "null|string",
		"viaLocal":   "int",
		"viaOther":   "?unknown",
		"viaBuiltin": "int",
		"shadowed":   "?unknown", // Shadow\strlen exists in the project
		"newObj":     `\Lib\Thing`,
		"arr":        "array{a: int, b: string}",
		"virtual":    "?unknown", // may dispatch to an override
		"final":      "float",
		"static":     `\Lib\Thing`,
		"named":      "int",
		"loose":      "?unknown",
		"inherited":  "?unknown",
		"sealed":     "true",
		"polyfill":   "?unknown", // a builtin's polyfill body is not trusted
	})
}
