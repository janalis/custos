package syntax

import (
	"fmt"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestComplexInterpolationVariableGrammar(t *testing.T) {
	for _, tc := range []struct {
		expr  string
		valid bool
	}{
		{"$a", true},
		{"$$a", true},
		{"${f()}", true},
		{"$a[1 + 2]", true},
		{"$a->{f()}", true},
		{"$a()", true},
		{"$a->f()", true},
		{"$a::$x", true},
		{"$a::f()", true},
		{"$a + 1", false},
		{"$a = 1", false},
		{"$a ? 1 : 2", false},
		{"$a && $b", false},
		{"$a instanceof X", false},
		{"$a++", false},
		{"$a::C", false},
	} {
		for _, delimiter := range []string{"quote", "heredoc", "backtick"} {
			t.Run(delimiter+"/"+tc.expr, func(t *testing.T) {
				body := "{" + tc.expr + "}"
				literal := "\"" + body + "\""
				switch delimiter {
				case "heredoc":
					literal = "<<<TEXT\n" + body + "\nTEXT"
				case "backtick":
					literal = "`" + body + "`"
				}
				src := "<?php $s = " + literal + ";\n$after = 1;"
				for _, version := range []phpversion.Version{phpversion.PHP56, phpversion.PHP85} {
					checkVariableGrammar(t, src, version, tc.valid)
				}
			})
		}
	}
	for _, src := range []string{
		`<?php $s = "{$o?->x}"; $after = 1;`,
		`<?php $s = "{$o->f($a + 1)[f()]}"; $after = 1;`,
		`<?php $s = "{$a["{$b}"]}"; $after = 1;`,
	} {
		checkVariableGrammar(t, src, phpversion.PHP85, true)
	}
}

func checkVariableGrammar(t *testing.T, src string, version phpversion.Version, valid bool) {
	t.Helper()
	f := parse(t, src, version)
	if got := len(f.Errors) == 0; got != valid {
		t.Fatalf("%s valid = %v, want %v: %s: %v", version, got, valid, src, f.Errors)
	}
	last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
	if last.Var.(*Variable).Name != "after" {
		t.Fatal("lost following statement")
	}
	if phpValid, ok := phpLint(t, version, src); ok && phpValid != valid {
		t.Fatalf("%s PHP valid = %v, want %v: %s", version, phpValid, valid, src)
	}
	if !valid {
		for _, opt := range []Options{{Version: version, Permissive: true}, {Version: version}} {
			var parsed *File
			if opt.Permissive {
				parsed = Parse("invalid.php", []byte(src), opt)
			} else {
				parsed = ParseBest("invalid.php", []byte(src), opt)
			}
			checkSpans(t, parsed)
			if len(parsed.Errors) == 0 || parsed.Version != version {
				t.Fatalf("lost grammar error or target version: %v", parsed.Errors)
			}
		}
	}
}

func TestWriteTargetGrammar(t *testing.T) {
	for _, target := range []string{"$a", "$a[0]", "$a->x", "C::$a", "[]", "[$a]", "list($a)"} {
		for _, op := range []string{"++", "--", "+= 1", "??= 1", "= [1]"} {
			t.Run(target+op, func(t *testing.T) {
				valid := strings.HasPrefix(target, "$a") || target == "C::$a" || op == "= [1]"
				// An empty destructuring target is rejected at compile time.
				if target == "[]" && op == "= [1]" {
					return
				}
				src := "<?php " + target + op + "; $after = 1;"
				checkVariableGrammar(t, src, phpversion.PHP85, valid)
			})
		}
	}
}

func TestReferenceAssignmentOperandGrammar(t *testing.T) {
	for _, operand := range []string{"$b", "$$b", "$b[0]", "$b->x", "C::$b", "f()", "$b->f()", "C::f()", "1", "(1+2)", "($b)", "[]", "++$b", "new A()"} {
		for _, version := range []phpversion.Version{phpversion.PHP56, phpversion.PHP71, phpversion.PHP85} {
			t.Run(fmt.Sprintf("%s/%s", version, operand), func(t *testing.T) {
				valid := operand != "1" && operand != "(1+2)" && operand != "($b)" && operand != "[]" && operand != "++$b" && (operand != "new A()" || version.Below(phpversion.PHP70))
				src := "<?php $a =& " + operand + "; $after = 1;"
				// Permissive parsing deliberately retains removed legacy new references.
				if operand == "new A()" && !valid {
					f := parse(t, src, version)
					if len(f.Errors) == 0 {
						t.Fatal("accepted removed reference-new grammar")
					}
					if f = Parse("legacy.php", []byte(src), Options{Version: version, Permissive: true}); len(f.Errors) != 0 {
						t.Fatal(f.Errors)
					}
					if phpValid, ok := phpLint(t, version, src); ok && phpValid {
						t.Fatal("PHP accepted removed reference-new grammar")
					}
					return
				}
				checkVariableGrammar(t, src, version, valid)
			})
		}
	}
	checkVariableGrammar(t, "<?php $a =& ($b)->x; $after = 1;", phpversion.PHP85, true)
}

func BenchmarkWriteAndInterpolationGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat(`$s = "{$a->f($b + 1)[0]}"; $a++; $a += 1; [$a] = [1]; $a =& $b->f(); `, 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("variables.php", src, Options{Version: phpversion.PHP85})
	}
}
