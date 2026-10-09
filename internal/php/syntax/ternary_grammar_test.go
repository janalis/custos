package syntax

import (
	"fmt"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestTernaryChainGrammar(t *testing.T) {
	for _, tc := range []struct {
		expr    string
		invalid bool
	}{
		{"$a ? $b : $c ? $d : $e", true},
		{"$a ? $b : $c ?: $d", true},
		{"$a ?: $b ? $c : $d", true},
		{"$a ?: $b ?: $c", false},
		{"$a ?: $b ?: $c ?: $d", false},
		{"$a ? $b ? $c : $d : $e", false},
		{"$a ? $b ?: $c : $d", false},
		{"($a ? $b : $c) ? $d : $e", false},
		{"$a ? $b : ($c ? $d : $e)", false},
		{"($a ? $b : $c) ?: $d", false},
		{"($a ?: $b) ? $c : $d", false},
		{"$a ? $b : ($c ?: $d)", false},
		{"$a ?: ($b ? $c : $d)", false},
	} {
		for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP74, phpversion.PHP80, phpversion.PHP85} {
			t.Run(fmt.Sprintf("%s/%s", tc.expr, version), func(t *testing.T) {
				src := "<?php " + tc.expr + "; $after = 1;"
				invalid := tc.invalid && version.AtLeast(phpversion.PHP80)
				for _, permissive := range []bool{false, true} {
					for _, adaptive := range []bool{false, true} {
						t.Run(fmt.Sprintf("permissive=%t/adaptive=%t", permissive, adaptive), func(t *testing.T) {
							options := Options{Version: version, Permissive: permissive}
							parseFn := Parse
							if adaptive {
								parseFn = ParseBest
							}
							f := parseFn("ternary.php", []byte(src), options)
							checkSpans(t, f)
							if f.Version != version {
								t.Fatalf("version = %s, want %s", f.Version, version)
							}
							wantErrors := 0
							if invalid {
								wantErrors = 1
							}
							if len(f.Errors) != wantErrors {
								t.Fatalf("errors = %v, want %d", f.Errors, wantErrors)
							}
							if invalid {
								position := uint32(strings.LastIndexByte(src, '?'))
								if f.Errors[0].Span != (Span{position, position + 1}) || f.Errors[0].Msg != "ternary operators require parentheses when chained" {
									t.Fatalf("incorrect error: %+v", f.Errors[0])
								}
							}
							if len(f.Stmts) != 2 {
								t.Fatalf("statement count = %d", len(f.Stmts))
							}
							n := firstExpr(t, f).(*Ternary)
							if n.Span() != (Span{6, uint32(6 + len(tc.expr))}) {
								t.Fatalf("ternary span = %+v", n.Span())
							}
							if tc.invalid {
								if _, ok := n.Cond.(*Ternary); !ok {
									t.Fatal("lost left-associated ternary AST")
								}
							}
							last := f.Stmts[1].(*ExprStmt).Expr.(*Assign)
							if last.Var.(*Variable).Name != "after" {
								t.Fatal("lost following assignment")
							}
						})
					}
				}
				if valid, ok := phpLint(t, version, src); ok && valid == invalid {
					t.Fatalf("PHP valid = %t, want %t", valid, !invalid)
				}
			})
		}
	}
}

func BenchmarkTernaryGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("$a ? $b : $c ? $d : $e; $a ?: $b ?: $c; $a ? $b ? $c : $d : $e; ", 128))
	for _, version := range []phpversion.Version{phpversion.PHP74, phpversion.PHP85} {
		b.Run(version.String(), func(b *testing.B) {
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			for b.Loop() {
				Parse("ternary.php", src, Options{Version: version})
			}
		})
	}
}
