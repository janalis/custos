package syntax

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/phpver"
)

func TestDNFTypeGrammar(t *testing.T) {
	for _, tc := range []struct {
		typ   string
		valid bool
	}{
		{"A", true},
		{"?A", true},
		{"A&B", true},
		{"A|B", true},
		{"(A&B)|C", true},
		{"A|(B&C)", true},
		{"(A&B)|(C&D)|null", true},
		{"(A)", false},
		{"(A)|B", false},
		{"(A&B)", false},
		{"?(A&B)", false},
		{"(A&B)&C", false},
		{"A&(B&C)", false},
		{"((A&B)&C)|D", false},
		{"(A&(B&C))|D", false},
	} {
		for _, declaration := range []string{
			"function f(%s $value) {}",
			"function f(): %s {}",
			"class C { public %s $value; }",
			"class C { const %s VALUE = null; }",
		} {
			src := "<?php " + fmt.Sprintf(declaration, tc.typ) + " $after = 1;"
			t.Run(fmt.Sprintf(declaration, tc.typ), func(t *testing.T) {
				f := parse(t, src, phpver.PHP85)
				if got := len(f.Errors) == 0; got != tc.valid {
					t.Fatalf("valid = %v, want %v: %v", got, tc.valid, f.Errors)
				}
				last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
				if last.Var.(*Variable).Name != "after" {
					t.Fatal("lost following assignment")
				}
				// A null initializer is a compile-time type error for some valid
				// constant types; function parameters isolate grammar acceptance.
				if declaration == "function f(%s $value) {}" {
					if valid, ok := phpLint(t, phpver.PHP85, src); ok && valid != tc.valid {
						t.Fatalf("PHP valid = %v, want %v", valid, tc.valid)
					}
				}
			})
		}
	}
}

func TestDNFTypeASTAndReferenceBinding(t *testing.T) {
	f := parse(t, "<?php function f((A&B)|C &$value, A&B &...$rest): D|(E&F) {}", phpver.PHP85)
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	fn := f.Stmts[0].(*Function)
	first := fn.Params[0]
	u := first.Type.(*UnionType)
	i := u.Types[0].(*IntersectionType)
	if len(i.Types) != 2 || i.Types[0].(*Name).Value != "A" || i.Types[1].(*Name).Value != "B" || u.Types[1].(*Name).Value != "C" || !first.ByRef {
		t.Fatalf("lost grouped intersection or reference: %+v", first)
	}
	if got := string(f.Src[i.Span().Start:i.Span().End]); got != "(A&B)" {
		t.Fatalf("group span = %q", got)
	}
	second := fn.Params[1]
	if len(second.Type.(*IntersectionType).Types) != 2 || !second.ByRef || !second.Variadic {
		t.Fatalf("lost intersection or variadic reference: %+v", second)
	}
	if len(fn.ReturnType.(*UnionType).Types[1].(*IntersectionType).Types) != 2 {
		t.Fatal("lost return intersection")
	}
}

func TestDNFTypeDepthLimit(t *testing.T) {
	src := "<?php function f(" + strings.Repeat("(", MaxDepth+10) + "A" + strings.Repeat(")", MaxDepth+10) + " $value) {}"
	f := Parse("deep-type.php", []byte(src), Options{Version: phpver.PHP85})
	if len(f.Errors) == 0 || !strings.Contains(f.Errors[len(f.Errors)-1].Msg, "nesting deeper") {
		t.Fatal(f.Errors)
	}
}

func TestAdaptiveGrammarValidation(t *testing.T) {
	for _, src := range []string{
		"<?php function f((A) $value) {}",
		"<?php ++($value);",
	} {
		f := ParseBest("invalid.php", []byte(src), Options{Version: phpver.PHP74})
		checkSpans(t, f)
		if len(f.Errors) == 0 || f.Version != phpver.PHP74 {
			t.Fatalf("adaptive parsing lost errors or target version: %v, %s", f.Errors, f.Version)
		}
	}
}

func TestPrefixIncrementGrammar(t *testing.T) {
	for _, tc := range []struct {
		operand string
		valid   bool
		compile bool
	}{
		{"$value", true, false},
		{"$$value", true, false},
		{"$value[0]", true, false},
		{"$value->field", true, false},
		{"C::$value", true, false},
		{"($value)->field", true, false},
		{"($value)[0]", true, false},
		{"f()", true, true},
		{"$value->f()", true, true},
		{"C::f()", true, true},
		{"1", false, false},
		{"'text'", false, false},
		{"($value)", false, false},
		{"[]", false, false},
		{"list($value)", false, false},
		{"-$value", false, false},
		{"++$value", false, false},
		{"$value++", false, false},
	} {
		for _, op := range []string{"++", "--"} {
			t.Run(op+tc.operand, func(t *testing.T) {
				src := "<?php " + op + tc.operand + "; $after = 1;"
				for _, version := range []phpver.Version{phpver.PHP53, phpver.PHP85} {
					f := parse(t, src, version)
					if got := len(f.Errors) == 0; got != tc.valid {
						t.Fatalf("%s: valid = %v, want %v: %v", version, got, tc.valid, f.Errors)
					}
					n := firstExpr(t, f).(*IncDec)
					if !n.Prefix || n.Op.Kind.String() != op || n.Span().Start != 6 || n.Var.Span().Start != 8 || n.Span().End != uint32(8+len(tc.operand)) {
						t.Fatalf("incorrect increment metadata: %+v", n)
					}
					last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
					if last.Var.(*Variable).Name != "after" {
						t.Fatal("lost following assignment")
					}
					if valid, ok := phpLint(t, version, src); ok && valid != (tc.valid && !tc.compile) {
						t.Fatalf("%s: PHP valid = %v, compile-time rejection = %v", version, valid, tc.compile)
					}
				}
			})
		}
	}
}

func BenchmarkTypeAndIncrementGrammar(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("function f((A&B)|C &$value, D&E &...$rest): F|(G&H) {} ++$value; --($value)->field; ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("grammar.php", src, Options{Version: phpver.PHP85})
	}
}
