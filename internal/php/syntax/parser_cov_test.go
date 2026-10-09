package syntax

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	phpversion "custos/internal/php/version"
)

// phpLint reports whether the given PHP version accepts src (php -l with
// that version's binary); ok is false when the binary is not installed.
func phpLint(t *testing.T, v phpversion.Version, src string) (valid, ok bool) {
	t.Helper()
	bin := filepath.Join("/opt/homebrew/opt", "php@"+v.String(), "bin", "php")
	if _, err := os.Stat(bin); err != nil {
		if bin, err = exec.LookPath("php"); err != nil || v != phpversion.Max {
			return false, false
		}
	}
	p := filepath.Join(t.TempDir(), "x.php")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	err := exec.Command(bin, "-n", "-l", p).Run()
	return err == nil, true
}

// parseCases exercise rarely used grammar and error recovery. valid is the
// expected outcome, cross-checked with php -l of that version when present;
// compile marks code PHP rejects only at compile time (php -l reports it,
// the parser rightly accepts it).
var parseCases = []struct {
	src     string
	ver     phpversion.Version
	valid   bool
	compile bool
}{
	{`<?php switch ($a) { case 1; f(); default; }`, phpversion.PHP84, true, false},
	{`<?php switch ($a) { foo; case 1: }`, phpversion.PHP84, false, false},
	{`<?php switch ($a) { case 1 f(); }`, phpversion.PHP84, false, false},
	{"<?php declare(ticks=1): f(); enddeclare;", phpversion.PHP84, true, false},
	{`<?php function f() { #[A] class X {} #[A] function g() {} }`, phpversion.PHP84, true, false},
	{`<?php function f() { #[A] const C = 1; }`, phpversion.PHP84, true, true},
	{`<?php function f() { #[A] final class X {} }`, phpversion.PHP84, true, false},
	{`<?php $f = #[A] static fn() => 1; $g = #[A] function () {};`, phpversion.PHP84, true, false},
	{`<?php #[A] static 1;`, phpversion.PHP84, false, false},
	{`<?php #[A] 1;`, phpversion.PHP84, false, false},
	{`<?php $o = new #[A] class {};`, phpversion.PHP84, true, false},
	{`<?php class A { 1 }`, phpversion.PHP84, false, false},
	{`<?php class A { public int $p { #[A] final get => 1; } public array $q { &get { return []; } } }`, phpversion.PHP84, true, false},
	{`<?php class A { public int $p { get; set; } }`, phpversion.PHP84, true, true},
	{`<?php interface I { public int $p { get; set; } }`, phpversion.PHP84, true, false},
	{`<?php class A { public int $p { get 1; } }`, phpversion.PHP84, false, false},
	{`<?php class A { public int $p { 1 } }`, phpversion.PHP84, false, false},
	{`<?php class A { use T { foo bar; } }`, phpversion.PHP84, false, false},
	{`<?php class A { use T { ; } }`, phpversion.PHP84, false, false},
	{`<?php class A { use T { 1 } }`, phpversion.PHP84, false, false},
	{`<?php class A { use T { foo as; } }`, phpversion.PHP84, false, false},
	{`<?php class A { use T { foo as protected; foo as list; foo as protected bar; } }`, phpversion.PHP84, true, false},
	{`<?php $a = clone();`, phpversion.PHP85, true, false},
	{`<?php $a = clone($b, ['x' => 1]);`, phpversion.PHP85, true, false},
	{`<?php $a = clone(...$b);`, phpversion.PHP85, true, false},
	{`<?php f()++;`, phpversion.PHP84, false, false},
	{`<?php $a->$$b; $a->{$b}; $a->b;`, phpversion.PHP84, true, false},
	{`<?php new $a->b; new $a['x']; new $a::$b; new $$a; new (f()); new A::$b; new static::$b(); new self::$x['k']->y;`, phpversion.PHP84, true, false},
	{`<?php $o instanceof A::$b; $o instanceof $a::$b::$c;`, phpversion.PHP84, true, false},
	{`<?php new $a::C;`, phpversion.PHP84, false, false},
	{`<?php new A::C;`, phpversion.PHP84, false, false},
	{`<?php new A->b;`, phpversion.PHP84, false, false},
	{`<?php new A[0];`, phpversion.PHP84, false, false},
	{`<?php new A()->b; new A()::C; new A()[0]; new class {}->x;`, phpversion.PHP84, true, false},
	{`<?php $s{0}; $s->a{1};`, phpversion.PHP74, true, false},
	{`<?php $s{0};`, phpversion.PHP84, false, false},
	{`<?php new $s{0};`, phpversion.PHP74, true, false},
	{`<?php new $s{0};`, phpversion.PHP84, false, false},
	{`<?php $a instanceof (B::class); $a instanceof $b; $a instanceof $b->c;`, phpversion.PHP84, true, false},
	{`<?php $a instanceof 1;`, phpversion.PHP84, false, false},
	{`<?php $a instanceof [1];`, phpversion.PHP84, false, false},
	{`<?php new 1;`, phpversion.PHP84, false, false},
	{`<?php $m = match(1) { default => 2 }; match(1) { default => 2 };`, phpversion.PHP84, true, false},
	{`<?php "${a[1]} ${a} ${a . 'b'} ${$a}";`, phpversion.PHP84, true, false},
	{`<?php readonly(1); function readonly() {}`, phpversion.PHP84, true, false},
	{`<?php readonly(1);`, phpversion.PHP81, true, false},
	{`<?php class A { public readonly (A&B)|null $x; }`, phpversion.PHP84, true, false},
	{`<?php class A { public readonly (A&B)|null $x; }`, phpversion.PHP81, false, false},
	{`<?php readonly;`, phpversion.PHP84, false, false},
	{`<?php readonly class A {} readonly final class B {} $o = new readonly class {};`, phpversion.PHP84, true, false},
	{`<?php class A { private( set ) int $a; }`, phpversion.PHP84, false, false},
	{`<?php class A { PRIVATE(SET) int $a; }`, phpversion.PHP84, true, false},
	{`<?php $a = 1..2;`, phpversion.PHP84, false, false},
	{`<?php $a = 1. . 2;`, phpversion.PHP84, true, false},
	{`<?php while (1) { break 3; }`, phpversion.PHP84, true, true},
	{`<?php $a = Foreach::x(); $b = Callable::C;`, phpversion.PHP84, false, false},
	{`<?php $a = Enum::x(); $b = Mixed::C; $c = self::C; $d = static::C;`, phpversion.PHP84, true, false},
}

