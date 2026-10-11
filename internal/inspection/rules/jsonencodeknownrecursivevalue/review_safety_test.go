package jsonencodeknownrecursivevalue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const self = "$a = []; $a['self'] = &$a; "
	const pair = "$a = []; $b = []; $a['b'] = &$b; $b['a'] = &$a; "
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"self", "<?php function f() { " + self + "return json_encode($a); }", 1},
		{"pair", "<?php function f() { " + pair + "return json_encode($a); }", 1},
		{"unrelated statements", "<?php function f() { $n = 1; " + self + "$n++; echo $n; return json_encode($a); }", 1},
		{"nested scope ignored", "<?php " + self + "function g() { $a = 1; return new X(); } json_encode($a);", 1},
		{"later rebinding", "<?php function f() { " + self + "$r = json_encode($a); [$a] = [[1]]; return $r; }", 1},
		{"assigned result", "<?php function f() { " + self + "$r = json_encode($a); return $r; }", 1},
		{"result replaces root", "<?php function f() { " + self + "$a = json_encode($a); return $a; }", 0},
		{"enclosing call", "<?php function f() { " + self + "return strlen(json_encode($a)); }", 1},
		{"earlier argument call", "<?php function f() { " + self + "return max(reset($a), json_encode($a)); }", 0},
		{"destructured root", "<?php function f() { " + self + "[$a] = [[1]]; return json_encode($a); }", 0},
		{"list root", "<?php function f() { " + self + "list($a) = [[1]]; return json_encode($a); }", 0},
		{"foreach root", "<?php function f() { " + self + "foreach ([[1]] as $a) {} return json_encode($a); }", 0},
		{"foreach key", "<?php function f() { " + self + "foreach ([[1]] as $a => $v) {} return json_encode($a); }", 0},
		{"foreach slot", "<?php function f() { " + pair + "foreach ([1] as $b['a']) {} return json_encode($a); }", 0},
		{"destructured slot", "<?php function f() { " + pair + "[$b['a']] = [1]; return json_encode($a); }", 0},
		{"decremented slot", "<?php function f() { " + pair + "$b['a']--; return json_encode($a); }", 0},
		{"incremented root", "<?php function f() { " + self + "$a++; return json_encode($a); }", 0},
		{"alias", "<?php function f() { " + self + "$c = &$a; $c = 1; return json_encode($a); }", 0},
		{"copy", "<?php function f() { " + self + "$c = $a; return json_encode($a); }", 0},
		{"unmodelled slot alias", "<?php function f() { " + self + "$d['x'] = &$a; $d['x'] = 5; return json_encode($a); }", 0},
		{"write through slot alias", "<?php function f() { $d = []; " + self + "$d['x'] = &$a; $d['x'] = 5; return json_encode($a); }", 0},
		{"rebound slot", "<?php function f() { $c = []; " + pair + "$b['a'] = &$c; return json_encode($a); }", 0},
		{"global", "<?php function f() { " + self + "global $a; return json_encode($a); }", 0},
		{"dynamic variable", "<?php function f($n) { " + self + "$$n = [1]; return json_encode($a); }", 0},
		{"static call", "<?php " + self + "Registry::reset(); json_encode($a);", 0},
		{"constructor", "<?php " + self + "new Resetter(); json_encode($a);", 0},
		{"include", "<?php " + self + "include 'reset.php'; json_encode($a);", 0},
		{"eval", "<?php " + self + "eval('$a = [];'); json_encode($a);", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
