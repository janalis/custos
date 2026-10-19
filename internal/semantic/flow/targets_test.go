package flow

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// environmentAt is environment for a PHP target.
func environmentAt(t testing.TB, src string, version phpversion.Version) (*Env, *syntax.File) {
	t.Helper()
	f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	return New(f, infer.NewEnv(f, names.New(f), ix, version), nil), f
}

// The declarations a call may run are the one it resolves to when no other
// can stand for it, or that one and each replacement in a class extending
// the receiver's; the method of each copy when the class looked up is
// declared more than once and each copy declares it; none when another class
// on the way is declared twice, the receiver is not followed or the method
// is not the one reached.
func TestTargetsAreTheDeclarationsThatMayRun(t *testing.T) {
	const base = `class K { function take(string $s) {} static function make(string $s) {} private function hidden(string $s) {} } `
	for _, tc := range []struct {
		name, src string
		want      []string
	}{
		{"typed receiver, nothing replaces", base + `function f(K $k, $v) { $k->take($v); }`, []string{"K"}},
		{"typed receiver, a class replaces", base + `class L extends K { function take($s) {} } class M extends L {} function f(K $k, $v) { $k->take($v); }`, []string{"K", "L"}},
		{"typed receiver, two classes replace", base + `class L extends K { function take($s) {} } class M extends K { function take(string $s) {} } function f(K $k, $v) { $k->take($v); }`, []string{"K", "L", "M"}},
		{"typed receiver, a class extending it declared twice", base + `if ($c) { class L extends K {} } else { class L extends K { function take($s) {} } } function f(K $k, $v) { $k->take($v); }`, nil},
		{"object made here", base + `class L extends K { function take($s) {} } function f($v) { $k = new K; $k->take($v); }`, []string{"K"}},
		{"receiver assigned on one path", base + `function f(K $k, $v, $c) { if ($c) { $k = other(); } $k->take($v); }`, nil},
		{"receiver in a scope not followed", base + `function f(K $k, $v, $code) { $k->take($v); eval($code); }`, nil},
		{"receiver in code never reached", base + `function f(K $k, $v) { return; $k->take($v); }`, nil},
		{"receiver that is the object itself", `class K { function take(string $s) {} function f($v) { $this->take($v); } } class L extends K { function take($s) {} }`, []string{"K", "L"}},
		{"receiver in a trait", `trait T { function take(string $s) {} function f($v) { $this->take($v); } }`, nil},
		{"property receiver", base + `class H { private K $k; function f($v) { $this->k->take($v); } }`, []string{"K"}},
		{"static call naming the class", base + `class L extends K { static function make($s) {} } function f($v) { K::make($v); }`, []string{"K"}},
		{"static call bound late", `class K { static function make(string $s) {} function f($v) { static::make($v); } } class L extends K { static function make($s) {} }`, []string{"K", "L"}},
		{"class declared twice", `if ($c) { class K { function take(string $s) {} } } else { class K { function take($s) {} } } function f(K $k, $v) { $k->take($v); }`, []string{"K", "K"}},
		{"class declared twice, one copy inheriting", `class I { function take($s) {} } if ($c) { class K { function take(string $s) {} } } else { class K extends I { } } function f(K $k, $v) { $k->take($v); }`, nil},
		{"method not visible", base + `function f(K $k, $v) { $k->hidden($v); }`, nil},
		{"unknown method", base + `function f(K $k, $v) { $k->other($v); }`, nil},
		{"unknown receiver", base + `function f($k, $v) { $k->take($v); }`, nil},
		{"method of a PHP class", `function f(DateTime $k, $v) { $k->format($v); }`, []string{"DateTime"}},
		{"method of a PHP class a class of the project extends", `class D extends DateTime { function format($f): string { return ''; } } function f(DateTime $k, $v) { $k->format($v); }`, []string{"DateTime", "D"}},
		{"inherited method", `class J { function take(string $s) {} } class K extends J {} function f(K $k, $v) { $k->take($v); }`, []string{"J"}},
		{"method inherited through a class declared twice", `class I { function take(string $s) {} } if ($c) { class J extends I {} } else { class J extends I { function take($s) {} } } class K extends J {} function f(K $k, $v) { $k->take($v); }`, nil},
		{"method below a class that extends itself", `class J extends K { } class K extends J { } function f(K $k, $v) { $k->take($v); }`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			var call syntax.Node
			syntax.InspectFile(f, func(n syntax.Node) bool {
				switch n := n.(type) {
				case *syntax.MethodCall:
					call = n
				case *syntax.StaticCall:
					call = n
				}
				return true
			})
			var got []string
			for _, m := range e.Targets(call) {
				got = append(got, strings.TrimPrefix(m.Class, `\`))
			}
			// The classes extending the receiver's are visited in no fixed order.
			slices.Sort(got)
			slices.Sort(tc.want)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("targets: got %v, want %v", got, tc.want)
			}
		})
	}
}

// The class a walk starts from is not visited again when a class below it
// names it as its parent, however the receiver is written.
func TestClassesBelowAreVisitedOnce(t *testing.T) {
	e, _ := environment(t, `class J extends K { } class K extends J { }`, nil)
	for _, receiver := range []string{"K", `\K`} {
		var visited []string
		e.below(receiver, func(sub string) bool {
			visited = append(visited, sub)
			return false
		})
		if !slices.Equal(visited, []string{"J"}) {
			t.Errorf("below %s: visited %v", receiver, visited)
		}
	}
}

// A label written before a node lets a jump land before it; one written
// after it does not. A scope too large to follow is still read for its
// labels.
func TestLabelledScopes(t *testing.T) {
	many := strings.Repeat("if ($c) { other(); } ", MaxBlocks)
	for _, tc := range []struct {
		name, src string
		labelled  bool
	}{
		{"no label", `function f($c) { probe($c); }`, false},
		{"label before", `function f($c) { again: probe($c); if ($c) { goto again; } }`, true},
		{"label after", `function f($c) { probe($c); if ($c) { goto done; } other(); done: ; }`, false},
		{"label in a closure before", `function f($c) { $g = function () { x: goto x; }; probe($c); }`, false},
		{"label before, at file level", `again: probe($c); if ($c) { goto again; }`, true},
		{"label before, in a scope too large", `function f($c) { $g = function () { x: goto x; }; again: probe($c); ` + many + ` }`, true},
		{"label after, in a scope too large", `function f($c) { class_exists('K'); probe($c); ` + many + ` done: ; }`, false},
		{"no label, in a scope too large", `function f($c) { probe($c); ` + many + ` }`, false},
		{"label before, in a file too large", `again: probe($c); ` + many, true},
		// Only a label written after the scope first speaks of what is read.
		{"label before the first mention", `function f($a, $t) { again: try { $u = parse_url($a); probe($u); } catch (E $x) { if (++$t < 3) { goto again; } } }`, false},
		{"label between the mention and the read", `function f($a) { $u = parse_url($a); again: probe($u); goto again; }`, true},
		{"label before a copy of an earlier variable", `function f($a) { $raw = parse_url($a); again: $u = $raw; probe($u); goto again; }`, true},
		{"label before a copy through two variables", `function f($a) { $raw = parse_url($a); $z = 1; again: $mid = $z ? $raw : null; $u = $mid; probe($u); goto again; }`, true},
		{"label before, groups joined later", `function f($a) { $raw = parse_url($a); again: $u = 1; $mid = 2; $u = $mid; $mid = $raw; probe($u); goto again; }`, true},
		{"earlier variable only handed to a call", `function f() { $a = 1; again: $u = parse_url($a); probe($u); goto again; }`, false},
		{"earlier variable assigned in an argument", `function f() { $raw = other(); again: other($u = $raw); probe($u); goto again; }`, true},
		{"variable a closure takes", `function f() { $u = 1; $g = function () use (&$u) { again: goto again; }; top: probe($u); goto top; }`, true},
		{"closure made after the label", `function f() { top: $g = function () use (&$u) { $u = 1; }; probe($u); goto top; }`, false},
		{"nothing read", `function f() { $u = 1; again: probe(1); goto again; }`, true},
		{"the object read", `function f() { again: probe($this->p); goto again; }`, true},
		{"variable of a function in the node", `function f() { again: probe(function () { return $inner; }); goto again; }`, true},
		{"code evaluated before the label", `function f($c) { eval($c); again: $u = 1; probe($u); goto again; }`, true},
		{"code evaluated after the label", `function f() { again: eval('$u = 1;'); $u = 1; probe($u); goto again; }`, false},
		{"file included before the label", `function f() { include 'a.php'; again: $u = 1; probe($u); goto again; }`, true},
		{"extract before the label", `function f() { extract(other()); again: $u = 1; probe($u); goto again; }`, true},
		{"compact before the label", `function f() { $z = compact('raw'); again: $u = 1; probe($u); goto again; }`, true},
		{"parse_str filling the scope before the label", `function f() { parse_str('a=1'); again: $u = 1; probe($u); goto again; }`, true},
		{"parse_str filling an array before the label", `function f() { parse_str('a=1', $into); again: $u = 1; probe($u); goto again; }`, false},
		{"name chosen at run time before the label", `function f() { $$n = 1; again: $u = 1; probe($u); goto again; }`, true},
		{"$GLOBALS before the label", `function f() { $GLOBALS['u'] = 1; again: $u = 1; probe($u); goto again; }`, true},
		{"a function of the project before the label", `function mine() {} function f() { mine(); again: $u = 1; probe($u); goto again; }`, false},
		// Only what can hold the same value ties two variables.
		{"parameter read beside the variable", `function f($a, $pfx, $t) { again: try { $u = parse_url($a); echo $pfx . $u['host']; probe($u); } catch (E $x) { if (++$t < 3) { goto again; } } }`, false},
		{"parameter joined to an element of the variable", `function f($a, $pfx) { again: $u = parse_url($a); $h = $u['host']; $r = $pfx . $h; probe($u); goto again; }`, false},
		{"parameter embedded with the variable", `function f($a, $pfx) { again: $u = parse_url($a); $r = "$pfx$u"; probe($u); goto again; }`, false},
		{"parameter the result is called on", `function f($a, $p) { again: $u = $p->transform($a); probe($u); goto again; }`, false},
		{"parameter constructed with", `function f($a, $p) { again: $u = new K($p); probe($u); goto again; }`, false},
		{"parameter as an element key", `function f($a, $p) { again: $list = other(); $u = $list[$p]; probe($u); goto again; }`, false},
		{"parameter appended to the variable", `function f($a, $p) { again: $u = other(); $u .= $p; probe($u); goto again; }`, false},
		{"parameter iterated into the variable", `function f($p) { again: foreach ($p as $u) { probe($u); } goto again; }`, true},
		{"parameter iterated into a key", `function f($p) { again: foreach ($p as $u => $v) { probe($u); } goto again; }`, true},
		{"parameter the variable defaults to", `function f($p) { again: $u = null; $u ??= $p; probe($u); goto again; }`, true},
		{"parameter or another value", `function f($p) { again: $u = other() ?? $p; probe($u); goto again; }`, true},
		{"parameter bound by reference", `function f($p) { again: $u = &$p; probe($u); goto again; }`, true},
		{"parameter stored in an element", `function f($p) { again: $u = []; $u['k'] = $p; probe($u); goto again; }`, true},
		{"parameter a closure takes", `function f($p) { again: $u = function () use ($p) { return $p; }; probe($u); goto again; }`, true},
		{"parameter read in a function made in place", `function f($p) { again: $u = fn () => $p; probe($u); goto again; }`, true},
		{"parameter read in a class declared in place", `function f($p) { again: $u = [1, new class { function m($p) { return $p; } }]; probe($u); goto again; }`, false},
		{"variable written from one a parameter names", `function f($p) { again: $u = $$p; probe($u); goto again; }`, true},
		{"label in a function declared before", `function f() { top: $u = 1; function inner() { x: goto x; } class K { function m() { y: goto y; } } probe($u); goto top; }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			if got := e.Labelled(callsNamed(f, "probe")[0]); got != tc.labelled {
				t.Fatalf("labelled=%v", got)
			}
		})
	}
}

// A value checked in one place and used in another is asked about with the
// variables of the first and the labels up to the second.
func TestLabelledBeforeAUse(t *testing.T) {
	for src, want := range map[string]bool{
		`function f($a, $pfx, $c) { $r = realpath($a); if ($c) { goto at; } checked($r); at: probe($r . $pfx); }`:             true,
		`function f($a, $pfx, $t) { again: $r = realpath($a); checked($r); probe($r . $pfx); if (++$t < 3) { goto again; } }`: false,
	} {
		e, f := environment(t, src, nil)
		checked, use := argument(callsNamed(f, "checked")[0]), callsNamed(f, "probe")[0]
		if got := e.LabelledBefore(checked, use); got != want {
			t.Errorf("%s: labelled=%v", src, got)
		}
		if !e.Labelled(use) {
			t.Errorf("%s: the use reads a parameter", src)
		}
	}
}

// A scope of jumps between many labels is swept, which settles it in few
// visits, and stops at a fraction of the time budget when it still does not.
func TestJumpsBetweenLabelsStayBounded(t *testing.T) {
	jumps := func(labels, locals int) string {
		var src strings.Builder
		src.WriteString("function f($c) {")
		for i := range locals {
			fmt.Fprintf(&src, "$v%d = $_GET['a%d'] ?? 'l';", i, i)
		}
		for i := range labels {
			fmt.Fprintf(&src, "l%d: if ($c > %d) { $v%d .= $v%d; goto l%d; }", i, i, i%locals, (i*7+3)%locals, (i*5+1)%max(i, 1))
		}
		src.WriteString("probe($v0); }")
		return src.String()
	}
	e, f := environment(t, jumps(12, 6), nil)
	if r := e.scope(f.Stmts[0].(*syntax.Function)); !r.complete || !e.Tainted(argument(callsNamed(f, "probe")[0]), Shell) {
		t.Fatalf("complete=%v copies=%d", r.complete, r.copies)
	}
	e, f = environment(t, jumps(300, 150), nil)
	if r := e.scope(f.Stmts[0].(*syntax.Function)); r.complete || r.copies > MaxCopied/jumpShare+MaxCopied/64 {
		t.Fatalf("complete=%v copies=%d", r.complete, r.copies)
	}
}

// What a PHP version allows decides what a step and a constant can yield.
func TestObjectsOfOlderTargets(t *testing.T) {
	const src = `function probe($x) {} function t($a, $o) { static $u; $u = parse_url($a); if ($u === false) { return; } %s probe($u); }`
	for _, tc := range []struct {
		body    string
		version phpversion.Version
		kept    bool
	}{
		// Before PHP 8.3 a step leaves an object as it is.
		{`$o++; echo $o;`, phpversion.PHP84, true},
		{`$o++; echo $o;`, phpversion.PHP74, false},
		{`$n = 1; $n++; echo $n;`, phpversion.PHP74, true},
		{`echo ++$o;`, phpversion.PHP74, false},
		{`echo $o--;`, phpversion.PHP82, false},
		{`echo $o--;`, phpversion.PHP83, true},
		// Since PHP 8.1 a constant of the project may hold an object.
		{`echo SOME_CONSTANT;`, phpversion.PHP80, true},
		{`echo SOME_CONSTANT;`, phpversion.PHP81, false},
		{`echo PHP_EOL, true, null;`, phpversion.PHP84, true},
	} {
		t.Run(fmt.Sprintf("%s at %v", tc.body, tc.version), func(t *testing.T) {
			e, f := environmentAt(t, fmt.Sprintf(src, tc.body), tc.version)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
	// A static variable stepped from another one holds what that one held.
	for version, kept := range map[phpversion.Version]bool{phpversion.PHP74: false, phpversion.PHP84: true} {
		e, f := environmentAt(t, `function probe($x) {} function t($a, $o) { static $u; $u = parse_url($a); again(); if ($u === false) { return; } echo $u['host']; probe($u); $u = $o++; }`, version)
		if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != kept {
			t.Errorf("non-false=%v at %v", got, version)
		}
	}
}

// A call or a construction spares the request tables when what it runs is
// known and stores nothing in them, itself or through what it calls.
func TestCallsThatSpareTheRequestTables(t *testing.T) {
	const declared = `
		function log_it($m) { error_log($m); }
		function fill() { $_GET['a'] = $_POST['a']; }
		function read_it() { return $_SERVER['X']; }
		function table() { return $GLOBALS; }
		function evaluated($c) { eval($c); }
		function included($c) { include $c; }
		function through() { log_it('x'); }
		function through_fill() { fill(); }
		function again($n) { again($n - 1); }
		function by_reference(&$a) { }
		function maker() { return function () { $_GET['a'] = 1; }; }
		function outer() { function inner() { $_GET['a'] = 1; } }
		function nested_loose(K $k) { return function () use ($k) { $k->open(); }; }
		function nested_named() { return function () { K::stat(); }; }
		function nested_php() { return function () { return strlen('a'); }; }
		function calls_method(K $k) { return $k->quiet(); }
		if ($c) { function twice() { } }
		function twice() { }
		class K {
			function open() { return 1; }
			final function quiet() { return 1; }
			final function loud() { $_POST['a'] = 1; }
			final function byRef(&$x) { }
			static function stat() { return 2; }
			function made() { return new static(); }
		}
		interface I { function go(); }
		final class Impl implements I { function go() { return 1; } }
		interface J { function go(); }
		class Plain { }
		class Loud { function __construct() { $_GET['a'] = 1; } }
		class Logs { function __construct() { error_log('x'); } }
		class Dropped { function __destruct() { $_GET['a'] = 1; } }
		class Old { function Old() { $_GET['a'] = 1; } }
		class P { function __construct() { } }
		if ($c) { class Q extends P { } } else { class Q extends P { function __construct() { $_GET['a'] = 1; } } }
		class R extends Q { }
		if ($c) { class Two { } } else { class Two { } }
	`
	for _, tc := range []struct {
		name, call string
		version    phpversion.Version
		spared     bool
	}{
		{"PHP function", `strlen('a');`, phpversion.Max, true},
		{"PHP function given a function", `array_map(function ($v) { return $v; }, $a);`, phpversion.Max, false},
		{"PHP function given the name of one that computes", `array_map('intval', $a);`, phpversion.Max, true},
		{"PHP function that loads a class", `class_exists('K');`, phpversion.Max, false},
		{"PHP function given more than it declares", `strlen('a', 'b');`, phpversion.Max, false},
		{"PHP function given unpacked arguments", `strlen(...$a);`, phpversion.Max, false},
		{"function chosen at run time", `$a();`, phpversion.Max, false},
		{"function that only calls PHP", `log_it('x');`, phpversion.Max, true},
		{"function that writes a table", `fill();`, phpversion.Max, false},
		{"function that reads a table", `read_it();`, phpversion.Max, false},
		{"function that names the globals table", `table();`, phpversion.Max, false},
		{"function that evaluates code", `evaluated($a);`, phpversion.Max, false},
		{"function that includes a file", `included($a);`, phpversion.Max, false},
		{"function that calls one of the project", `through();`, phpversion.Max, false},
		{"function that calls one that writes", `through_fill();`, phpversion.Max, false},
		{"function that calls itself", `again(3);`, phpversion.Max, false},
		{"function taking a reference", `by_reference($a);`, phpversion.Max, false},
		{"function making a function that writes", `maker();`, phpversion.Max, false},
		{"function declaring one that writes", `outer();`, phpversion.Max, true},
		{"function whose closure calls an open method", `nested_loose($k);`, phpversion.Max, false},
		{"function whose closure calls a method of the project", `nested_named();`, phpversion.Max, false},
		{"function whose closure calls PHP", `nested_php();`, phpversion.Max, true},
		{"function calling a final method of the project", `calls_method($k);`, phpversion.Max, false},
		{"function declared twice", `twice();`, phpversion.Max, false},
		{"final method", `$k->quiet();`, phpversion.Max, true},
		{"final method that writes", `$k->loud();`, phpversion.Max, false},
		{"final method taking a reference", `$k->byRef($a);`, phpversion.Max, false},
		{"static method", `K::stat();`, phpversion.Max, true},
		{"method of an unknown receiver", `$a->quiet();`, phpversion.Max, false},
		{"method every implementer declares sparing", `$i->go();`, phpversion.Max, true},
		{"method nothing implements", `$j->go();`, phpversion.Max, false},
		{"method of a PHP class", `$d->format('Y');`, phpversion.Max, true},
		{"method of a PHP class given a function", `$o->uasort(function ($x, $y) { return 0; });`, phpversion.Max, false},
		{"method that invokes a function", `$f->call($k);`, phpversion.Max, false},
		{"method that resumes a generator", `$g->next();`, phpversion.Max, false},
		{"method of the reflection classes", `$r->invoke();`, phpversion.Max, false},
		{"PHP class constructed", `new DateTime();`, phpversion.Max, true},
		{"PHP class constructed without arguments", `new DateTime;`, phpversion.Max, true},
		{"PHP class constructed with a function", `new CallbackFilterIterator($it, function () { return true; });`, phpversion.Max, false},
		{"class without constructor", `new Plain();`, phpversion.Max, true},
		{"constructor that writes", `new Loud();`, phpversion.Max, false},
		{"constructor that only calls PHP", `new Logs();`, phpversion.Max, true},
		{"destructor that writes", `new Dropped();`, phpversion.Max, false},
		{"method named like its class", `new Old();`, phpversion.Max, true},
		{"method named like its class before PHP 8", `new Old();`, phpversion.PHP74, false},
		{"constructor found through a class declared twice", `new R();`, phpversion.Max, false},
		{"class declared twice", `new Two();`, phpversion.Max, false},
		{"unknown class", `new Missing();`, phpversion.Max, false},
		{"class chosen at run time", `new $a();`, phpversion.Max, false},
		{"class bound late", `$k->made();`, phpversion.Max, false},
		{"class declared inline", `new class { };`, phpversion.Max, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := declared + `function t($a, K $k, I $i, J $j, DateTime $d, ArrayObject $o, Closure $f, Generator $g, ReflectionFunction $r, Iterator $it) { ` + tc.call + ` }`
			f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %v", f.Errors)
			}
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			types := infer.NewEnv(f, names.New(f), ix, tc.version)
			_, own := Summarize(f, types, NewSnapshot())
			snapshot := NewSnapshot(own)
			e := New(f, types, snapshot)
			body := f.Stmts[len(f.Stmts)-1].(*syntax.Function).Body
			call := body.Stmts[0].(*syntax.ExprStmt).Expr
			if got := e.Spares(call); got != tc.spared {
				t.Fatalf("spares=%v", got)
			}
			if e.Spares(body) {
				t.Fatal("a block is no call")
			}
		})
	}
}

// What was established of a static variable holds across code that runs
// nothing of the project, and is gone past code that may.
func TestStaticGuardAcrossValuesAndCallsOfPHP(t *testing.T) {
	const src = `function probe($x) {}
		interface Named { const IC = 'i'; }
		class Other { const X = 'x'; }
		final class K implements Named {
			const C = 'c'; const D = Other::X; const E = self::C; const F = [self::C, 'a', self::E]; const G = self::D . 'x'; const H = self::C . self::MISSING; const I = self::IC;
			const L0 = 1; const L1 = self::L0; const L2 = self::L1; const L3 = self::L2; const L4 = self::L3; const L5 = self::L4; const L6 = self::L5; const L7 = self::L6; const L8 = self::L7; const L9 = self::L8;
			private string $text = 'x'; private array $list = []; private $any; private ?K $next = null;
			function t($a, $o, int $i, string $s, array $l, array $xs, $k) { static $u; $u = parse_url($a); if ($u === false) { return; } %s probe($u); }
		}`
	for _, tc := range []struct {
		body string
		kept bool
	}{
		// A destructuring of an array that holds no object fills its
		// variables with none.
		{`list($t, $w) = explode(',', $s); $t = $w;`, true},
		{`[$t, [$w, $x]] = [$i, [$s, $i]]; $w = 1; $x = 2;`, true},
		{`[, $t] = [$i, $s]; $t = 1;`, true},
		{`[$t, $w] = $l; $t = 1;`, false},
		{`[&$t, $w] = $xs; $t = 1;`, false},
		// The keys of an array, and what JSON text decodes to as arrays.
		{`$t = array_keys($l); $t = [];`, true},
		{`$t = array_flip($l); $t = [];`, false},
		{`$t = array_flip(explode(',', $s)); $t = [];`, true},
		{`$t = json_decode($s, true); $t = null;`, true},
		{`$t = json_decode($s, associative: true); $t = null;`, true},
		{`$t = json_decode($s); $t = null;`, false},
		{`$t = json_decode($s, false); $t = null;`, false},
		{`$t = json_decode(...$l); $t = null;`, false},
		{`$t = json_decode(); $t = null;`, false},
		{`extract();`, true},
		{`$t = $o->array_keys($l); $t = [];`, false},
		// A function PHP calls runs code of the project unless it is one of
		// its own that only computes, given nothing to convert.
		{`$t = array_filter(explode(',', $s)); $t = [];`, true},
		{`$t = array_filter(explode(',', $s), null); $t = [];`, true},
		{`$t = array_map('intval', explode(',', $s)); $t = [];`, true},
		{`$t = array_map('intval', $l);`, false},
		{`$t = array_map('strval', $l);`, false},
		{`$t = array_map('class_exists', explode(',', $s));`, false},
		{`$t = array_map('usort', [explode(',', $s)], [$s]);`, false},
		{`$t = array_map($k, explode(',', $s));`, false},
		{`usort($xs, $k);`, false},
		{`$t = new CallbackFilterIterator;`, true},
		{`$t = array_map(...$l);`, false},
		{`$t = array_filter(explode(',', $s), 'intval', 0, 1);`, false},
		// Functions of PHP that run code of the project by themselves.
		{`session_start();`, false},
		{`gc_collect_cycles();`, false},
		{`array_shift($xs);`, false},
		{`array_pop($xs);`, false},
		{`$t = array_shift($xs);`, true},
		{`$t = explode(',', $s); array_shift($t);`, true},
		// extract() of an array written with its keys assigns each of them.
		{`extract(['o' => 1]);`, false},
		{`extract(['i' => 1]);`, true},
		{`$t = [1]; extract(['w' => $t]); $w = 2;`, true},
		// A constant whose value names only constants of its own class.
		{`$x = self::C;`, true},
		{`$x = self::D;`, false},
		{`$x = self::E;`, true},
		{`$x = self::F;`, true},
		{`$x = self::G;`, false},
		{`$x = self::H;`, false},
		{`$x = self::I;`, false},
		{`$x = self::L7;`, true},
		{`$x = self::L9;`, false},
		// A property read as declared holds what its type admits.
		{`echo $this->text;`, true},
		{`echo $this->any;`, false},
		{`echo $this->next;`, false},
		{`$x = $this->text; $x = 1;`, true},
		{`$x = $this->list; $x = 1;`, false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			e, f := environmentAt(t, fmt.Sprintf(src, tc.body), phpversion.PHP84)
			calls := callsNamed(f, "probe")
			if got := e.Excludes(argument(calls[len(calls)-1]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
	// extract() given anything else may write any variable.
	for _, body := range []string{`extract($l);`, `extract(['o' => 1], EXTR_OVERWRITE);`, `extract(...$l);`, `extract(array: ['o' => 1]);`, `extract([$s => 1]);`, `extract(['o' => &$i]);`, `extract([...$l]);`, `extract([1]);`, `extract(['_GET' => 1]);`, `extract(['this' => 1]);`, `extract(['GLOBALS' => 1]);`} {
		e, f := environmentAt(t, fmt.Sprintf(src, `$w = 'x'; `+body+` other($w);`), phpversion.PHP84)
		if _, known := e.Resolve(argument(callsNamed(f, "other")[0])); known {
			t.Errorf("%s: the variable kept its value", body)
		}
	}
	e, f := environmentAt(t, fmt.Sprintf(src, `$w = 'x'; extract(['o' => 1]); other($w);`), phpversion.PHP84)
	if _, known := e.Resolve(argument(callsNamed(f, "other")[0])); !known {
		t.Error("a variable the array does not name lost its value")
	}
}
