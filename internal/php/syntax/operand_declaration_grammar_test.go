package syntax

import (
	"fmt"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestUnsetAndIssetGrammar(t *testing.T) {
	for _, tc := range []struct {
		statement string
		valid     bool
		compile   bool
		errorText string
	}{
		{"unset();", false, false, ")"},
		{"isset();", false, false, ")"},
		{"unset(1);", false, false, "1"},
		{"unset('text');", false, false, "'text'"},
		{"unset(($x));", false, false, "("},
		{"unset($x + 1);", false, false, "$x"},
		{"unset($x = 1);", false, false, "$x"},
		{"unset(++$x);", false, false, "++"},
		{"unset($x++);", false, false, "$x"},
		{"unset([$x]);", false, false, "["},
		{"unset(list($x));", false, false, "list"},
		{"unset(C::X);", false, false, "C"},
		{"unset($x, 1, $y);", false, false, "1"},
		{"unset($x);", true, false, ""},
		{"unset($x, $y,);", true, false, ""},
		{"unset($$name);", true, false, ""},
		{"unset($x[0]);", true, false, ""},
		{"unset($x->field);", true, false, ""},
		{"unset(C::$field);", true, false, ""},
		{"unset(($x)->field);", true, false, ""},
		{"unset(($x)[0]);", true, false, ""},
		{"unset(f());", true, true, ""},
		{"unset($x->f());", true, true, ""},
		{"unset(C::f());", true, true, ""},
		{"unset($x[]);", true, true, ""},
		{"unset($x?->field);", true, true, ""},
		{"isset($x);", true, false, ""},
		{"isset($x, $y,);", true, false, ""},
		{"isset(1);", true, true, ""},
		{"isset($x + 1);", true, true, ""},
	} {
		t.Run(tc.statement, func(t *testing.T) {
			for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP85} {
				src := "<?php " + tc.statement + " $after = 1;"
				checkOperandDeclarationGrammar(t, src, version, tc.valid, tc.compile, tc.errorText, func(f *File) {
					// Retain all operands even if a middle operand is invalid.
					if tc.statement == "unset($x, 1, $y);" && len(f.Stmts[0].(*Unset).Vars) != 3 {
						t.Fatal("lost unset operands")
					}
				})
			}
		})
	}
}

func TestDeclarationClauseGrammar(t *testing.T) {
	for _, tc := range []struct {
		declaration string
		version     phpversion.Version
		valid       bool
		errorText   string
		extends     int
		implements  int
	}{
		{"class C", phpversion.PHP53, true, "", 0, 0},
		{"class C extends A implements I, J", phpversion.PHP53, true, "", 1, 2},
		{"class C extends A, B, D", phpversion.PHP53, false, ",", 3, 0},
		{"interface C extends A, B, D", phpversion.PHP53, true, "", 3, 0},
		{"interface C implements I", phpversion.PHP53, false, "implements", 0, 1},
		{"interface C extends A, B implements I, J", phpversion.PHP53, false, "implements", 2, 2},
		{"trait C", phpversion.PHP54, true, "", 0, 0},
		{"trait C extends A, B", phpversion.PHP54, false, "extends", 2, 0},
		{"trait C implements I, J", phpversion.PHP54, false, "implements", 0, 2},
		{"enum C", phpversion.PHP81, true, "", 0, 0},
		{"enum C implements I, J", phpversion.PHP81, true, "", 0, 2},
		{"enum C extends A", phpversion.PHP81, false, "extends", 1, 0},
		{"$x = new class extends A implements I, J", phpversion.PHP70, true, "", 1, 2},
		{"$x = new class extends A, B, D", phpversion.PHP70, false, ",", 3, 0},
	} {
		t.Run(tc.declaration, func(t *testing.T) {
			for _, version := range []phpversion.Version{tc.version, phpversion.PHP85} {
				src := "<?php " + tc.declaration + " { const X = 1; }"
				if strings.HasPrefix(tc.declaration, "$x") {
					src += ";"
				}
				src += " $after = 1;"
				checkOperandDeclarationGrammar(t, src, version, tc.valid, false, tc.errorText, func(f *File) {
					var declaration *ClassLike
					if s, ok := f.Stmts[0].(*ClassLike); ok {
						declaration = s
					} else {
						declaration = f.Stmts[0].(*ExprStmt).Expr.(*Assign).Value.(*New).Class.(*ClassLike)
					}
					if len(declaration.Extends) != tc.extends || len(declaration.Implements) != tc.implements || len(declaration.Members) != 1 {
						t.Fatalf("lost clause names or body: %+v", declaration)
					}
				})
			}
		})
	}
}

func checkOperandDeclarationGrammar(t *testing.T, src string, version phpversion.Version, valid, compile bool, errorText string, inspect func(*File)) {
	t.Helper()
	for _, mode := range []string{"strict", "permissive", "adaptive"} {
		t.Run(fmt.Sprintf("%s/%s", version, mode), func(t *testing.T) {
			opt := Options{Version: version, Permissive: mode == "permissive"}
			var f *File
			if mode == "adaptive" {
				f = ParseBest("grammar.php", []byte(src), opt)
			} else {
				f = Parse("grammar.php", []byte(src), opt)
			}
			checkSpans(t, f)
			if (len(f.Errors) == 0) != valid || f.Version != version {
				t.Fatalf("errors = %v, version = %s, want valid = %v, version = %s", f.Errors, f.Version, valid, version)
			}
			if !valid {
				if len(f.Errors) != 1 || string(f.Src[f.Errors[0].Span.Start:f.Errors[0].Span.End]) != errorText {
					t.Fatalf("incorrect error range: %v, want %q", f.Errors, errorText)
				}
			}
			assertFollowingAssignment(t, f)
			inspect(f)
		})
	}
	if phpValid, ok := phpLint(t, version, src); ok && phpValid != (valid && !compile) {
		t.Fatalf("%s: PHP valid = %v, grammar valid = %v, compile rejection = %v", version, phpValid, valid, compile)
	}
}

func BenchmarkOperandAndDeclarationGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("unset($x[0], $x->field,); isset($x, $y,); class C extends A implements I, J {} interface I extends J, K {} trait T {} enum E implements I {} ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("operands-declarations.php", src, Options{Version: phpversion.PHP85})
	}
}
