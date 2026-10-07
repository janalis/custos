package syntax

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"custos/internal/phpver"
)

func parse(t *testing.T, src string, v phpver.Version) *File {
	t.Helper()
	f := Parse("test.php", []byte(src), Options{Version: v})
	checkSpans(t, f)
	return f
}

// checkSpans verifies span nesting/ordering and parent links.
func checkSpans(t *testing.T, f *File) {
	t.Helper()
	var visit func(parent Node)
	visit = func(parent Node) {
		ps := parent.Span()
		var prevEnd uint32
		first := true
		Children(parent, func(c Node) {
			cs := c.Span()
			if !ps.Contains(cs) {
				t.Fatalf("%s %v not within parent %s %v: %q", c.Kind(), cs, parent.Kind(), ps, f.Src[cs.Start:min(int(cs.End), len(f.Src))])
			}
			if !first && cs.Start < prevEnd {
				t.Fatalf("%s %v overlaps previous sibling ending at %d (parent %s)", c.Kind(), cs, prevEnd, parent.Kind())
			}
			if c.Parent() != parent {
				t.Fatalf("%s parent link broken", c.Kind())
			}
			first = false
			prevEnd = cs.End
			visit(c)
		})
	}
	for _, s := range f.Stmts {
		visit(s)
	}
}

func firstExpr(t *testing.T, f *File) Expr {
	t.Helper()
	for _, s := range f.Stmts {
		if es, ok := s.(*ExprStmt); ok {
			return es.Expr
		}
	}
	t.Fatal("no expression statement")
	return nil
}

func TestPrecedence(t *testing.T) {
	cases := []struct {
		src  string
		ver  phpver.Version
		want string // S-expression of operators
	}{
		{"$a + $b * $c;", phpver.PHP84, "(+ a (* b c))"},
		{"$a = $b + 1;", phpver.PHP84, "(= a (+ b 1))"},
		{"!$a instanceof B;", phpver.PHP84, "(! (instanceof a B))"},
		{"-$a ** 2;", phpver.PHP84, "(- (** a 2))"},
		{"$a ?? $b ?? $c;", phpver.PHP84, "(?? a (?? b c))"},
		{"$a . $b + $c;", phpver.PHP84, "(. a (+ b c))"},
		{"$a . $b + $c;", phpver.PHP74, "(+ (. a b) c)"},
		{"$a and $b = $c;", phpver.PHP84, "(and a (= b c))"},
		{"$a = $b and $c;", phpver.PHP84, "(and (= a b) c)"},
		{"!$a = f();", phpver.PHP84, "(! (= a f()))"},
		{"$a ? $b : $c ? $d : $e;", phpver.PHP74, "(? (? a b c) d e)"},
		{"$a ?: $b;", phpver.PHP84, "(? a _ b)"},
		{"$a |> f(...) |> g(...);", phpver.PHP85, "(|> (|> a f()) g())"},
		{"$a == $b && $c;", phpver.PHP84, "(&& (== a b) c)"},
		{"$a & $b == $c;", phpver.PHP84, "(& a (== b c))"},
		{"(int)$a + 1;", phpver.PHP84, "(+ ((int) a) 1)"},
		{"print $a and $b;", phpver.PHP84, "(and (print a) b)"},
		{"$x = &$y;", phpver.PHP84, "(=& x y)"},
		{"$a .= $b . $c;", phpver.PHP84, "(.= a (. b c))"},
	}
	for _, c := range cases {
		f := parse(t, "<?php "+c.src, c.ver)
		if len(f.Errors) > 0 {
			t.Errorf("%s: errors %v", c.src, f.Errors)
			continue
		}
		if got := sexpr(firstExpr(t, f), f.Src); got != c.want {
			t.Errorf("%s (%s): got %s want %s", c.src, c.ver, got, c.want)
		}
	}
}

