package syntax

import (
	"slices"
	"testing"

	phpversion "github.com/janalis/custos/internal/php/version"
)

// TestPartialApplicationGrammar checks `?` and trailing `...` arguments
// (PHP 8.6) against php -l at 8.5 and 8.6 when those binaries exist.
func TestPartialApplicationGrammar(t *testing.T) {
	for _, tc := range []struct {
		src   string
		valid bool // at 8.6; nothing here is valid at 8.5
		err   string
	}{
		{`f(?, 2);`, true, ""},
		{`f(1, ...);`, true, ""},
		{`f(?, ...);`, true, ""},
		{`f(?, ?, ...);`, true, ""},
		{`f(1, ?, 3);`, true, ""},
		{`f(b: ?, a: 1);`, true, ""},
		{`f(x: ?, ...);`, true, ""},
		{`f(?, a: ?, ...);`, true, ""},
		{`f(?)(?);`, true, ""},
		{`f(?,);`, true, ""},
		{`f(...,);`, true, ""},
		{`$o->m(?, ?);`, true, ""},
		{`C::m(?);`, true, ""},
		{`$f(?, 1);`, true, ""},
		{`clone(?);`, true, ""},
		{`clone($o, ?);`, true, ""},
		{`exit(?);`, true, ""},
		{`f(..., 1);`, false, "a variadic placeholder must be the last argument"},
		{`f(..., ?);`, false, "a variadic placeholder must be the last argument"},
		{`f(...$a, ?);`, false, "a partial application cannot unpack arguments"},
		{`f(?, ...$a);`, false, "a partial application cannot unpack arguments"},
		{`f(? ?: 1);`, false, ""},
		{`f(?->x);`, false, ""},
		{`isset(?);`, false, ""},
		{`array(?);`, false, ""},
	} {
		t.Run(tc.src, func(t *testing.T) {
			src := "<?php " + tc.src + " $after = 1;"
			f := parse(t, src, phpversion.PHP86)
			if (len(f.Errors) == 0) != tc.valid {
				t.Fatalf("8.6: errors = %v, valid = %v", f.Errors, tc.valid)
			}
			if tc.err != "" && (len(f.Errors) != 1 || f.Errors[0].Msg != tc.err) {
				t.Fatalf("8.6: errors = %v, want %q", f.Errors, tc.err)
			}
			assertFollowingAssignment(t, f)
			if old := parse(t, src, phpversion.PHP85); len(old.Errors) == 0 {
				t.Fatal("8.5 accepts partial application")
			}
			if php, ok := phpLint(t, phpversion.PHP86, src); ok && php != tc.valid {
				t.Fatalf("php 8.6 valid = %v, want %v", php, tc.valid)
			}
			if php, ok := phpLint(t, phpversion.PHP85, src); ok && php {
				t.Fatal("php 8.5 accepts it")
			}
		})
	}
}

func TestArgPlaceholderNodes(t *testing.T) {
	f := parse(t, "<?php f(?, 3, n: ?, ...);", phpversion.PHP86)
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	list := firstExpr(t, f).(*FuncCall).Args
	if !list.Partial() || list.FirstClassCallable() || len(list.Args) != 4 {
		t.Fatalf("list: %+v", list)
	}
	first, named := list.Args[0].(*ArgPlaceholder), list.Args[2].(*ArgPlaceholder)
	if first.Name != nil || named.Name == nil || named.Name.Value != "n" {
		t.Fatal("placeholder names")
	}
	if got := string(f.Src[named.Span().Start:named.Span().End]); got != "n: ?" {
		t.Fatalf("named placeholder span = %q", got)
	}
	if _, ok := list.Args[3].(*VariadicPlaceholder); !ok {
		t.Fatal("trailing variadic placeholder")
	}
	var seen []string
	Inspect(named, func(n Node) bool {
		seen = append(seen, n.Kind().String())
		return true
	})
	if len(seen) != 2 || seen[1] != "Identifier" {
		t.Fatalf("walk = %v", seen)
	}
	if (*ArgList)(nil).Partial() || parse(t, "<?php f(1);", phpversion.PHP86).Stmts[0].(*ExprStmt).Expr.(*FuncCall).Args.Partial() {
		t.Fatal("plain calls are not partial")
	}
}

