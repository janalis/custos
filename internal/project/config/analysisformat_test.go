package config

import (
	"path/filepath"
	"testing"

	"github.com/janalis/custos/internal/format"
)

func TestAnalysisFormat(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{
		FileName:                   `{"format": {"exclude": ["legacy"], "settings": {"indent_size": 2}}}`,
		"src/.editorconfig":        "[*.php]\nindent_size = 3\n",
		"broken/.editorconfig/.ok": "",
	})
	settings := c.Analysis().Format
	if v, ok := settings(filepath.Join(dir, "a.php")); !ok || v.(*format.Settings).IndentSize != 2 {
		t.Fatalf("custos.json settings not applied: %#v %v", v, ok)
	}
	if _, ok := settings(filepath.Join(dir, "legacy", "a.php")); ok {
		t.Fatal("an excluded file is formatted")
	}
	if _, ok := settings(filepath.Join(dir, "broken", "a.php")); ok {
		t.Fatal("a file whose .editorconfig cannot be read is formatted")
	}
	// Relative paths are taken from base.
	rooted := c.AnalysisFormat(dir)
	if v, ok := rooted("src/a.php"); !ok || v.(*format.Settings).IndentSize != 2 {
		t.Fatalf("relative path: %#v %v", v, ok)
	}
	if _, ok := rooted("legacy/a.php"); ok {
		t.Fatal("a relative excluded path is formatted")
	}
	if v, ok := rooted(filepath.Join(dir, "a.php")); !ok || v == nil {
		t.Fatal("absolute path under a base")
	}
}

func TestAnalysisFormatEditorConfig(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{"src/.editorconfig": "[*.php]\nindent_size = 3\n"})
	v, ok := c.AnalysisFormat("")(filepath.Join(dir, "src", "a.php"))
	if !ok || v.(*format.Settings).IndentSize != 3 {
		t.Fatalf(".editorconfig not applied: %#v %v", v, ok)
	}
}

func TestAnalysisFormatInvalid(t *testing.T) {
	c := &Config{Root: t.TempDir()}
	c.Format = FormatConfig{Preset: AutoPreset, imported: func() toolFormat {
		return toolFormat{preset: format.DefaultPreset, layers: []format.Layer{{Source: "tool", Values: map[string]any{"nope": 1}}}}
	}}
	if _, ok := c.AnalysisFormat("")(filepath.Join(c.Root, "a.php")); ok {
		t.Fatal("invalid settings format files")
	}
}
