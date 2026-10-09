package syntax

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"custos/internal/phpver"
)

func TestDeclarationListGrammar(t *testing.T) {
	php, _ := exec.LookPath("php")
	if php != "" {
		if err := exec.Command(php, "-n", "-r", "exit(PHP_VERSION_ID < 80000 ? 1 : 0);").Run(); err != nil {
			php = ""
		}
	}
	for _, tc := range []struct {
		declaration string
		minimum     phpver.Version
		errorToken  string
	}{
		{"#[] function f() { $inside = 1; }", phpver.PHP80, "]"},
		{"#[ /* empty */ ] function f() { $inside = 1; }", phpver.PHP80, "]"},
		{"#[A] #[] function f() { $inside = 1; }", phpver.PHP80, "]"},
		{"#[A] function f() { $inside = 1; }", phpver.PHP80, ""},
		{"#[A, B(),] function f() { $inside = 1; }", phpver.PHP80, ""},
		{"$f = function() use () { $inside = 1; };", phpver.PHP53, ")"},
		{"$f = function() use ( /* empty */ ) { $inside = 1; };", phpver.PHP53, ")"},
		{"$f = function() { $inside = 1; };", phpver.PHP53, ""},
		{"$f = function() use ($x, &$y) { $inside = 1; };", phpver.PHP53, ""},
		{"$f = function() use ($x,) { $inside = 1; };", phpver.PHP80, ""},
		{"namespace; $inside = 1;", phpver.PHP53, ";"},
		{"namespace /* empty */ ; $inside = 1;", phpver.PHP53, ";"},
		{"namespace { $inside = 1; }", phpver.PHP53, ""},
		{"namespace App; $inside = 1;", phpver.PHP53, ""},
		{"namespace App\\Feature; $inside = 1;", phpver.PHP53, ""},
		{"namespace App { $inside = 1; }", phpver.PHP53, ""},
	} {
		t.Run(tc.declaration, func(t *testing.T) {
			src := "<?php " + tc.declaration + " $after = 1;"
			for _, version := range []phpver.Version{tc.minimum, phpver.PHP85} {
				for _, mode := range []string{"strict", "permissive", "adaptive"} {
					t.Run(fmt.Sprintf("%s/%s", version, mode), func(t *testing.T) {
						opt := Options{Version: version, Permissive: mode == "permissive"}
						var f *File
						if mode == "adaptive" {
							f = ParseBest("declaration-lists.php", []byte(src), opt)
						} else {
							f = Parse("declaration-lists.php", []byte(src), opt)
						}
						checkSpans(t, f)
						if f.Version != version {
							t.Fatalf("version = %s, want %s", f.Version, version)
						}
						if tc.errorToken == "" {
							if len(f.Errors) != 0 {
								t.Fatalf("valid declaration: %v", f.Errors)
							}
						} else if len(f.Errors) != 1 || string(f.Src[f.Errors[0].Span.Start:f.Errors[0].Span.End]) != tc.errorToken {
							t.Fatalf("errors = %v, want one diagnostic on %q", f.Errors, tc.errorToken)
						}
						retained := map[string]bool{}
						for _, stmt := range f.Stmts {
							Inspect(stmt, func(n Node) bool {
								if assign, ok := n.(*Assign); ok {
									if variable, ok := assign.Var.(*Variable); ok {
										retained[variable.Name] = true
									}
								}
								return true
							})
						}
						if !retained["inside"] || !retained["after"] {
							t.Fatalf("lost declaration body or following statement: %v", retained)
						}
					})
				}
			}
			if php != "" {
				// TOKEN_PARSE checks grammar without imposing compile-time
				// restrictions on code outside a braced namespace.
				cmd := exec.Command(php, "-n", "-r", `try { token_get_all(stream_get_contents(STDIN), TOKEN_PARSE); } catch (ParseError $e) { exit(1); }`)
				cmd.Stdin = strings.NewReader(src)
				if valid := cmd.Run() == nil; valid != (tc.errorToken == "") {
					t.Fatalf("PHP TOKEN_PARSE valid = %v, want %v", valid, tc.errorToken == "")
				}
			}
		})
	}
}

func BenchmarkDeclarationListGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("namespace App { #[A, B(),] function f() { $f = function() use ($x, &$y,) { $inside = 1; }; } } ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("declaration-lists.php", src, Options{Version: phpver.PHP85})
	}
}
