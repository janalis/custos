package rules

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/analysis"
	"custos/internal/fix"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// FuzzRules runs every rule (and every fix) on fuzzed sources and fails on
// any rule panic (reported by the engine as an "internal" finding) or on a
// fix that cannot be applied. Seeded with the own fixtures.
func FuzzRules(f *testing.F) {
	seeds, _ := filepath.Glob("../../testdata/rules/*/*.php")
	for i, p := range seeds {
		if i%3 != 0 { // keep the seed corpus small
			continue
		}
		if b, err := os.ReadFile(p); err == nil && len(b) < 4096 {
			f.Add(string(b))
		}
	}
	f.Add("<?php class A { function f($x) { return $x?->y ?? throw new E(); } }")
	e, err := analysis.NewEngine(All(), analysis.Config{PHP: phpver.PHP84, EnableAll: true})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, src string) {
		for _, v := range []phpver.Version{phpver.PHP56, phpver.PHP84} {
			file := syntax.ParseBest("fuzz.php", []byte(src), syntax.Options{Version: v, Permissive: true})
			for _, fd := range e.Analyze(file) {
				if fd.Rule == "internal" {
					t.Fatalf("PHP %s: %s", v, fd.Message)
				}
				for _, fx := range fd.Fixes {
					fix.Apply(file.Src, fx.Edits())
				}
			}
		}
	})
}
