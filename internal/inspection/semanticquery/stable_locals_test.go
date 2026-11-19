package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// probeStable runs check on the first argument of every strrev() and sort()
// call: builtins, so the probe itself is not code able to run application
// logic.
func probeStable(t *testing.T, src string, check func(*analysis.Context, syntax.Expr)) {
	t.Helper()
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && (name.Value == "strrev" || name.Value == "sort") {
			ran = true
			check(ctx, c.Args.Args[0].(*syntax.Arg).Value)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %q: %v", src, f.Errors)
	}
	e.Analyze(f)
	if !ran {
		t.Fatalf("no probe in %q", src)
	}
}

func TestStableLocalValue(t *testing.T) {
	const fn = "function f($p, &$r) { "
	shared := LocalStability{SharedOpaque: true}
	copies := LocalStability{CopiesOnly: true}
	own := LocalStability{TrustOwnCaptures: true}
	for _, tc := range []struct {
		name string
		src  string
		opts LocalStability
		want bool
	}{
		{"literal", fn + "strrev('x'); }", LocalStability{}, true},
		{"plain local", fn + "$a = 'x'; strrev($a); }", LocalStability{}, true},
		{"dynamic read only", fn + "$a = 'x'; strrev($$a); }", LocalStability{}, true},
		{"increment", fn + "$a = 'x'; $a++; strrev($a); }", LocalStability{}, false},
		{"element increment", fn + "$a = ['x']; $a[0]++; strrev($a); }", LocalStability{}, false},
		{"property increment", fn + "$a = new S; $a->n++; strrev($a); }", LocalStability{}, false},
		{"dynamic increment", fn + "$a = 'x'; $n = 'a'; $$n++; strrev($a); }", LocalStability{}, false},
		{"dynamic assignment", fn + "$a = 'x'; $n = 'a'; $$n = 'y'; strrev($a); }", LocalStability{}, false},
		{"dynamic destructuring", fn + "$a = 'x'; $n = 'a'; [$$n] = ['y']; strrev($a); }", LocalStability{}, false},
		{"destructuring feeds", fn + "$s = 'x'; $s++; [$a, [$b]] = [$s, [1]]; list($c) = [2]; strrev($a); }", LocalStability{}, false},
		{"computed feed", fn + "$s = 1; $s++; $a = $s + 1; strrev($a); }", LocalStability{}, false},
		{"computed feed, copies only", fn + "$s = 1; $s++; $a = $s + 1; strrev($a); }", copies, true},
		{"copy, copies only", fn + "$s = 1; $s++; $a = ($s); strrev($a); }", copies, false},
		{"copy cycle", fn + "$a = 'x'; $b = $a; $a = $b; strrev($a); }", LocalStability{}, true},
		{"reference assignment target", fn + "$a = 'x'; $a = &$o; strrev($a); }", LocalStability{}, false},
		{"reference assignment source", fn + "$a = 'x'; $o = &$a; strrev($a); }", LocalStability{}, false},
		{"foreach value", fn + "$a = 'x'; foreach ($p as $k => $a) {} strrev($a); }", LocalStability{}, false},
		{"foreach by reference", fn + "$a = ['x']; foreach ($a as &$v) {} strrev($a); }", LocalStability{}, false},
		{"foreach by value", fn + "$a = ['x']; foreach ($a as $v) {} strrev($a); }", LocalStability{}, true},
		{"array item by reference", fn + "$a = 'x'; $l = [&$a]; strrev($a); }", LocalStability{}, false},
		{"catch", fn + "$a = 'x'; try {} catch (E $a) {} strrev($a); }", LocalStability{}, false},
		{"catch without variable", fn + "$a = 'x'; try {} catch (E) {} strrev($a); }", LocalStability{}, true},
		{"nested capture", fn + "$a = 'x'; $c = function () use (&$a) {}; strrev($a); }", LocalStability{}, false},
		{"nested capture by value", fn + "$a = 'x'; $c = function () use ($a) {}; strrev($a); }", LocalStability{}, true},
		{"global", fn + "global $a; $a = 'x'; strrev($a); }", LocalStability{}, false},
		{"static", fn + "static $a = 'x'; strrev($a); }", LocalStability{}, false},
		{"global, shared without code", fn + "global $a; $a = 'x'; strrev($a); }", shared, true},
		{"global, shared with code", fn + "global $a; $a = 'x'; g(); strrev($a); }", shared, false},
		{"static, shared with method", fn + "static $a = 'x'; $o->m(); strrev($a); }", shared, false},
		{"reference parameter, shared with new", fn + "$r = 'x'; new S; strrev($r); }", shared, false},
		{"plain local, shared with code", fn + "$a = 'x'; S::m(); strrev($a); }", shared, true},
		{"shared with closure", fn + "global $a; $c = function () {}; strrev($a); }", shared, false},
		{"shared with arrow", fn + "global $a; $c = fn () => 1; strrev($a); }", shared, false},
		{"include", fn + "$a = 'x'; include 'y.php'; strrev($a); }", LocalStability{}, false},
		{"eval", fn + "$a = 'x'; eval('1;'); strrev($a); }", LocalStability{}, false},
		{"extract", fn + "$a = 'x'; extract($p); strrev($a); }", LocalStability{}, false},
		{"namespaced extract", "namespace N; function extract($x) {} " + fn + "$a = 'x'; extract($p); strrev($a); }", LocalStability{}, true},
		{"parse_str one argument", fn + "$a = 'x'; parse_str($p); strrev($a); }", LocalStability{}, false},
		{"parse_str result", fn + "$a = 'x'; parse_str($p, $o); strrev($a); }", LocalStability{}, true},
		{"parse_str into the local", fn + "$a = 'x'; parse_str($p, $a); strrev($a); }", LocalStability{}, false},
		{"call-time reference", fn + "$a = 'x'; g(&$a); strrev($a); }", LocalStability{}, false},
		{"element by reference", fn + "$a = ['x']; settype($a[0], 'int'); strrev($a); }", LocalStability{}, false},
		{"unresolved callee", fn + "$a = 'x'; g($a); strrev($a); }", LocalStability{}, false},
		{"unresolved callee, element", fn + "$a = ['x']; g($a[0]); strrev($a); }", LocalStability{}, false},
		{"unresolved callee, unpacked", fn + "$a = ['x']; g(...$a); strrev($a); }", LocalStability{}, true},
		{"unresolved callee, other local", fn + "$a = 'x'; g($b); strrev($a); }", LocalStability{}, true},
		{"by-value argument", fn + "$a = 'x'; strlen($a); strrev($a); }", LocalStability{}, true},
		{"own argument", fn + "$a = ['x']; sort($a); }", LocalStability{}, true},
		{"first-class callable", fn + "$a = 'x'; $c = strlen(...); strrev($a); }", LocalStability{}, true},
		{"globals entry", fn + "$a = 'x'; $GLOBALS['a'] = 1; strrev($a); }", LocalStability{}, false},
		{"other globals entry", fn + "$a = 'x'; $GLOBALS['b'] = 1; strrev($a); }", LocalStability{}, true},
		{"unknown globals entry", fn + "$a = 'x'; $GLOBALS[$p] = 1; strrev($a); }", LocalStability{}, false},
		{"globals", fn + "$a = 'x'; $g = $GLOBALS; strrev($a); }", LocalStability{}, false},
		{"nested scopes", fn + "$a = 'x'; function g() { $a++; } $o = new class ($p) { function m() { $a++; } }; $c = function () { $a++; }; $d = fn () => $a++; strrev($a); }", LocalStability{}, true},
		{"anonymous class arguments", fn + "$a = 'x'; $o = new class ($a = &$p) {}; strrev($a); }", LocalStability{}, false},
		{"top level", "$a = 'x'; strrev($a);", LocalStability{}, true},
		{"top level global elsewhere", "function g() { global $a; } $a = 'x'; strrev($a);", LocalStability{}, false},
		{"top level globals elsewhere", "function g() { $GLOBALS[$n] = 1; } $a = 'x'; strrev($a);", LocalStability{}, false},
		{"top level, shared with code", "$a = 'x'; g(); strrev($a);", shared, false},
		{"top level, shared global elsewhere", "function g() { global $a; } $a = 'x'; strrev($a);", shared, true},
		{"own capture", "$c = function () use (&$a) { $a = 'x'; strrev($a); };", LocalStability{}, false},
		{"own capture, trusted", "$c = function () use (&$a) { $a = 'x'; strrev($a); };", own, true},
		{"own capture, shared without code", "$c = function () use (&$a) { $a = 'x'; strrev($a); };", shared, true},
		{"arrow reads the enclosing scope", fn + "$a = 'x'; $a++; $c = fn () => strrev($a); }", LocalStability{}, false},
		{"arrow over a stable local", fn + "$a = 'x'; $c = fn () => strrev($a); }", LocalStability{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probeStable(t, tc.src, func(ctx *analysis.Context, e syntax.Expr) {
				if got := StableLocalValue(ctx, e, tc.opts); got != tc.want {
					t.Fatalf("StableLocalValue(%s) = %v, want %v", ctx.Text(e), got, tc.want)
				}
			})
		})
	}
	probeStable(t, "strrev('x');", func(ctx *analysis.Context, _ syntax.Expr) {
		if !StableLocalValue(ctx, nil, LocalStability{}) {
			t.Fatal("an absent expression reads no local")
		}
	})
}

// BenchmarkStableLocalValue asks about every probed local of a long scope:
// the scope's facts are gathered once per file and options, not per call.
func BenchmarkStableLocalValue(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php function f($p) {\n")
	for i := range 300 {
		fmt.Fprintf(&src, "  $a%d = 'x' . $p; $b%d = $a%d; g($b%d, $p); strrev($a%d);\n", i, i, i, i, i)
	}
	src.WriteString("}\n")
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "strrev" {
			StableLocalValue(ctx, c.Args.Args[0].(*syntax.Arg).Value, LocalStability{})
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	f := []byte(src.String())
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", f, syntax.Options{}))
	}
}