func TestParseCases(t *testing.T) {
	for _, c := range parseCases {
		f := Parse("t.php", []byte(c.src), Options{Version: c.ver})
		checkSpans(t, f)
		if got := len(f.Errors) == 0; got != c.valid {
			t.Errorf("%s (%s): valid = %v, want %v (errors %v)", c.src, c.ver, got, c.valid, f.Errors)
		}
		if php, ok := phpLint(t, c.ver, c.src); ok && php != (c.valid && !c.compile) {
			t.Errorf("%s (%s): php -l valid = %v, table says %v (compile-time error: %v)", c.src, c.ver, php, c.valid, c.compile)
		}
	}
}

// TestNodeKinds parses a file using every construct and checks that each
// node kind occurs and names its own Go type.
func TestNodeKinds(t *testing.T) {
	src, err := os.ReadFile("testdata/cov/allkinds.php")
	if err != nil {
		t.Fatal(err)
	}
	if php, ok := phpLint(t, phpversion.PHP85, string(src)); ok && !php {
		t.Fatal("allkinds.php is not valid PHP")
	}
	files := []*File{Parse("allkinds.php", src, Options{Version: phpversion.PHP85})}
	if errs := files[0].Errors; len(errs) > 0 {
		t.Fatalf("allkinds.php: %v", errs)
	}
	// Recovery-only nodes.
	files = append(files, Parse("bad.php", []byte("<?php #[A] ; $a = ; ) abstract 1;"), Options{Version: phpversion.PHP85}))
	seen := map[NodeKind]bool{}
	for _, f := range files {
		InspectFile(f, func(n Node) bool {
			k := n.Kind()
			seen[k] = true
			if name := reflect.TypeOf(n).Elem().Name(); name != k.String() {
				t.Errorf("%s.Kind() = %s", name, k)
			}
			return true
		})
	}
	for k := KindInvalid + 1; k < NumNodeKinds; k++ {
		if !seen[k] {
			t.Errorf("node kind %s never produced", k)
		}
	}
	if NumNodeKinds.String() != "NodeKind(?)" || TokenKind(255).String() != "TOKEN(?)" || TFunction.String() != "function" {
		t.Fatal("kind names")
	}
}

func TestMemoAndErrorText(t *testing.T) {
	f := &File{}
	// fn stores the key itself (as a concurrent first use would): the
	// value stored first wins.
	v := f.Memo("k", func() any {
		f.Memo("k", func() any { return 1 })
		return 2
	})
	if v != 1 || f.Memo("k", func() any { return 3 }) != 1 {
		t.Fatalf("memo = %v", v)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); f.Memo("c", func() any { return 1 }) }()
	}
	wg.Wait()
	if (Error{Span: Span{3, 4}, Msg: "m"}).Error() != "3: m" {
		t.Fatal("Error()")
	}
	p := &parser{src: []byte("abc")}
	if p.text(Token{Start: 3, End: 3}) != "" || p.text(Token{Start: 0, End: 2}) != "ab" {
		t.Fatal("text")
	}
}

func TestDeepExpressionAndTreeDepth(t *testing.T) {
	// Parenthesised nesting trips the parser's own recursion limit.
	f := Parse("p.php", []byte("<?php "+strings.Repeat("(", MaxDepth+10)+"1"+strings.Repeat(")", MaxDepth+10)+";"), Options{Version: phpversion.PHP84})
	if len(f.Errors) == 0 || !strings.Contains(f.Errors[len(f.Errors)-1].Msg, "nesting deeper") {
		t.Fatalf("parens: %v", f.Errors[len(f.Errors)-1])
	}
	// With an odd number of recursion levels before the parentheses, the
	// limit trips in parseExpr instead of parseUnary.
	for _, pre := range []string{"", "++"} {
		for n := MaxDepth/2 + 2; n < MaxDepth/2+6; n++ {
			f := Parse("p.php", []byte("<?php "+pre+strings.Repeat("(", n)+"$a"+strings.Repeat(")", n)+";"), Options{Version: phpversion.PHP84})
			if len(f.Errors) == 0 || len(f.Stmts) != 0 {
				t.Fatalf("%q depth %d: not reported", pre, n)
			}
		}
	}
	// A first statement deeper than MaxDepth stops the walk before the next.
	var s Stmt = &Block{}
	for i := 0; i < MaxDepth+5; i++ {
		s = &Block{Stmts: []Stmt{s}}
	}
	if d := TreeDepth([]Stmt{s, &Block{}}); d != MaxDepth+1 {
		t.Fatalf("depth %d", d)
	}
}
