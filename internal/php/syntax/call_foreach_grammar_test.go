package syntax

import (
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestCallablePlaceholderGrammar(t *testing.T) {
	for _, call := range []string{"f", "$object->f", "C::f", "$callable"} {
		for _, args := range []struct {
			text  string
			count int
			valid bool
		}{
			{"...", 1, true},
			{"1, ...", 2, false},
			{"...$items, ...", 2, false},
			{"name: 1, ...", 2, false},
		} {
			t.Run(call+"("+args.text+")", func(t *testing.T) {
				src := "<?php " + call + "(" + args.text + "); $after = 1;"
				for _, version := range []phpversion.Version{phpversion.PHP81, phpversion.PHP85} {
					f := parse(t, src, version)
					if (len(f.Errors) == 0) != args.valid {
						t.Fatalf("%s: errors = %v, valid = %v", version, f.Errors, args.valid)
					}
					if !args.valid && (len(f.Errors) != 1 || string(f.Src[f.Errors[0].Span.Start:f.Errors[0].Span.End]) != "...") {
						t.Fatalf("incorrect placeholder error range: %v", f.Errors)
					}
					var list *ArgList
					Children(f.Stmts[0], func(expr Node) {
						Children(expr, func(child Node) {
							if a, ok := child.(*ArgList); ok {
								list = a
							}
						})
					})
					if list == nil || len(list.Args) != args.count {
						t.Fatalf("lost argument list: %+v", list)
					}
					placeholder := list.Args[len(list.Args)-1].(*VariadicPlaceholder)
					if got := string(f.Src[placeholder.Span().Start:placeholder.Span().End]); got != "..." {
						t.Fatalf("placeholder span = %q", got)
					}
					assertFollowingAssignment(t, f)
					if valid, ok := phpLint(t, version, src); ok && valid != args.valid {
						t.Fatalf("%s: PHP valid = %v, want %v", version, valid, args.valid)
					}
				}
			})
		}
	}
}

func TestForeachTargetGrammar(t *testing.T) {
	for _, tc := range []struct {
		target  string
		valid   bool
		compile bool
	}{
		{"$value", true, false},
		{"$$name", true, false},
		{"$values[0]", true, false},
		{"$object->field", true, false},
		{"C::$field", true, false},
		{"($object)->field", true, false},
		{"($values)[0]", true, false},
		{"&$value", true, false},
		{"$key => &$value", true, false},
		{"list($value)", true, false},
		{"[$value]", true, false},
		{"[$key] => $value", true, true},
		{"list($key) => $value", true, true},
		{"&$key => $value", true, true},
		{"f()", true, true},
		{"$object->f()", true, true},
		{"C::f()", true, true},
		{"&f()", true, true},
		{"$object?->field", true, true},
		{"123", false, false},
		{"&123", false, false},
		{"'value'", false, false},
		{"($value)", false, false},
		{"-$value", false, false},
		{"$value++", false, false},
		{"++$value", false, false},
		{"$value = 1", false, false},
		{"array($value)", false, false},
		{"&list($value)", false, false},
		{"&[$value]", false, false},
		{"$key => 123", false, false},
	} {
		t.Run(tc.target, func(t *testing.T) {
			src := "<?php foreach ($items as " + tc.target + ") {} $after = 1;"
			f := parse(t, src, phpversion.PHP85)
			if (len(f.Errors) == 0) != tc.valid {
				t.Fatalf("errors = %v, valid = %v", f.Errors, tc.valid)
			}
			loop := f.Stmts[0].(*Foreach)
			targets := strings.Split(tc.target, "=>")
			if loop.ByRef != strings.HasPrefix(strings.TrimSpace(targets[len(targets)-1]), "&") {
				t.Fatal("lost value reference marker")
			}
			assertFollowingAssignment(t, f)
			if valid, ok := phpLint(t, phpversion.PHP85, src); ok && valid != (tc.valid && !tc.compile) {
				t.Fatalf("PHP valid = %v, grammar valid = %v, compile rejection = %v", valid, tc.valid, tc.compile)
			}
		})
	}
}

func TestAdaptiveCallableAndForeachGrammar(t *testing.T) {
	for _, statement := range []string{"f(1, ...);", "foreach ($items as 123) {}"} {
		f := ParseBest("invalid.php", []byte("<?php "+statement+" $after = 1;"), Options{Version: phpversion.PHP53})
		checkSpans(t, f)
		if len(f.Errors) == 0 || f.Version != phpversion.PHP53 {
			t.Fatalf("adaptive parser lost grammar errors or version: %v, %s", f.Errors, f.Version)
		}
		assertFollowingAssignment(t, f)
	}
}

func assertFollowingAssignment(t *testing.T, f *File) {
	t.Helper()
	last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
	if last.Var.(*Variable).Name != "after" {
		t.Fatal("lost following assignment")
	}
}

func BenchmarkCallableAndForeachGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("$f = C::f(...); foreach ($items as $key => [$value]) {} foreach ($items as &$value) {} ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("call-foreach.php", src, Options{Version: phpversion.PHP85})
	}
}
