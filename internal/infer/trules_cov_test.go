package infer

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// trulesLabels types the second argument of every `t('label', expr)` call
// of src with a T-rules typer configured by setup.
func trulesLabels(t *testing.T, src string, ver phpver.Version, setup func(*TRules)) map[string]string {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: ver})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := NewTRules(NewEnv(f, names.New(f), ix, ver))
	if setup != nil {
		setup(tr)
	}
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = strings.Join(tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).Atoms(), "|")
			}
		}
		return true
	})
	return got
}

func checkLabels(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, w := range want {
		if g, ok := got[k]; !ok || g != w {
			t.Errorf("%s: got %q want %q", k, g, w)
		}
	}
}

func TestTRulesCoverage(t *testing.T) {
	src := `<?php
class A {
    const C = 1;
    private int $typed;
    private $untyped;
    public function m(): int { return 1; }
    public static function s(): string { return ''; }
    public function self() {
        t('prop', $this->typed);
        t('propUntyped', $this->untyped);
        t('this', $this);
    }
}
class B extends A { public function b(): string { return ''; } }
/** @param string $doc */
function g($o, $m, $name, int $i, $doc) {
    t('classConst', A::C);
    t('not', !$i);
    t('static', A::s());
    t('staticMissing', A::nope());
    t('staticDyn', $o::s());
    t('unresolved', nope());
    t('mbString', mb_convert_encoding('a', 'UTF-8'));
    t('strstr', strstr('a', 'b'));
    t('docParam', $doc);
    t('concat', $i . 'x');
    t('ternUnknown', $i ? $nope : 1);
    $arr = [1];
    t('dim', $arr[0]);
    t('interp', "a$name");
    $s = 'a';
    t('compound', $s .= 'b');
    t('dynMethod', $o->$m());
    t('dynStatic', A::{$m}());
    t('argv', $_SERVER['argv']);
    t('rtf', $_SERVER['REQUEST_TIME_FLOAT']);
    t('srvOther', $_SERVER['HOME']);
    t('substr', substr_replace('abc', 'x', 1));
    t('pcba', preg_replace_callback_array([], ['a']));
    t('replInt', str_replace('a', 'b', $i));
    t('replNoSubject', str_replace('a', 'b', other: 'c'));
    t('mbUnknown', mb_convert_encoding($o, 'UTF-8'));
    t('mbArray', mb_convert_encoding(['a'], 'UTF-8'));
    t('mbObject', mb_convert_encoding(new B, 'UTF-8'));
    t('getClass', get_class($o));
    t('explode1', explode(','));
    t('explodeVar', explode($name, 'a'));
    t('purlHost', parse_url('x', PHP_URL_HOST));
    t('purl1', parse_url('x'));
    t('varvar', $$name);
    $x = new A;
    if ($x instanceof B) {
        t('narrowOutside', $x);
        t('envRecv', $x->b());
    }
}
`
	checkLabels(t, trulesLabels(t, src, phpver.PHP84, nil), map[string]string{
		"classConst": "int", "not": "bool", "static": "string", "staticMissing": "", "staticDyn": "",
		"unresolved": "", "mbString": "bool|string", "strstr": "bool|string", "docParam": "string", "concat": "string", "ternUnknown": "", "dim": "int",
		"prop": "int", "propUntyped": "", "this": "\\A", "interp": "string", "compound": "string",
		"dynMethod": "", "dynStatic": "", "envRecv": "string",
		"argv": "array", "rtf": "float", "srvOther": "string",
		"substr": "string", "pcba": "array|null", "replInt": "string", "replNoSubject": "array|string",
		"mbUnknown": "array|false|string", "mbArray": "array|bool", "mbObject": "array|false|string",
		"getClass": "string", "explode1": "array|bool", "explodeVar": "array|bool",
		"purlHost": "null|string", "purl1": "array|bool", "varvar": "", "narrowOutside": "",
	})
}

// A named argument for a parameter the callee does not declare (guard
// against stub data lacking it) matches nothing.
func TestTRulesNamedArgBeyondParams(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php str_replace('a', 'b', subject: 1);`), syntax.Options{Version: phpver.PHP84})
	tr := NewTRules(NewEnv(f, names.New(f), index.New(stubs.Index()), phpver.PHP84))
	call := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	fn := &index.Function{FQN: "str_replace"}
	if tr.arg(call, fn, 2) != nil {
		t.Error("no parameter 2: no argument")
	}
	if got, ok := tr.override(call, fn); !ok || strings.Join(got.Atoms(), "|") != "array|string" {
		t.Errorf("override: %v %v", got.Atoms(), ok)
	}
}

func TestTRulesSoundArithmeticArrays(t *testing.T) {
	src := `<?php
$a = [1];
$b = [2];
t('plus', $a + $b);
t('minus', $a - 1);
t('mixed', $a + 1);
t('plain', [] + []);
$i = 2;
t('float', $i * 1.5);
t('div', $i / 2);
t('pow', $i ** 2);
t('powVar', $i ** $i);
t('numStr', $i + '1');
t('intInt', $i - 1);
t('unknown', $i + $nope);
`
	checkLabels(t, trulesLabels(t, src, phpver.PHP84, func(r *TRules) { r.SoundArithmetic = true }),
		map[string]string{"plus": "array", "minus": "", "mixed": "", "plain": "array",
			"float": "float", "div": "float|int", "pow": "int", "powVar": "float|int", "numStr": "float|int", "intInt": "int", "unknown": ""})
}

// More than maxVarDefs assignments to one variable: unknown (cost cap).
func TestTRulesMaxVarDefs(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\n")
	for i := 0; i <= maxVarDefs; i++ {
		fmt.Fprintf(&b, "$v = %d;\n", i)
	}
	b.WriteString("t('capped', $v);\n$w = 1;\nt('small', $w);\n")
	checkLabels(t, trulesLabels(t, b.String(), phpver.PHP84, nil), map[string]string{"capped": "", "small": "int"})
}

func TestTRulesNilAndReentrant(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php $x = 1;`), syntax.Options{Version: phpver.PHP84})
	tr := NewTRules(NewEnv(f, names.New(f), index.New(stubs.Index()), phpver.PHP84))
	if !tr.TypeOf(nil).IsUnknown() {
		t.Error("nil expression")
	}
	// The recursion guard: an expression already being typed reads as
	// unknown (and is not cached as such).
	x := f.Stmts[0].(*syntax.ExprStmt).Expr
	tr.busy[x] = true
	if !tr.TypeOf(x).IsUnknown() {
		t.Error("busy expression must be unknown")
	}
	delete(tr.busy, x)
	if tr.TypeOf(x).IsUnknown() {
		t.Error("guard result must not be cached")
	}
}