func sexpr(e Expr, src []byte) string {
	tok := func(r TokenRef) string { return string(src[r.Span.Start:r.Span.End]) }
	switch n := e.(type) {
	case *Binary:
		return fmt.Sprintf("(%s %s %s)", tok(n.Op), sexpr(n.Left, src), sexpr(n.Right, src))
	case *Assign:
		op := tok(n.Op)
		if n.ByRef {
			op += "&"
		}
		return fmt.Sprintf("(%s %s %s)", op, sexpr(n.Var, src), sexpr(n.Value, src))
	case *Unary:
		return fmt.Sprintf("(%s %s)", tok(n.Op), sexpr(n.Expr, src))
	case *Instanceof:
		return fmt.Sprintf("(instanceof %s %s)", sexpr(n.Expr, src), sexpr(n.Class, src))
	case *Ternary:
		then := "_"
		if n.Then != nil {
			then = sexpr(n.Then, src)
		}
		return fmt.Sprintf("(? %s %s %s)", sexpr(n.Cond, src), then, sexpr(n.Else, src))
	case *Print:
		return fmt.Sprintf("(print %s)", sexpr(n.Expr, src))
	case *Variable:
		return n.Name
	case *FuncCall:
		return sexpr(n.Name, src) + "()"
	case *ConstFetch:
		return n.Name.Value
	case *Name:
		return n.Value
	case *Literal:
		return n.Raw
	case *Paren:
		return sexpr(n.Expr, src)
	}
	s := e.Span()
	return string(src[s.Start:s.End])
}

func TestParseConstructs(t *testing.T) {
	srcs := []string{
		`namespace A\B; use C\D as E, F; use function G\h; use I\{J, function k, const L as M};`,
		`#[Attr(1), Other] final readonly class Foo extends Bar implements Baz, Qux {
			use T1, T2 { T1::m insteadof T2; T2::m as protected n; m as public; }
			public const int X = 1, Y = 2;
			private static ?array $a = [], $b;
			public function __construct(private readonly int $x = 0, protected string|int &...$rest) {}
			abstract protected function f(): static;
			public string $name { get => strtoupper($this->name); set(string $v) { $this->name = $v; } }
			public private(set) int $count = 0;
			public function m(A&B $x, (C&D)|null $y): never { throw new E(); }
		}`,
		`enum Suit: string implements HasLabel { case Hearts = 'H'; case Spades = 'S'; const Wild = self::Spades; public function label(): string { return match($this) { self::Hearts => 'h', default => 's' }; } }`,
		`interface I extends A, B { public function f(); } trait T { abstract public function g(); }`,
		`function &gen(int ...$n): \Generator { yield 1; yield $k => $v; $x = yield; yield from other(); }`,
		`$f = static fn &(array $a = []): int => count($a); $g = function () use ($x, &$y): void {}; $h = #[A] fn() => 1;`,
		`if ($a): echo 1; elseif ($b): echo 2; else: echo 3; endif; while ($x): $x--; endwhile; for ($i = 0, $j = 1; $i < 10; $i++): endfor; foreach ($a as $k => &$v): endforeach; switch ($x): case 1: break; default: endswitch;`,
		`if ($a) { } elseif ($b) { } else if ($c) { } else { }`,
		`do { $i++; } while ($i < 10); declare(strict_types=1); declare(ticks=1) { tick(); }`,
		`try { f(); } catch (A | B $e) { } catch (C) { } finally { }`,
		`$a = [1, 'k' => 2, ...$rest, &$ref]; [$x, [, $y]] = $p; list('a' => $q, 'b' => list($r)) = $s; [$a[], $b->c] = f();`,
		`echo "a $b {$c->d['e']} ${f} ${g['h']} $i[0] $j[k] $l[$m] $n->o $p?->q \$r", <<<EOT
  x $y {$z}
  EOT, <<<'N'
 raw $x
 N, ` + "`ls $dir`" + `;`,
		`$x = new class(1) extends Base implements I { public $p; }; $y = new Foo; $z = new $cls['k']->name(); $w = new (trim($n))(); $v = new static; new Foo()->bar();`,
		`A::$b::c(); A::{$m}(); $obj::CONST; static::f(); parent::__construct(); self::$x[0] = 1; $a->{'b'}(); $a?->b()?->c; $f(...); strlen(...);`,
		`f(a: 1, b: $c, ...$rest); exit; die(1); exit(); isset($a, $b[0]); empty($x); eval('1;'); unset($a, $b,);`,
		`global $a, $$b; static $s = 1, $t; goto end; end: echo 1; __halt_compiler(); raw data here`,
		`$a = clone $b->c; $d = include 'x.php'; require_once __DIR__ . '/y.php'; print 'z'; $e = @file('f'); $g = (object)['a' => 1];`,
		`const A = 1, B = A * 2; function f($a = null, $b = [1, 2], $c = self::X) {} $x = $y ?: $z; $q ??= 5;`,
		`$s = "${a}" . "{$b}" . "$c[-1]"; $m = $a{0};`,
		`$copy = clone($this, ['x' => 1]); $c2 = clone $a;`,
	}
	for _, src := range srcs {
		v := phpver.PHP85
		if strings.Contains(src, "$a{0}") {
			v = phpver.PHP73
		}
		f := parse(t, "<?php\n"+src, v)
		if len(f.Errors) > 0 {
			t.Errorf("%.60s…: %v", src, f.Errors)
		}
	}
}

