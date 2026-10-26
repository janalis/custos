package config

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
)

// TestCSFixerStyle: without a comparisonStyle in custos.json, the style
// comes from the PHP-CS-Fixer configuration.
func TestCSFixerStyle(t *testing.T) {
	const head = "<?php\nreturn (new PhpCsFixer\\Config())->setRules(["
	for _, tc := range []struct {
		name, file, rules string
		want              analysis.ComparisonStyle
		found             bool
	}{
		{"symfony set", ".php-cs-fixer.dist.php", `'@Symfony' => true`, analysis.StyleYoda, true},
		{"risky set", ".php-cs-fixer.php", `"@PhpCsFixer:risky" => TRUE`, analysis.StyleYoda, true},
		{"set off", ".php-cs-fixer.dist.php", `'@Symfony' => false`, analysis.StyleRegular, false},
		{"other set", ".php-cs-fixer.dist.php", `'@PSR12' => true, 'array_syntax' => ['syntax' => 'short']`, analysis.StyleRegular, false},
		{"rule off", ".php-cs-fixer.dist.php", `'@Symfony' => true, 'yoda_style' => false`, analysis.StyleRegular, true},
		{"rule regular", ".php-cs-fixer.dist.php", `'@Symfony' => true, 'yoda_style' => ['equal' => false, 'identical' => \false, 'less_and_greater' => null]`, analysis.StyleRegular, true},
		{"rule partly yoda", ".php-cs-fixer.dist.php", `'yoda_style' => ['equal' => false, 'identical' => true]`, analysis.StyleYoda, true},
		{"rule on", ".php-cs-fixer.dist.php", `'yoda_style' => true, 0 => 'x'`, analysis.StyleYoda, true},
		{"broken rule list", ".php-cs-fixer.dist.php", `'yoda_style' => [, 'equal' => false]`, analysis.StyleYoda, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte(head+tc.rules+"]);\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			got, found := csFixerStyleRead(func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(dir, name)) })
			if got != tc.want || found != tc.found {
				t.Fatalf("got %v %v, want %v %v", got, found, tc.want, tc.found)
			}
		})
	}
	// The local file wins over the distributed one, and custos.json over both.
	dir := t.TempDir()
	for name, body := range map[string]string{
		".php-cs-fixer.php":      head + "'yoda_style' => false]);",
		".php-cs-fixer.dist.php": head + "'@Symfony' => true]);",
		"composer.json":          `{}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if c, err := Load(dir); err != nil || c.ComparisonStyle != analysis.StyleRegular {
		t.Fatalf("local file: %v %v", c, err)
	}
	if err := os.Remove(filepath.Join(dir, ".php-cs-fixer.php")); err != nil {
		t.Fatal(err)
	}
	if c, err := Load(dir); err != nil || c.ComparisonStyle != analysis.StyleYoda {
		t.Fatalf("dist file: %v %v", c, err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(`{"comparisonStyle": "regular"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if c, err := Load(dir); err != nil || c.ComparisonStyle != analysis.StyleRegular {
		t.Fatalf("custos.json: %v %v", c, err)
	}
}
