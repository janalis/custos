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

func TestTRules(t *testing.T) {
	src := `<?php
$count = 4;
function f(int $n, $loose, string ...$rest) {
    t('div', $n / 2);
    t('mixf', $n * 0.5);
    t('unk', $n + $missing);
    t('neg', -2.25);
    t('coal', $loose ?? 1);
    t('coal2', $n ?? 'x');
    t('tern', $n ? 'a' : 1);
    t('var', $rest);
    t('mt0', microtime());
    t('mt1', microtime(true));
    t('expl', explode(',', 'a'));
    t('expl0', explode('', 'a'));
    t('purl', parse_url('x', PHP_URL_PORT));
    t('server', $_SERVER['argc']);
    t('get', $_GET['x']);
    t('repl', str_replace('a', 'b', 'c'));
    t('replnamed', str_replace(['a'], replace: 'b', subject: 'c'));
    t('replnamed2', str_replace(subject: 'c', search: 'a', replace: 'b'));
    t('abs', abs($n));
    t('end', end($rest));
    $x = 1.5;
    t('local', $x);
}
t('top', $count * 3);
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
	want := map[string]string{
		"div": "int", "mixf": "float", "unk": "float", "neg": "float", "coal": "", "coal2": "int|string",
		"tern": "int|string", "var": "array", "mt0": "int", "mt1": "float", "expl": "array", "expl0": "bool",
		"purl": "int|null", "server": "int", "get": "array|string", "repl": "string", "replnamed": "string", "replnamed2": "string", "abs": "int",
		"end": "mixed", "local": "float", "top": "int",
	}
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				ty := tr.TypeOf(c.Args.Args[1].(*syntax.Arg).Value)
				got[label] = strings.Join(ty.Atoms(), "|")
			}
		}
		return true
	})
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %q want %q", k, got[k], w)
		}
	}
}

func TestTRulesUnknownReachingDefinition(t *testing.T) {
	src := `<?php
function g($o, bool $c) {
    $f = $o->x;
    if ($c) {
        $f = 'x';
    }
    t('partial', $f);
    foreach ($o->list as $v) {
        if ($c) {
            $v = 'a';
        }
        t('element', $v);
    }
    foreach ($o->list as $w) {
        $w = 'a';
        t('overwritten', $w);
    }
    $g = 'y';
    if ($c) {
        $g = 'z';
    }
    t('known', $g);
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
	want := map[string]string{"partial": "", "known": "string", "element": "", "overwritten": "string"}
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
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %q want %q", k, got[k], w)
		}
	}
}

func TestTRulesForeachRebindingHidesEarlierAssignments(t *testing.T) {
	src := `<?php
function g(array $rows, array $names) {
    foreach ($rows as $item) {
        $item = 'text';
    }
    foreach ($names as $item) {
        t('rebound', $item);
        $item = 2;
        t('inner', $item);
    }
    t('after', $item);
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	tr := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
	want := map[string]string{"rebound": "", "inner": "int", "after": ""} // after: the binding also reaches (unmodelled)
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
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s: got %q want %q", k, got[k], w)
		}
	}
}

func TestTRulesSpecOnly(t *testing.T) {
	src := `<?php
class A { public function m(): int { return 1; } }
class B {}
function g(array $rows, A|B $o, A $a) {
    foreach ($rows as $item) {
        $item = 'text';
    }
    foreach ($rows as $item) {
        t('rebound', $item);
    }
    $s = 1;
    $s .= 'x';
    t('compound', $s);
    $q = [];
    $q = 'x';
    $r = &$q;
    t('byref', $r);
    foreach ([$a] as $obj) {
        t('recv', $obj->m());
    }
    t('union', $o->m());
    $subj = [];
    $subj = 'x';
    t('subject', str_replace('a', 'b', $subj));
    $list = [1];
    t('elems', $list + [2]);
    $arr = [];
    t('parlit', $arr + (1));
    $n = null;
    t('elvis', $n ?: 'a');
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	collect := func(tr *infer.TRules) map[string]string {
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
	spec := infer.NewTRules(env)
	spec.SpecOnly = true
	for mode, c := range map[string]struct {
		tr   *infer.TRules
		want map[string]string
	}{
		"default": {infer.NewTRules(env), map[string]string{
			"rebound": "", "compound": "string", "byref": "string", "recv": "int", "union": "",
			"subject": "string", "elems": "int", "parlit": "array", "elvis": "string",
		}},
		"specOnly": {spec, map[string]string{
			"rebound": "string", "compound": "int", "byref": "array|string", "recv": "", "union": "int",
			"subject": "array|string", "elems": "array", "parlit": "int", "elvis": "",
		}},
	} {
		got := collect(c.tr)
		for k, w := range c.want {
			if got[k] != w {
				t.Errorf("%s %s: got %q want %q", mode, k, got[k], w)
			}
		}
	}
}
