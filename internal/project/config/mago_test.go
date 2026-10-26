package config

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
)

// TestMagoStyle: without a comparisonStyle in custos.json or a PHP-CS-Fixer
// configuration, Mago's yoda-conditions rule decides.
func TestMagoStyle(t *testing.T) {
	for _, tc := range []struct {
		name, file, body string
		want             analysis.ComparisonStyle
		found            bool
	}{
		{"require", "mago.toml", "[linter.rules]\nyoda-conditions = { enabled = true, level = \"warning\", mode = \"require\" }\n", analysis.StyleYoda, true},
		{"default mode", "mago.dist.toml", "[linter.rules]\n\"yoda-conditions\" = {enabled=true}\n", analysis.StyleYoda, true},
		{"deny", "mago.toml", "[linter.rules]\nyoda-conditions = { enabled = true, mode = 'deny' } # keep $x === null\n", analysis.StyleRegular, true},
		{"table", "mago.toml", "[linter]\n[linter.rules.yoda-conditions]\nenabled = true # on\nmode = \"deny\"\n[formatter]\nmode = \"require\"\n", analysis.StyleRegular, true},
		{"disabled", "mago.toml", "[linter.rules]\nyoda-conditions = { enabled = false }\n", analysis.StyleRegular, false},
		{"off by default", "mago.toml", "[linter.rules]\nyoda-conditions = { mode = \"require\" }\n", analysis.StyleRegular, false},
		{"other rule", "mago.toml", "# yoda-conditions = { enabled = true }\n[linter.rules]\nno-is-null = { enabled = true }\n[other]\nyoda-conditions = { enabled = true }\nname = \"a#b\"\n", analysis.StyleRegular, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			got, found := magoStyleRead(func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(dir, name)) })
			if got != tc.want || found != tc.found {
				t.Fatalf("got %v %v, want %v %v", got, found, tc.want, tc.found)
			}
		})
	}
	// PHP-CS-Fixer comes first; Mago applies without it.
	dir := t.TempDir()
	for name, body := range map[string]string{
		"mago.toml":              "[linter.rules]\nyoda-conditions = { enabled = true }\n",
		".php-cs-fixer.dist.php": "<?php return (new PhpCsFixer\\Config())->setRules(['yoda_style' => false]);",
		"composer.json":          `{}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if c, err := Load(dir); err != nil || c.ComparisonStyle != analysis.StyleRegular {
		t.Fatalf("php-cs-fixer first: %v %v", c, err)
	}
	if err := os.Remove(filepath.Join(dir, ".php-cs-fixer.dist.php")); err != nil {
		t.Fatal(err)
	}
	if c, err := Load(dir); err != nil || c.ComparisonStyle != analysis.StyleYoda {
		t.Fatalf("mago: %v %v", c, err)
	}
}
