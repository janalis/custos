package syntax

import (
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestHeredocIndentation(t *testing.T) {
	cases := []struct {
		body, closing string
		valid         bool
	}{
		{"  text\n", "  ", true},
		{"\ttext\n", "\t", true},
		{" text\n", "  ", false},
		{"  \ttext\n", "  ", true},
		{" \ttext\n", "  ", false},
		{"\t text\n", "  ", false},
		{"  text\n", "\t", false},
		{"\t text\n", "\t", true},
		{"  text\n", " \t", false},
		{"\ttext\n", "\t ", false},
		{"\n \n   \n  text\n", "  ", true},
		{"\t\n", "  ", false},
		{" \n", "\t", false},
		{"\t\n", "\t\t", true},
		{"", "  ", true},
		{"\t text\n", "", true},
	}
	for _, c := range cases {
		for _, label := range []string{"EOT", "'EOT'"} {
			for _, newline := range []string{"\n", "\r\n", "\r"} {
				src := "<?php $s = <<<" + label + "\n" + c.body + c.closing + "EOT;\n$after = 1;"
				src = strings.ReplaceAll(src, "\n", newline)
				f := parse(t, src, phpversion.PHP85)
				if got := len(f.Errors) == 0; got != c.valid {
					t.Fatalf("%q: valid %v, want %v: %v", src, got, c.valid, f.Errors)
				}
				last := f.Stmts[len(f.Stmts)-1].(*ExprStmt).Expr.(*Assign)
				if last.Var.(*Variable).Name != "after" {
					t.Fatal("lost following statement")
				}
				if valid, ok := phpLint(t, phpversion.PHP85, src); ok && valid != c.valid {
					t.Fatalf("%q: PHP accepts %v, want %v", src, valid, c.valid)
				}
			}
		}
	}
}

func TestInterpolatedHeredocIndentation(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{"  $a\n  {$o->{\n'x'\n}}\n", true},
		{"  $a\n $b\n", false},
		{"  $a\n$b\n", false},
		{"$a\n", false},
		{"  {$o->{<<<INNER\nx\nINNER\n}}\n", true},
		{"  \\n\n  $a\n", true},
		{"  \\\n x\n", false},
		{"  \\\r\n x\r\n", false},
	} {
		src := "<?php $s = <<<OUTER\n" + tc.body + "  OUTER;\n"
		f := parse(t, src, phpversion.PHP85)
		if (len(f.Errors) == 0) != tc.valid {
			t.Fatalf("%q: %v", src, f.Errors)
		}
		if valid, ok := phpLint(t, phpversion.PHP85, src); ok && valid != tc.valid {
			t.Fatalf("%q: PHP accepts %v", src, valid)
		}
	}
	for _, version := range []phpversion.Version{phpversion.PHP72, phpversion.PHP73} {
		src := "<?php $s = <<<EOT\n  text\n  EOT;\n"
		f := parse(t, src, version)
		if (len(f.Errors) == 0) != version.AtLeast(phpversion.PHP73) {
			t.Fatal(f.Errors)
		}
	}
}

func TestHeredocIndentErrorPositionAndCap(t *testing.T) {
	src := "<?php $s = <<<EOT\n x\n  EOT;"
	_, errors := Lex([]byte(src), LexOptions{Version: phpversion.PHP85})
	if len(errors) != 1 || errors[0].Pos != uint32(strings.Index(src, "x")) {
		t.Fatal(errors)
	}
	src = "<?php " + strings.Repeat("$s = <<<EOT\n x\n  EOT;\n", MaxErrors+10)
	_, errors = Lex([]byte(src), LexOptions{Version: phpversion.PHP85})
	if len(errors) != MaxErrors+1 {
		t.Fatal(len(errors))
	}
}

func BenchmarkHeredocIndentation(b *testing.B) {
	for _, line := range []string{"    literal text\n", "    $variable text\n"} {
		b.Run(strings.TrimSpace(line), func(b *testing.B) {
			src := []byte("<?php $s = <<<EOT\n" + strings.Repeat(line, 1024) + "    EOT;\n")
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			for b.Loop() {
				Lex(src, LexOptions{Version: phpversion.PHP85})
			}
		})
	}
}
