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

// BenchmarkTypeOfVariables types every variable of a file mixing guards,
// assertion calls, method-template calls and untyped properties (the paths
// narrowing walks for each use).
func BenchmarkTypeOfVariables(b *testing.B) {
	var sb strings.Builder
	sb.WriteString(`<?php
class Assert {
    /** @psalm-assert string $v */
    public static function string($v) {}
    /**
     * @template T
     * @param class-string<T> $c
     * @return T
     */
    public static function make(string $c) {}
}
class Foo { public ?int $id = null; }
final class Subject {
    private $count = 0;
    private $items = [];
`)
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, `    public function m%d(?string $s, int|string $u, array $xs) {
        $this->count++;
        $this->items[] = $s;
        Assert::string($u);
        $f = Assert::make(Foo::class);
        if ($s === null) { return $f->id; }
        foo($s, $u, $xs, $this->count, $this->items);
        $x = $s . $u;
        foreach ($xs as $k => $v) { bar($k, $v, $x, $f); }
        return $u;
    }
`, i)
	}
	sb.WriteString("}\n")
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Variable, *syntax.PropertyFetch, *syntax.MethodCall, *syntax.StaticCall:
				env.TypeOf(n.(syntax.Expr))
			}
			return true
		})
	}
}

// BenchmarkTypeOfElementWrites types every variable and element read of
// functions building arrays by element writes (whole-variable reads apply
// the reaching writes, element reads widen key by key).
func BenchmarkTypeOfElementWrites(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, `function f%d(array $rows, string $k) {
    $out = ['count' => 0, 'names' => []];
    $list = [];
    foreach ($rows as $i => $row) {
        $out['count']++;
        $out['names'][] = $row['name'];
        $list[] = $row;
        $list[$k] = $i;
        if ($out['count'] > 10) { break; }
    }
    $copy = $out;
    $n = $copy['count'] + count($list);
    [$first] = $list;
    foreach ($list as $key => $value) { use_it($key, $value, $first, $n); }
    return $out;
}
`, i)
	}
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Variable, *syntax.ArrayDimFetch:
				env.TypeOf(n.(syntax.Expr))
			}
			return true
		})
	}
}

// BenchmarkTypeOfCallables types every variable, property read and call of
// methods using closures, callbacks, out parameters, conditional return
// types and property writes.
func BenchmarkTypeOfCallables(b *testing.B) {
	var sb strings.Builder
	sb.WriteString(`<?php
class Foo { public function go(): void {} }
/** @return ($x is string ? int : float) */
function conv($x) {}
final class Subject {
    private ?Foo $foo = null;
`)
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, `    public function m%d(string $s, array $xs, ?array $opt) {
        $make = fn(int $n) => new Foo();
        $f = $make(1);
        $names = array_map(fn($x) => $x . 'a', $xs);
        $kept = array_filter([$f, null]);
        if (preg_match('/(a)/', $s, $m)) { use_it($m[1]); }
        $this->foo = new Foo();
        $g = $this->foo;
        $n = conv($s);
        $opt['k'] = $n;
        $cb = function () use ($g) { return $g; };
        return [$f, $names, $kept, $g, $opt, $cb()];
    }
`, i)
	}
	sb.WriteString("}\n")
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Variable, *syntax.PropertyFetch, *syntax.FuncCall, *syntax.MethodCall:
				env.TypeOf(n.(syntax.Expr))
			}
			return true
		})
	}
}

// BenchmarkTypeOfConditions types every variable of functions narrowed by
// compound guards, elseif chains, match arms and switch cases.
func BenchmarkTypeOfConditions(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php\nclass Foo { public ?Foo $next = null; public function ok(): bool { return true; } }\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, `function f%d(int|string|array|null $x, ?Foo $o, $c) {
    if (!is_scalar($x) && !$x instanceof \Stringable) { return; }
    if ($x === null || is_string($x)) { $a = $x; } elseif (is_int($x)) { $a = $x; } else { $a = $x; }
    $m = match (true) { is_string($x) => $x, $x === null => $x, default => $x };
    switch (true) {
        case is_int($x): $y = $x; break;
        case $o?->ok(): $y = $o; break;
        default: $y = $x;
    }
    if (gettype($x) === 'string' || in_array($x, [1, 2], true)) { $z = $x; }
    return $o !== null && $o->next !== null ? $o->next : $x;
}
`, i)
	}
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Variable, *syntax.PropertyFetch:
				env.TypeOf(n.(syntax.Expr))
			}
			return true
		})
	}
}

// BenchmarkTypeOfPipesAndGenerators includes parsing and index-time annotations.
func BenchmarkTypeOfPipesAndGenerators(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("<?php\nfunction source() { yield 'x' => 1; yield from [2, 3]; }\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, "function f%d() { $n = 'abc' |> strlen(...) |> (fn(int $n): int => $n + 1); foreach (source() as $k => $v) { useValue($v); } return $n; }\n", i)
	}
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP85}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		fs := index.Extract(f)
		infer.AnnotateReturns(f, fs, stubs.Index(), phpversion.PHP85)
		ix.Add(fs)
		env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP85)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if x, ok := n.(syntax.Expr); ok {
				env.TypeOf(x)
			}
			return true
		})
	}
}
