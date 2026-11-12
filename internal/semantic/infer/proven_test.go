package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// provenProbe returns the native and the proven type of the second argument
// of the `t('v', …)` call in src.
func provenProbe(t *testing.T, src string) (native, proven string) {
	t.Helper()
	// The probe takes its argument by value: it must not count as a write.
	src += "\nfunction t($label, $value) {}"
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	found := false
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" {
				x := c.Args.Args[1].(*syntax.Arg).Value
				native, proven, found = env.Native().TypeOf(x).String(), env.ProvenTypeOf(x).String(), true
			}
		}
		return true
	})
	if !found {
		t.Fatal("no t() probe")
	}
	return native, proven
}

// TestProvenTypes is the audit of the native types: for each way a local
// may change without a definition showing it, the native type (before) and
// the proven one (after). A known before type that misses the runtime value
// is what ProvenTypeOf exists for.
func TestProvenTypes(t *testing.T) {
	const unknown = "?unknown"
	cases := []struct {
		name, src      string
		native, proven string
	}{
		// By-reference writes.
		{"byref param of a known callee", `<?php function set(&$x) { $x = 'a'; } function f() { $a = 1; set($a); t('v', $a); }`, unknown, unknown},
		{"element to a by-ref param", `<?php function set(&$x) { $x = 'a'; } function f() { $a = [1]; set($a[0]); t('v', $a); }`, "int[]", unknown},
		{"unresolved function", `<?php function f() { $a = 1; unknown_fn($a); t('v', $a); }`, "int", unknown},
		{"callable variable", `<?php function f(callable $g) { $a = 1; $g($a); t('v', $a); }`, "int", unknown},
		{"method on unknown receiver", `<?php function f($o) { $a = 1; $o->m($a); t('v', $a); }`, "int", unknown},
		{"static call on a class string", `<?php function f(string $c) { $a = 1; $c::m($a); t('v', $a); }`, "int", unknown},
		{"new of an unknown class", `<?php function f() { $a = 1; new Unknown($a); t('v', $a); }`, "int", unknown},
		{"reference assignment", `<?php function f() { $a = 1; $r = &$a; $r = 'x'; t('v', $a); }`, "int|string", unknown},
		{"by-ref foreach leaves a reference", `<?php function f() { $l = [1, 2]; foreach ($l as &$v) {} $v = 'x'; t('v', $l); }`, "int[]", unknown},
		{"by-ref parameter after a call", `<?php function f(&$x) { $x = 1; g(); t('v', $x); }`, "int", unknown},
		// By-reference closure captures.
		{"capture created before the definition", `<?php function f() { $g = function () use (&$a) { $a = 'x'; }; $a = 1; $g(); t('v', $a); }`, "int", unknown},
		{"capture called by a builtin", `<?php function f() { $g = function () use (&$a) { $a = 'x'; }; $a = 1; array_map($g, [1]); t('v', $a); }`, "int", unknown},
		{"capture created after the definition", `<?php function f() { $a = 1; $g = function () use (&$a) { $a = 'x'; }; $g(); t('v', $a); }`, unknown, unknown},
		// global.
		{"global after a call", `<?php function f() { global $g; $g = 1; other(); t('v', $g); } function other() { global $g; $g = 'x'; }`, "int", unknown},
		{"global written through $GLOBALS", `<?php function f() { global $g; $g = 1; $GLOBALS['g'] = 'x'; t('v', $g); }`, "int", unknown},
		{"top-level variable after a call", `<?php $g = 1; other(); t('v', $g); function other() { global $g; $g = 'x'; }`, "int", unknown},
		{"top-level variable through $GLOBALS", `<?php $g = 1; $GLOBALS['g'] = 'x'; t('v', $g);`, "int", unknown},
		// static.
		{"static after a recursive call", `<?php function f() { static $n = 0; $n = 1; f(); t('v', $n); }`, "int", unknown},
		{"static read at entry", `<?php function f() { static $n = 0; t('v', $n); $n = 'x'; }`, "int|string", "int|string"},
		// extract, $$, eval, include: unknown in both.
		{"extract", `<?php function f(array $arr) { $a = 1; extract($arr); t('v', $a); }`, unknown, unknown},
		{"variable variable", `<?php function f(string $name) { $a = 1; $$name = 'x'; t('v', $a); }`, unknown, unknown},
		{"include", `<?php function f() { $a = 1; include 'x.php'; t('v', $a); }`, unknown, unknown},
		// Property re-reads after calls: native types keep the declaration.
		{"typed property after a call", `<?php class C { public int|string $p = 0; function m() { $this->p = 1; $this->o(); t('v', $this->p); } function o() { $this->p = 'x'; } }`, "int|string", "int|string"},
		{"narrowed property after a call", `<?php class C { public ?C $p = null; function m() { if ($this->p !== null) { $this->reset(); t('v', $this->p); } } function reset() { $this->p = null; } }`, `\C|null`, `\C|null`},
		{"untyped property after a call", `<?php class C { public $p; function m() { $this->p = 1; $this->o(); t('v', $this->p); } function o() { $this->p = 'x'; } }`, unknown, unknown},
		{"static property after a call", `<?php class C { public static ?C $p = null; function m() { if (self::$p !== null) { $this->o(); t('v', self::$p); } } function o() {} }`, `\C|null`, `\C|null`},
		// $this in a closure or an arrow function: Closure::bind, bindTo and
		// call may run it bound to another object.
		{"$this in a closure", `<?php class C { function m() { return function () { t('v', $this); }; } }`, `\C`, unknown},
		{"$this in an arrow function", `<?php class C { function m() { return fn() => t('v', $this); } }`, `\C`, unknown},
		{"property of $this in a closure", `<?php class C { public int $p = 0; function m() { return function () { t('v', $this->p); }; } }`, "int", unknown},
		{"narrowed property of $this in a closure", `<?php class C { public ?C $p = null; function m() { return function () { if ($this->p !== null) { t('v', $this->p); } }; } }`, `\C`, unknown},
		{"method of $this in an arrow function", `<?php class C { function n(): int { return 1; } function m() { return fn() => t('v', $this->n()); } }`, "int", unknown},
		{"$this in a static closure", `<?php class C { function m() { return static function () { t('v', $this); }; } }`, `\C`, unknown},
		{"$this in a method of a class in a closure", `<?php function f() { return function () { return new class { public int $p = 0; function m() { t('v', $this->p); } }; }; }`, "int", "int"},
		// Precision kept.
		{"$this in a method", `<?php class C { function m() { t('v', $this); } }`, `\C`, `\C`},
		{"plain local after a call", `<?php function f() { $a = 1; g(); t('v', $a); }`, "int", "int"},
		{"by-value callee", `<?php function g(int $x) {} function f() { $a = 1; g($a); t('v', $a); }`, "int", "int"},
		{"top-level variable after a plain builtin", `<?php $a = 1; strlen('x'); t('v', $a);`, "int", "int"},
		{"static read without a call", `<?php function f() { static $n = null; if ($n === null) { $n = 1; } t('v', $n); }`, "int", "int"},
		{"by-value capture", `<?php function f() { $a = 1; $g = function () use ($a) { t('v', $a); }; }`, "int", "int"},
		{"loop-carried local", `<?php function f() { $a = 1; while (g()) { t('v', $a); $a = 2; } }`, "int", "int"},
		{"loop-carried global", `<?php function f() { global $a; $a = 1; while (true) { t('v', $a); $a = 2; g(); } }`, "int", unknown},
	}
	for _, c := range cases {
		native, proven := provenProbe(t, c.src)
		if native != c.native || proven != c.proven {
			t.Errorf("%s: native %s proven %s, want %s and %s", c.name, native, proven, c.native, c.proven)
		}
	}
}

