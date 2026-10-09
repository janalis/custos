package syntax

import (
	"strings"
	"testing"

	"custos/internal/phpver"
)

func TestNumericLiteralVersionsAndRecovery(t *testing.T) {
	cases := []struct {
		literal string
		version phpver.Version
		valid   bool
	}{
		{"078", phpver.PHP56, true},
		{"078", phpver.PHP70, false},
		{"079", phpver.PHP85, false},
		{"07_8", phpver.PHP85, false},
		{"077", phpver.PHP85, true},
		{"078.0", phpver.PHP85, true},
		{"079e1", phpver.PHP85, true},
		{"1_000", phpver.PHP73, false},
		{"1_000", phpver.PHP74, true},
		{"1_000.2_5e1_0", phpver.PHP74, true},
		{"1__0", phpver.PHP85, false},
		{"1_", phpver.PHP85, false},
		{"0x1_2", phpver.PHP73, false},
		{"0x1_2", phpver.PHP74, true},
		{"0Xf_F", phpver.PHP85, true},
		{"0x1_", phpver.PHP85, false},
		{"0x1__2", phpver.PHP85, false},
		{"0x_1", phpver.PHP85, false},
		{"0b1_0", phpver.PHP73, false},
		{"0b1_0", phpver.PHP74, true},
		{"0B0_1", phpver.PHP85, true},
		{"0b1_", phpver.PHP85, false},
		{"0b1__0", phpver.PHP85, false},
		{"0b_1", phpver.PHP85, false},
		{"0o17", phpver.PHP80, false},
		{"0o17", phpver.PHP81, true},
		{"0O7_1", phpver.PHP85, true},
		{"0o1_", phpver.PHP85, false},
		{"0o1__2", phpver.PHP85, false},
		{"0o_1", phpver.PHP85, false},
		{"0o18", phpver.PHP85, false},
		{"0_7", phpver.PHP73, false},
		{"0_7", phpver.PHP74, true},
		{"0_8", phpver.PHP74, false},
		{"0x8000_0000_0000_0000", phpver.PHP85, true},
		{"0b10000000000000000000000000000000000000000000000000000000000000000", phpver.PHP85, true},
		{"0o1000000000000000000000", phpver.PHP85, true},
	}
	for _, c := range cases {
		t.Run(c.literal+"/"+c.version.String(), func(t *testing.T) {
			src := "<?php $x = " + c.literal + "; $after = 1;"
			f := Parse("number.php", []byte(src), Options{Version: c.version})
			checkSpans(t, f)
			if valid := len(f.Errors) == 0; valid != c.valid {
				t.Fatalf("valid = %v, want %v: %v", valid, c.valid, f.Errors)
			}
			last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
			if variable, ok := last.Var.(*Variable); !ok || variable.Name != "after" {
				t.Fatalf("recovery lost following assignment: %v", last)
			}
			if valid, ok := phpLint(t, c.version, src); ok && valid != c.valid {
				t.Fatalf("PHP accepts = %v, want %v", valid, c.valid)
			}
		})
	}
}

func TestNumericTokenBoundaries(t *testing.T) {
	for _, literal := range []string{"0x1_", "0x1__2", "0b1_", "0o1_", "1_"} {
		src := []byte("<?php " + literal)
		tokens, errs := Lex(src, LexOptions{Version: phpver.PHP85})
		if len(errs) != 0 || len(tokens) != 3 {
			t.Fatalf("%s: tokens %v, errors %v", literal, tokens, errs)
		}
		separator := strings.IndexByte(literal, '_')
		if tokens[1].Kind != TLNumber || tokens[1].Start != 6 || tokens[1].End != uint32(6+separator) ||
			tokens[2].Kind != TString || tokens[2].Start != tokens[1].End || tokens[2].End != uint32(len(src)) {
			t.Fatalf("%s: wrong token boundaries: %v", literal, tokens)
		}
	}
	f := Parse("octal.php", []byte("<?php $x = 078;"), Options{Version: phpver.PHP85})
	if len(f.Errors) != 1 || f.Errors[0].Span.Start != 11 {
		t.Fatalf("invalid octal location: %v", f.Errors)
	}
	src := []byte("<?php " + strings.Repeat("$x = 078; ", MaxErrors+10))
	_, errs := Lex(src, LexOptions{Version: phpver.PHP85})
	if len(errs) != MaxErrors+1 {
		t.Fatalf("lexical errors not bounded: %d", len(errs))
	}
	f = Parse("many.php", src, Options{Version: phpver.PHP85})
	if len(f.Errors) != MaxErrors+1 || !strings.Contains(f.Errors[MaxErrors].Msg, "not reported") {
		t.Fatalf("syntax errors not capped: %v", f.Errors)
	}
}

func TestComparisonAssociativity(t *testing.T) {
	cases := []struct {
		expr  string
		valid bool
	}{
		{"$a < $b < $c", false},
		{"$a <= $b > $c", false},
		{"$a == $b == $c", false},
		{"$a === $b != $c", false},
		{"$a <=> $b !== $c", false},
		{"$a < $b < $c < $d", false},
		{"($a < $b) < $c", true},
		{"$a < ($b < $c)", true},
		{"($a == $b) == $c", true},
		{"$a == ($b == $c)", true},
		{"$a < $b == $c < $d", true},
		{"$a == $b && $c == $d", true},
		{"$a < $b && $c < $d", true},
		{"$a + $b < $c * $d", true},
	}
	for _, version := range []phpver.Version{phpver.PHP53, phpver.PHP56, phpver.PHP74, phpver.PHP85} {
		for _, c := range cases {
			src := "<?php $x = " + c.expr + "; $after = 1;"
			f := Parse("compare.php", []byte(src), Options{Version: version})
			checkSpans(t, f)
			if valid := len(f.Errors) == 0; valid != c.valid || len(f.Stmts) != 2 {
				t.Fatalf("%s/%s: errors %v, statements %d", version, c.expr, f.Errors, len(f.Stmts))
			}
			// The spaceship operator was introduced in PHP 7.0.
			if !strings.Contains(c.expr, "<=>") || version.AtLeast(phpver.PHP70) {
				if valid, ok := phpLint(t, version, src); ok && valid != c.valid {
					t.Fatalf("%s/%s: PHP accepts = %v, want %v", version, c.expr, valid, c.valid)
				}
			}
		}
	}
}

func TestNumericComparisonAdaptiveParsing(t *testing.T) {
	for _, literal := range []string{"1_000", "0o17"} {
		f := ParseBest("number.php", []byte("<?php $x = "+literal+";"), Options{Version: phpver.PHP73})
		if len(f.Errors) != 0 || f.Version != phpver.PHP73 {
			t.Fatalf("adaptive numeric grammar: version %s, errors %v", f.Version, f.Errors)
		}
	}
	f := ParseBest("compare.php", []byte("<?php $x = $a < $b < $c;"), Options{Version: phpver.PHP73})
	if len(f.Errors) != 1 {
		t.Fatalf("adaptive parse accepted comparison chain: %v", f.Errors)
	}
}

func BenchmarkNumericComparisons(b *testing.B) {
	src := []byte("<?php " + strings.Repeat("$x = 0x1_2 + 0b1_0 + 0o1_7 + 1_000.2_5e1_0; $y = $a < $b == $c < $d; ", 128))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("bench.php", src, Options{Version: phpver.PHP85})
	}
}
