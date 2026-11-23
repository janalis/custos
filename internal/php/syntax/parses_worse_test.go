package syntax

import (
	"slices"
	"strings"
	"testing"

	phpversion "github.com/janalis/custos/internal/php/version"
)

// TestParsesWorse: a fix may not trade a compile error for a parse error, even
// when the number of errors stays the same.
func TestParsesWorse(t *testing.T) {
	parse := func(src string) *File {
		return Parse("w.php", []byte(src), Options{Version: phpversion.Max})
	}
	compile := parse("<?php f(a: 1, 2);")
	if len(compile.Errors) != 1 || compile.Recovered() {
		t.Fatalf("compile error expected: %+v", compile.Errors)
	}
	for src, want := range map[string]bool{
		"<?php f(a: 1, ,);":   true,  // as many errors, one recovered from
		"<?php f(a: 1, 2);":   false, // unchanged
		"<?php f(a: 1);":      false, // fixed
		"<?php f(a: 1, 2) +;": true,  // one more
	} {
		if got := parse(src).ParsesWorse(compile); got != want {
			t.Errorf("%s: ParsesWorse = %v", src, got)
		}
	}
}

// TestCappedCompileErrors: past MaxErrors the cap message is a compile
// error only when every error was one.
func TestCappedCompileErrors(t *testing.T) {
	many := strings.Repeat("f(a: 1, 2);\n", MaxErrors+2)
	if f := Parse("c.php", []byte("<?php\n"+many), Options{Version: phpversion.Max}); len(f.Errors) != MaxErrors+1 || f.Recovered() {
		t.Fatalf("compile errors only: %d, recovered %v", len(f.Errors), f.Recovered())
	}
	if f := Parse("c.php", []byte("<?php\n"+many+"f(,);"), Options{Version: phpversion.Max}); !f.Recovered() {
		t.Fatal("a dropped parse error makes the file recovered")
	}
}

// TestParseBestKeepsTargetCompileErrors: reparsing with the newest grammar
// keeps the compile errors only the target version raises.
func TestParseBestKeepsTargetCompileErrors(t *testing.T) {
	// The enum's error, raised only below 8.6, goes before the call's, which
	// every version raises.
	src := []byte("<?php enum E { public function __debugInfo(): array { return []; } }\nf(a: 1, 2);\n$f = strlen(?);\n")
	f := ParseBest("e.php", src, Options{Version: phpversion.PHP85})
	compile := 0
	for _, e := range f.Errors {
		if e.Compile {
			compile++
		}
	}
	if compile != 2 || f.Version != phpversion.PHP85 || f.RecoveryErrorCount() != 0 {
		t.Fatalf("errors %v", f.Errors)
	}
	if !slices.IsSortedFunc(f.Errors, func(a, b Error) int { return int(a.Span.Start) - int(b.Span.Start) }) {
		t.Fatalf("unsorted %v", f.Errors)
	}
}

// TestParseBestDeduplicatesCompileErrors: a compile error both grammars
// raise is kept once.
func TestParseBestDeduplicatesCompileErrors(t *testing.T) {
	f := ParseBest("d.php", []byte("<?php f(a: 1, 2);\n$f = strlen(?);\n"), Options{Version: phpversion.PHP85})
	if len(f.Errors) != 1 || !f.Errors[0].Compile {
		t.Fatalf("errors %v", f.Errors)
	}
}