// BenchmarkProvenTypeOf types every variable read of a file with aliased
// locals and calls, natively and proven (the extra mutation scans).
func BenchmarkProvenTypeOf(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	for i := range 200 {
		fmt.Fprintf(&sb, `function f%d(array $xs, &$out) {
    global $g;
    static $n = 0;
    $a = 1; $r = &$a; $g = $a;
    $cb = function () use (&$a) { $a++; };
    foreach ($xs as $k => $v) { helper($k, $v); $n += $a; strlen($v); }
    $cb();
    $out = $a + $n + $g;
    unknown_fn($xs);
    return [$a, $r, $n, $g, $xs, $out];
}
`, i)
	}
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	for _, mode := range []string{"native", "proven"} {
		b.Run(mode, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				f := syntax.Parse("t.php", src, opt)
				ix := index.New(stubs.Index())
				ix.Add(index.Extract(f))
				env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
				if mode == "proven" {
					env = env.Proven()
				}
				syntax.InspectFile(f, func(n syntax.Node) bool {
					if v, ok := n.(*syntax.Variable); ok {
						env.TypeOf(v)
					}
					return true
				})
			}
		})
	}
}

func TestProvenEnv(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php $a = 1; strlen('x'); $a; $a;`), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	p := env.Proven()
	if !p.IsProven() || !p.IsNative() || env.IsProven() || env.Native().IsProven() {
		t.Fatal("Proven flags")
	}
	if p.Proven() != p || env.Native().Proven() != p || p.Native() != p {
		t.Fatal("Proven is cached per native Env")
	}
	if w := p.WithOptions(infer.Options{Assertions: true}); !w.IsProven() {
		t.Fatal("WithOptions keeps the proven mode")
	}
	// Two reads of one scope share the calls that may run project code.
	var reads []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && strings.EqualFold(v.Name, "a") {
			reads = append(reads, v)
		}
		return true
	})
	for _, r := range reads[1:] {
		if got := p.TypeOf(r).String(); got != "int" {
			t.Fatalf("read: %s", got)
		}
	}
}
