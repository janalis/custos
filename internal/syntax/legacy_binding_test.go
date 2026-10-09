package syntax

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/phpver"
)

// bindingShape exposes dynamic-name ownership rather than merely source text.
func bindingShape(e Expr) string {
	switch n := e.(type) {
	case *Variable:
		if n.NameExpr != nil {
			return "var(" + bindingShape(n.NameExpr) + ")"
		}
		return "$" + n.Name
	case *ArrayDimFetch:
		return bindingShape(n.Var) + "[" + bindingShape(n.Dim) + "]"
	case *PropertyFetch:
		return bindingShape(n.Var) + "->{" + bindingShape(n.Name) + "}"
	case *MethodCall:
		return bindingShape(n.Var) + "->{" + bindingShape(n.Name) + "}()"
	case *StaticPropertyFetch:
		return bindingShape(n.Class) + "::prop(" + bindingShape(n.Name) + ")"
	case *StaticCall:
		return bindingShape(n.Class) + "::{" + bindingShape(n.Name) + "}()"
	case *FuncCall:
		return bindingShape(n.Name) + "()"
	case *Name:
		return n.Value
	case *Literal:
		return n.Raw
	}
	return "_"
}

func TestLegacyDynamicNameBinding(t *testing.T) {
	for _, tc := range []struct{ src, legacy, modern string }{
		{"$$a[0]", "var($a[0])", "var($a)[0]"},
		{"$$a[0][1]", "var($a[0][1])", "var($a)[0][1]"},
		{"$$$a[0]", "var(var($a[0]))", "var(var($a))[0]"},
		{"$o->$a[0]", "$o->{$a[0]}", "$o->{$a}[0]"},
		{"$o->$a[0]()", "$o->{$a[0]}()", "$o->{$a}[0]()"},
		{"C::$a[0]()", "C::{$a[0]}()", "C::prop($a)[0]()"},
		{"C::$a[0][1]()", "C::{$a[0][1]}()", "C::prop($a)[0][1]()"},
		{"C::$a[0]", "C::prop($a)[0]", "C::prop($a)[0]"},
		{"C::$a[0][1]", "C::prop($a)[0][1]", "C::prop($a)[0][1]"},
		{"C::$a()", "C::{$a}()", "C::{$a}()"},
		{"C::$a", "C::prop($a)", "C::prop($a)"},
		{"${$a}[0]", "var($a)[0]", "var($a)[0]"},
		{"$o->{$a}[0]", "$o->{$a}[0]", "$o->{$a}[0]"},
		{"$$a{0}", "var($a[0])", "var($a)[0]"},
		{"$$a[]", "var($a[_])", "var($a)[_]"},
	} {
		for _, version := range []phpver.Version{phpver.PHP53, phpver.PHP56, phpver.PHP70, phpver.PHP74} {
			t.Run(fmt.Sprintf("%s/%s", tc.src, version), func(t *testing.T) {
				src := "<?php " + tc.src + "; $after = 1;"
				f := parse(t, src, version)
				if len(f.Errors) != 0 {
					t.Fatal(f.Errors)
				}
				want := tc.modern
				if version.Below(phpver.PHP70) {
					want = tc.legacy
				}
				if got := bindingShape(firstExpr(t, f)); got != want {
					t.Fatalf("got %s, want %s", got, want)
				}
				// Empty offsets are accepted by the grammar and rejected at compile time.
				if valid, ok := phpLint(t, version, src); ok && !valid && tc.src != "$$a[]" {
					t.Fatal("PHP rejects fixture")
				}
			})
		}
	}
}

func TestLegacyDynamicNameRecoveryAndDepth(t *testing.T) {
	for _, src := range []string{"$$; $after = 1;", "$$a[; $after = 1;", "$o->$a[; $after = 1;", "C::$a[; $after = 1;"} {
		f := parse(t, "<?php "+src, phpver.PHP56)
		if len(f.Errors) == 0 {
			t.Fatal("missing error")
		}
		last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
		if last.Var.(*Variable).Name != "after" {
			t.Fatal("lost following assignment")
		}
	}
	for _, version := range []phpver.Version{phpver.PHP56, phpver.PHP85} {
		f := Parse("deep.php", []byte("<?php "+strings.Repeat("$", MaxDepth+20)+"a;"), Options{Version: version})
		if len(f.Errors) == 0 || !strings.Contains(f.Errors[len(f.Errors)-1].Msg, "nesting deeper") {
			t.Fatal(f.Errors)
		}
	}
}

func BenchmarkDynamicNameBinding(b *testing.B) {
	for _, version := range []phpver.Version{phpver.PHP56, phpver.PHP85} {
		b.Run(version.String(), func(b *testing.B) {
			src := []byte("<?php " + strings.Repeat("$$a[0]; $o->$a[0](); C::$a[0][1](); C::$a[0][1]; ", 128))
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			for b.Loop() {
				Parse("binding.php", src, Options{Version: version})
			}
		})
	}
}
