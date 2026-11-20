package project

import (
	"path/filepath"
	"testing"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/catalogue"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// The findings of a rule formatting code are taken from the cache only
// while each file's format settings are the same.
func TestFindingsCacheFormatSettings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "a.php"), "<?php\nfunction f($a,$b){\n  return $a;\n}\n")
	writeFile(t, filepath.Join(root, "src", "b.php"), "<?php\nfunction g() {}\n")
	dir := t.TempDir()
	settings := func(indent int) *format.Settings {
		r, err := format.Resolve(format.DefaultPreset, false, format.Layer{Source: "test", Values: map[string]any{"indent_size": indent}})
		if err != nil {
			t.Fatal(err)
		}
		return &r.Settings
	}
	two, four := settings(2), settings(4)
	with := func(s *format.Settings, ok bool) func(*Project) {
		return func(p *Project) {
			p.Engine, _ = analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84, EnableAll: true, Format: func(string) (any, bool) { return s, ok }})
		}
	}
	for _, step := range []struct {
		label     string
		configure func(*Project)
		kept      int64
	}{
		{"cold", with(four, true), 0},
		{"unchanged", with(four, true), 2},
		{"other settings", with(two, true), 0},
		{"same settings", with(two, true), 2},
		{"excluded", with(two, false), 0},
	} {
		if run := runFindings(t, root, dir, step.configure); run.kept != step.kept {
			t.Fatalf("%s: %d files taken from the cache, want %d", step.label, run.kept, step.kept)
		}
	}
}