func TestParseVersionSpecific(t *testing.T) {
	// fn/match/enum are identifiers in old versions.
	f := parse(t, "<?php function fn() {} match(1); class enum {}", phpver.PHP73)
	if len(f.Errors) > 0 {
		t.Fatalf("7.3: %v", f.Errors)
	}
	// #[ is a comment before 8.0.
	f = parse(t, "<?php #[Attr]\nfunction f() {}", phpver.PHP74)
	if len(f.Errors) > 0 || len(f.Stmts) != 1 {
		t.Fatalf("7.4 attributes: %v", f.Errors)
	}
}

func TestErrorRecovery(t *testing.T) {
	srcs := []string{
		"<?php $a = ;\n$b = 1;",
		"<?php function f( { }\nclass A { public function g() { $x = } }",
		"<?php if ($a { echo 1; }",
		"<?php $a->",
		"<?php class",
		"<?php \"unterminated $x",
		"<?php foo(1, 2",
		"<?php [1, 2",
		"<?php }}} $x = 1;",
	}
	for _, src := range srcs {
		f := parse(t, src, phpver.PHP84)
		if len(f.Errors) == 0 {
			t.Errorf("%q: expected errors", src)
		}
	}
	f := parse(t, "<?php $a = ;\n$b = 1;", phpver.PHP84)
	if len(f.Stmts) < 2 {
		t.Errorf("recovery lost statements:\n%v", f.Stmts)
	}
}

// TestParseCorpus parses every corpus file (PHP 8.5) and requires zero
// errors for files that `php -l` accepts.
func TestParseCorpus(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	files := corpusFiles(t)
	bad := 0
	var badFiles []string
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		f := Parse(path, src, Options{Version: phpver.PHP85})
		checkSpans(t, f)
		if len(f.Errors) > 0 {
			badFiles = append(badFiles, path)
			if len(badFiles) <= 3000 {
				continue
			}
		}
	}
	// Files PHP itself rejects are expected to fail.
	for _, path := range badFiles {
		if out, err := exec.Command("php", "-l", path).CombinedOutput(); err != nil && strings.Contains(string(out), "error") {
			continue
		}
		src, _ := os.ReadFile(path)
		f := Parse(path, src, Options{Version: phpver.PHP85})
		bad++
		if bad <= 20 {
			e := f.Errors[0]
			ctx := src[max(0, int(e.Span.Start)-40):min(len(src), int(e.Span.Start)+40)]
			t.Errorf("%s: %s near %q", path, e.Msg, ctx)
		}
	}
	t.Logf("%d files, %d unexpected failures", len(files), bad)
}

func BenchmarkParse(b *testing.B) {
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		b.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		b.Skip(err)
	}
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse(path, src, Options{Version: phpver.PHP85})
	}
}

func BenchmarkLex(b *testing.B) {
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		b.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		b.Skip(err)
	}
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Lex(src, LexOptions{Version: phpver.PHP85})
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"<?php $a = 1 + 2 * 3;", "<?php class A { function b() { return $this?->c; } }",
		"<?php if ($a): elseif: endif;", "<?php fn($x) => match($x) { 1, 2 => 3, default => 4 };",
		"<?php \"$a[0] {$b} ${c}\";", "<?php #[A] enum E: int { case X = 1; }",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		for _, v := range []phpver.Version{phpver.PHP56, phpver.PHP85} {
			file := Parse("fuzz.php", []byte(s), Options{Version: v})
			checkSpans(t, file)
		}
	})
}

func TestParseBest(t *testing.T) {
	src := []byte("<?php class A { public private(set) int $x = 0; }")
	if f := Parse("a.php", src, Options{Version: phpver.PHP81}); len(f.Errors) == 0 {
		t.Fatal("8.1 parse should fail on asymmetric visibility")
	}
	if f := ParseBest("a.php", src, Options{Version: phpver.PHP81}); len(f.Errors) != 0 || f.Version != phpver.PHP81 {
		t.Fatalf("ParseBest: %v %s", f.Errors, f.Version)
	}
	old := []byte("<?php function match($a) { return $a; } echo match(1);")
	if f := ParseBest("b.php", old, Options{Version: phpver.PHP74}); len(f.Errors) != 0 {
		t.Fatalf("legacy identifiers: %v", f.Errors)
	}
}
