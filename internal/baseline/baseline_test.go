package baseline

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/report"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	php := filepath.Join(dir, "a.php")
	if err := os.WriteFile(php, []byte("<?php\n$a = !!$b;\n$c = !!$d;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []report.Item{
		{Path: php, Line: 2, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 3, Rule: "NestedNotOperators", Message: "m"},
	}
	bl := filepath.Join(dir, "custos-baseline.json")
	if n, err := Write(bl, items); err != nil || n != 2 {
		t.Fatalf("write: %d %v", n, err)
	}
	// Code moves down two lines and a new finding appears.
	if err := os.WriteFile(php, []byte("<?php\n// x\n// y\n$a = !!$b;\n$c = !!$d;\n$e = !!$f;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := []report.Item{
		{Path: php, Line: 4, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 5, Rule: "NestedNotOperators", Message: "m"},
		{Path: php, Line: 6, Rule: "NestedNotOperators", Message: "m"},
	}
	b, err := Load(bl)
	if err != nil {
		t.Fatal(err)
	}
	rest, suppressed := b.Filter(moved)
	if suppressed != 2 || len(rest) != 1 || rest[0].Line != 6 {
		t.Fatalf("suppressed=%d rest=%+v", suppressed, rest)
	}
}