// TestCallCompileErrors: argument lists PHP's grammar accepts but its
// compiler rejects are reported as compile errors.
func TestCallCompileErrors(t *testing.T) {
	for _, tc := range []struct{ src, err string }{
		{`new Foo(...);`, "new cannot create a closure"},
		{`new Foo(1, ?);`, "new cannot create a closure"},
		{`new class(...) {};`, "new cannot create a closure"},
		{`$o?->m(...);`, "a nullsafe call cannot create a closure"},
		{`$o?->m(?);`, "a nullsafe call cannot create a closure"},
		{`f(a: 1, 2);`, "a positional argument cannot follow a named argument"},
		{`f(a: 1, ?);`, "a positional argument cannot follow a named argument"},
		{`f(a: 1, ...$x);`, "arguments cannot be unpacked after a named argument"},
		{`f(..., 1);`, "a variadic placeholder must be the last argument"},
		{`f(?, ...$a);`, "a partial application cannot unpack arguments"},
		{`#[A(...)] function g() {}`, "an attribute argument cannot create a closure"},
		{`#[A(1, ?)] function g() {}`, "an attribute argument cannot create a closure"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			src := "<?php " + tc.src + " $after = 1;"
			f := parse(t, src, phpversion.PHP86)
			if len(f.Errors) != 1 || f.Errors[0].Msg != tc.err || !f.Errors[0].Compile {
				t.Fatalf("errors = %+v, want compile error %q", f.Errors, tc.err)
			}
			assertFollowingAssignment(t, f)
			if php, ok := phpLint(t, phpversion.PHP86, src); ok && php {
				t.Fatal("php 8.6 accepts it")
			}
		})
	}
	for _, src := range []string{`f(a: 1, ...);`, `f(...$x, a: 1);`, `$o?->m(1)(...);`, `new Foo()(...);`, `$o->m(?);`, `f(a: ?, b: 2, ...);`, `#[A(f(?))] function g() {}`} {
		if f := parse(t, "<?php "+src, phpversion.PHP86); len(f.Errors) != 0 {
			t.Errorf("%s: %v", src, f.Errors)
		}
	}
	if f := parse(t, "<?php f(1, ...);", phpversion.PHP85); len(f.Errors) != 1 || f.Errors[0].Compile {
		t.Fatalf("a placeholder after arguments is a parse error before 8.6: %+v", f.Errors)
	}
}

// TestEnumMagicMethods: an enum may not declare most magic methods; PHP 8.6
// allows __debugInfo.
func TestEnumMagicMethods(t *testing.T) {
	for _, tc := range []struct {
		method string
		v      phpversion.Version
		bad    bool
	}{
		{"public function __toString(): string { return ''; }", phpversion.PHP86, true},
		{"public function __CONSTRUCT() {}", phpversion.PHP86, true},
		{"public function __sleep() { return []; }", phpversion.PHP86, true},
		{"public static function __set_state(array $a) { return self::A; }", phpversion.PHP86, true},
		{"public function __debugInfo(): array { return []; }", phpversion.PHP85, true},
		{"public function __debugInfo(): array { return []; }", phpversion.PHP86, false},
		{"public function __invoke() {}", phpversion.PHP86, false},
		{"public function __call($n, $a) {}", phpversion.PHP86, false},
		{"public function label() {}", phpversion.PHP86, false},
	} {
		src := "<?php enum E { case A; " + tc.method + " }"
		f := parse(t, src, tc.v)
		if got := len(f.Errors) == 1 && f.Errors[0].Compile; got != tc.bad || !tc.bad && len(f.Errors) != 0 {
			t.Errorf("%s (%s): errors %+v", tc.method, tc.v, f.Errors)
		}
		if _, recovered := f.RecoveryError(); recovered {
			t.Errorf("%s: a compile error needs no recovery", tc.method)
		}
		// php -l misses some of them, which PHP raises when it declares the
		// class: only an accepted declaration is checked against it.
		if php, ok := phpLint(t, tc.v, src); ok && !tc.bad && !php {
			t.Errorf("%s: php %s rejects it", tc.method, tc.v)
		}
	}
	// A class may declare them all.
	if f := parse(t, "<?php class C { public function __toString(): string { return ''; } }", phpversion.PHP86); len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	if e, ok := parse(t, "<?php enum E { case A; public function __clone() {} } $x = ;", phpversion.PHP86).RecoveryError(); !ok || e.Compile {
		t.Fatal("the parse error is the one recovered from")
	}
}

// TestCompileErrorsOnceInOrder: a misplaced `...` is reported once however
// many arguments follow it, and compile errors found after later arguments
// still come out in source order.
func TestCompileErrorsOnceInOrder(t *testing.T) {
	for src, want := range map[string][]string{
		"<?php f(..., a: 1, 2, 3);": {
			"a variadic placeholder must be the last argument",
			"a positional argument cannot follow a named argument",
			"a positional argument cannot follow a named argument",
		},
		"<?php f(?, ...$a, new Foo(...));": {
			"a partial application cannot unpack arguments",
			"new cannot create a closure",
		},
	} {
		f := parse(t, src, phpversion.PHP86)
		var got []string
		for i, e := range f.Errors {
			got = append(got, e.Msg)
			if i > 0 && e.Span.Start < f.Errors[i-1].Span.Start {
				t.Errorf("%s: errors out of order: %+v", src, f.Errors)
			}
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s: errors = %q, want %q", src, got, want)
		}
	}
}

// TestParseErrorPrevails: a parse error at the position of a compile error
// replaces it, so that the file counts as recovered.
func TestParseErrorPrevails(t *testing.T) {
	for src, v := range map[string]phpversion.Version{
		"<?php f(a: 1, ,);": phpversion.PHP86,
		"<?php f(a: 1, ?);": phpversion.PHP85,
	} {
		f := parse(t, src, v)
		if _, recovered := f.RecoveryError(); !recovered {
			t.Errorf("%s (%s): %+v", src, v, f.Errors)
		}
	}
}

// BenchmarkParseCallArguments covers argument lists with unpacking,
// placeholders and named arguments.
func BenchmarkParseCallArguments(b *testing.B) {
	src := []byte("<?php f(...$a); g(...); h(?, 2, x: ?, ...); k($a, $b, name: $c); m(...$a, x: 1);")
	b.ReportAllocs()
	for b.Loop() {
		Parse("args.php", src, Options{Version: phpversion.PHP86})
	}
}
