package syntax

import (
	"os/exec"
	"strings"
	"testing"

	"custos/internal/phpver"
)

func TestPostfixReceiverGrammar(t *testing.T) {
	php, _ := exec.LookPath("php")
	if php != "" {
		// The matrix includes attributes and nullsafe access, both PHP 8.0+.
		if err := exec.Command(php, "-n", "-r", "exit(PHP_VERSION_ID < 80000 ? 1 : 0);").Run(); err != nil {
			php = ""
		}
	}
	for _, receiver := range []struct {
		text  string
		valid bool
	}{
		{"1", false},
		{"1.5", false},
		{"function() {}", false},
		{"static function() {}", false},
		{"#[A] function() {}", false},
		{"#[A] static function() {}", false},
		{"(1)", true},
		{"(1.5)", true},
		{"(function() {})", true},
		{"(static function() {})", true},
		{"(new C)", true},
		{"'text'", true},
		{"\"$text\"", true},
		{"[1]", true},
		{"array(1)", true},
		{"true", true},
		{"null", true},
		{"$value", true},
	} {
		for _, operation := range []struct {
			text string
			op   string
		}{
			{"($child)", "("},
			{"[$child]", "["},
			{"->field", "->"},
			{"?->field", "?->"},
			{"::FIELD", "::"},
		} {
			t.Run(receiver.text+operation.text, func(t *testing.T) {
				src := "<?php $result = " + receiver.text + operation.text + "; $after = 1;"
				for _, version := range []phpver.Version{phpver.PHP80, phpver.PHP85} {
					for _, mode := range []string{"strict", "permissive", "adaptive"} {
						opt := Options{Version: version, Permissive: mode == "permissive"}
						var f *File
						if mode == "adaptive" {
							f = ParseBest("postfix.php", []byte(src), opt)
						} else {
							f = Parse("postfix.php", []byte(src), opt)
						}
						checkSpans(t, f)
						if (len(f.Errors) == 0) != receiver.valid || f.Version != version {
							t.Fatalf("%s/%s: errors = %v, version = %s", version, mode, f.Errors, f.Version)
						}
						if !receiver.valid && (len(f.Errors) != 1 || string(f.Src[f.Errors[0].Span.Start:f.Errors[0].Span.End]) != operation.op) {
							t.Fatalf("incorrect operator diagnostic: %v", f.Errors)
						}
						assertFollowingAssignment(t, f)
						expression := f.Stmts[0].(*ExprStmt).Expr.(*Assign).Value
						if got := string(f.Src[expression.Span().Start:expression.Span().End]); got != receiver.text+operation.text {
							t.Fatalf("lost operation during recovery: %q", got)
						}
						if operation.op == "(" || operation.op == "[" {
							childFound := false
							Inspect(expression, func(n Node) bool {
								if v, ok := n.(*Variable); ok && v.Name == "child" {
									childFound = true
								}
								return true
							})
							if !childFound {
								t.Fatal("lost argument or offset child")
							}
						}
					}
				}
				if php != "" {
					cmd := exec.Command(php, "-n", "-r", `try { token_get_all(stream_get_contents(STDIN), TOKEN_PARSE); } catch (ParseError $e) { exit(1); }`)
					cmd.Stdin = strings.NewReader(src)
					if valid := cmd.Run() == nil; valid != receiver.valid {
						t.Fatalf("PHP TOKEN_PARSE valid = %v, want %v", valid, receiver.valid)
					}
				}
			})
		}
	}
}

func TestPostfixNewReceiverCompatibility(t *testing.T) {
	for _, expression := range []string{"new C()->field", "new C()[0]", "new C()::FIELD", "new class {}->field"} {
		for _, version := range []phpver.Version{phpver.PHP84, phpver.PHP85} {
			f := parse(t, "<?php $result = "+expression+"; $after = 1;", version)
			if len(f.Errors) != 0 {
				t.Fatalf("%s: %s: %v", version, expression, f.Errors)
			}
			assertFollowingAssignment(t, f)
		}
	}
}

func TestLegacyPostfixReceiverGrammar(t *testing.T) {
	for _, version := range []phpver.Version{phpver.PHP53, phpver.PHP54, phpver.PHP56, phpver.PHP74} {
		for _, receiver := range []string{"1", "1.5", "function() {}", "static function() {}"} {
			if version == phpver.PHP53 && strings.HasPrefix(receiver, "static") {
				continue // Static closures were introduced in PHP 5.4.
			}
			for _, operation := range []string{"($child)", "[$child]", "->field", "::FIELD"} {
				src := "<?php $result = " + receiver + operation + "; $after = 1;"
				f := parse(t, src, version)
				if len(f.Errors) != 1 || f.Errors[0].Span.Start != uint32(len("<?php $result = ")+len(receiver)) {
					t.Fatalf("%s: %s: incorrect postfix diagnostic: %v", version, src, f.Errors)
				}
				assertFollowingAssignment(t, f)
				if valid, ok := phpLint(t, version, src); ok && valid {
					t.Fatalf("%s: PHP accepted %s", version, src)
				}
			}
		}
	}
}

func BenchmarkPostfixReceiverGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("$a = (function() {})($child); $b = 'text'[0]; $c = $object->method($child); $d = (1)::FIELD; ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("postfix.php", src, Options{Version: phpver.PHP85})
	}
}
