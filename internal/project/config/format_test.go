package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/format/editorconfig"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func loadFormat(t *testing.T, files map[string]string) (*Config, string) {
	t.Helper()
	dir := t.TempDir()
	files["composer.json"] = "{}"
	writeFiles(t, dir, files)
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return c, dir
}

func TestFormatSectionValidation(t *testing.T) {
	for _, tc := range []struct{ json, want string }{
		{`{"format": {"preset": "nope"}}`, `custos.json: format: unknown preset "nope"`},
		{`{"format": {"risky": ["concat_space"]}}`, `custos.json: format: risky entry "concat_space" is not a risky setting`},
		{`{"format": {"risky": ["nope"]}}`, `risky entry "nope"`},
		{`{"format": {"settings": {"nope": 1}}}`, `format: custos.json: unknown format setting "nope"`},
		{`{"format": {"settings": {"concat_space": "two"}}}`, `format: custos.json: concat_space: want one of`},
	} {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{FileName: tc.json})
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want %q", tc.json, err, tc.want)
		}
	}
}

func TestFormatDefaults(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{})
	if c.Format.Preset != AutoPreset || !c.Format.EditorConfig || c.Format.AllowRisky {
		t.Fatalf("defaults: %+v", c.Format)
	}
	r, err := c.FormatResolver(FormatFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if name, source := r.Preset(); name != format.DefaultPreset || source != "default" {
		t.Fatalf("preset %s from %s", name, source)
	}
	res, err := c.FormatSettings(filepath.Join(dir, "a.php"), FormatFlags{})
	if err != nil || res.Preset != format.DefaultPreset {
		t.Fatalf("%v %+v", err, res)
	}
	if _, err := c.FormatSettings("a.php", FormatFlags{Preset: "nope"}); err == nil {
		t.Fatal("unknown preset flag accepted")
	}
}

// TestFormatImport: with the preset "auto", the style comes from Pint, else
// PHP-CS-Fixer, else Mago.
func TestFormatImport(t *testing.T) {
	const fixer = "<?php\nreturn (new PhpCsFixer\\Config())"
	for _, tc := range []struct {
		name   string
		files  map[string]string
		preset string
		source string
		values map[string]any
	}{
		{"none", nil, "per", "default", nil},
		{"pint preset", map[string]string{"pint.json": `{"preset": "PSR12"}`}, "psr12", "pint.json", nil},
		{"pint empty", map[string]string{"pint.json": `{"preset": "empty"}`}, "psr1", "pint.json", nil},
		{"pint default", map[string]string{"pint.json": `{}`}, "laravel", "pint.json", nil},
		{"pint unknown", map[string]string{"pint.json": `{"preset": "mine"}`}, "laravel", "pint.json", nil},
		{"pint broken", map[string]string{"pint.json": `{`}, "laravel", "pint.json", nil},
		{"pint first", map[string]string{"pint.json": `{"preset": "per"}`, ".php-cs-fixer.php": fixer + "->setRules(['@Symfony' => true]);", "mago.toml": "[formatter]\npreset = \"drupal\"\n"}, "per", "pint.json", nil},
		{"fixer default", map[string]string{".php-cs-fixer.dist.php": fixer + ";"}, "psr12", ".php-cs-fixer.dist.php", nil},
		{"fixer local first", map[string]string{".php-cs-fixer.php": fixer + "->setRules(['@PhpCsFixer' => true]);", ".php-cs-fixer.dist.php": fixer + "->setRules(['@PSR12' => true]);"}, "symfony", ".php-cs-fixer.php", nil},
		{"fixer last set", map[string]string{".php-cs-fixer.php": fixer + "->setRules(['@PSR2' => true, '@PER-CS2.0' => true, '@PER-CS2.0:risky' => true, '@Symfony:risky' => true, 'array_syntax' => true, '@Symfony' => false]);"}, "per", ".php-cs-fixer.php", nil},
		{"fixer per and psr1", map[string]string{".php-cs-fixer.php": fixer + "->setRules(['@PER' => true, '@PSR1' => true]);"}, "psr1", ".php-cs-fixer.php", nil},
		{"fixer tabs crlf", map[string]string{".php-cs-fixer.php": fixer + "->setIndent(\"\\t\")->setLineEnding(\"\\r\\n\")->setRules(['@Symfony' => true]);"}, "symfony", ".php-cs-fixer.php", map[string]any{"indent_style": "tab", "line_ending": "crlf"}},
		{"fixer spaces lf", map[string]string{".php-cs-fixer.php": fixer + "->setIndent('  ')->setLineEnding(\"\\n\");"}, "psr12", ".php-cs-fixer.php", map[string]any{"indent_style": "space", "indent_size": 2, "line_ending": "lf"}},
		{"fixer unmapped", map[string]string{".php-cs-fixer.php": "<?php\n$c->setIndent(' \\t')->setLineEnding(PHP_EOL)->setLineEnding('\\r')->setIndent()->setIndent(...)->setIndent(<<<X\n  \nX)->{$m}('x')->setIndent('');"}, "psr12", ".php-cs-fixer.php", nil},
		{"mago", map[string]string{"mago.toml": "[linter]\nprint-width = 10\n[formatter]\npreset = \"Laravel\" # style\nprint_width = 100\n\"tab-width\" = 2\nuse-tabs = true\nsingle-quote = true\nbad\nprint-width-x = 3\ntab-width = x\n"}, "laravel", "mago.toml", map[string]any{"line_length": 100, "indent_size": 2, "tab_width": 2, "indent_style": "tab"}},
		{"mago dist", map[string]string{"mago.dist.toml": "[formatter]\npreset = \"psr-12\"\nuse-tabs = false\n"}, "psr12", "mago.dist.toml", map[string]any{"indent_style": "space"}},
		{"mago unknown preset", map[string]string{"mago.toml": "[formatter]\npreset = \"x\"\n"}, "per", "mago.toml", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{".editorconfig": "root = true\n"}
			for k, v := range tc.files {
				files[k] = v
			}
			c, dir := loadFormat(t, files)
			r, err := c.FormatResolver(FormatFlags{})
			if err != nil {
				t.Fatal(err)
			}
			if name, source := r.Preset(); name != tc.preset || source != tc.source {
				t.Fatalf("preset %s from %s, want %s from %s", name, source, tc.preset, tc.source)
			}
			res, err := r.Settings(filepath.Join(dir, "a.php"))
			if err != nil {
				t.Fatal(err)
			}
			for id, want := range tc.values {
				d, _ := format.Lookup(id)
				if got := d.Get(&res.Settings); got != want || res.Source[id] != tc.source {
					t.Errorf("%s = %v from %s, want %v from %s", id, got, res.Source[id], want, tc.source)
				}
			}
			for id, source := range res.Source {
				if _, ok := tc.values[id]; !ok && source == tc.source && source != "default" {
					t.Errorf("%s unexpectedly from %s", id, source)
				}
			}
		})
	}
	if got := importFormat(nil); got.preset != format.DefaultPreset {
		t.Fatal(got)
	}
}

// TestFormatLayers: the preset, the imported style, .editorconfig,
// custos.json and the command line apply in this order.
func TestFormatLayers(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{
		"mago.toml":         "[formatter]\nprint-width = 90\ntab-width = 3\nuse-tabs = true\n",
		".editorconfig":     "root = true\n[*]\nindent_style = space\nmax_line_length = 100\nend_of_line = crlf\ninsert_final_newline = false\ntrim_trailing_whitespace = false\n[*.php]\nindent_size = 2\n[Makefile]\nindent_style = tab\n",
		"sub/.editorconfig": "[*.php]\nindent_size = 8\n",
		FileName:            `{"format": {"settings": {"line_length": 110, "yoda": "keep"}, "risky": ["yoda", "increment_style", "single_quote"], "allowRisky": false}}`,
	})
	r, err := c.FormatResolver(FormatFlags{Set: map[string]any{"concat_space": "none", "single_quote": false}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := r.Settings(filepath.Join(dir, "a.php"))
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string][2]any{
		"indent_style":             {"space", ".editorconfig"},
		"indent_size":              {2, ".editorconfig"},
		"tab_width":                {2, ".editorconfig"},
		"line_ending":              {"crlf", ".editorconfig"},
		"final_newline":            {false, ".editorconfig"},
		"trim_trailing_whitespace": {false, ".editorconfig"},
		"line_length":              {110, "custos.json"},
		"yoda":                     {"keep", "custos.json"},
		"increment_style":          {"pre", "custos.json"},
		"concat_space":             {"none", "command line"},
		"single_quote":             {false, "command line"},
		"closing_tag":              {"remove", "preset psr12"},
	} {
		d, _ := format.Lookup(id)
		if got := d.Get(&res.Settings); got != want[0] || res.Source[id] != want[1] {
			t.Errorf("%s = %v from %s, want %v from %s", id, got, res.Source[id], want[0], want[1])
		}
	}
	// Files under the same sections share their settings; others do not.
	again, _ := r.Settings(filepath.Join(dir, "b.php"))
	sub, _ := r.Settings(filepath.Join(dir, "sub", "c.php"))
	mk, _ := r.Settings(filepath.Join(dir, "Makefile"))
	if again != res || sub == res || sub.IndentSize != 8 || mk.IndentStyle != "tab" {
		t.Fatalf("cache: %p %p %d %s", again, res, sub.IndentSize, mk.IndentStyle)
	}
	// A preset named on the command line or in custos.json skips the
	// imported style; risky entries take the preset's risky value.
	r, err = c.FormatResolver(FormatFlags{Preset: "symfony"})
	if err != nil {
		t.Fatal(err)
	}
	if name, source := r.Preset(); name != "symfony" || source != "command line" {
		t.Fatal(name, source)
	}
	c.Format.Preset, c.Format.EditorConfig, c.Format.Risky = "laravel", false, []string{"increment_style"}
	r, _ = c.FormatResolver(FormatFlags{Preset: "AUTO", AllowRisky: true})
	res, _ = r.Settings(filepath.Join(dir, "a.php"))
	if name, source := r.Preset(); name != "laravel" || source != "custos.json" || res.IndentSize != 4 || res.IncrementStyle != "post" || !res.SingleQuote || res.Source["indent_style"] == "mago.toml" {
		t.Fatalf("%s %s %+v", name, source, res.Settings)
	}
	if _, err := c.FormatResolver(FormatFlags{Set: map[string]any{"nope": 1}}); err == nil || !strings.Contains(err.Error(), "command line") {
		t.Fatal(err)
	}
}

func TestFormatEditorConfigErrors(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{})
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A directory named .editorconfig cannot be read.
	if err := os.Mkdir(filepath.Join(dir, "sub", ".editorconfig"), 0o755); err != nil {
		t.Fatal(err)
	}
	r, _ := c.FormatResolver(FormatFlags{})
	if _, err := r.Settings(filepath.Join(dir, "sub", "a.php")); err == nil {
		t.Fatal("unreadable .editorconfig accepted")
	}
}

func TestFormatHandBuiltConfig(t *testing.T) {
	c := &Config{Root: t.TempDir()}
	c.Format.EditorConfig = false
	r, err := c.FormatResolver(FormatFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if name, _ := r.Preset(); name != format.DefaultPreset {
		t.Fatal(name)
	}
}

func TestFormatExcluded(t *testing.T) {
	c, dir := loadFormat(t, map[string]string{FileName: `{"format": {"exclude": ["legacy/**", "generated", "*.tpl.php"]}}`})
	for rel, want := range map[string]bool{
		"legacy/a.php":        true,
		"src/legacy/a.php":    false,
		"src/generated/a.php": true,
		"src/view.tpl.php":    true,
		"src/a.php":           false,
		"../outside.php":      false,
	} {
		if got := c.FormatExcluded(filepath.Join(dir, rel)); got != want {
			t.Errorf("%s: %v", rel, got)
		}
	}
	plain, _ := loadFormat(t, map[string]string{})
	if plain.FormatExcluded("a.php") {
		t.Fatal("excluded without globs")
	}
}

func TestOverlayKeepsFormat(t *testing.T) {
	c, _ := loadFormat(t, map[string]string{FileName: `{"format": {"preset": "drupal", "editorconfig": false}}`})
	merged, err := Overlay(c, File{})
	if err != nil || merged.Format.Preset != "drupal" || merged.Format.EditorConfig {
		t.Fatal(err, merged.Format.Preset)
	}
}

func TestConfinedRead(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "ws")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{dir, inner} {
		if err := os.WriteFile(filepath.Join(d, ".editorconfig"), []byte("[*]\nindent_size = 2\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := confinedRead(inner)
	if _, err := read(filepath.Join(dir, ".editorconfig")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("outside: %v", err)
	}
	if _, err := read(filepath.Join(inner, ".editorconfig")); err != nil {
		t.Errorf("inside: %v", err)
	}
	if _, err := confinedRead("")(filepath.Join(dir, ".editorconfig")); err != nil {
		t.Errorf("unconfined: %v", err)
	}
}

func TestFormatUnknownPresetFlag(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.FormatResolver(FormatFlags{Preset: "nope"}); err == nil || !strings.Contains(err.Error(), `unknown preset "nope" (want auto,`) {
		t.Errorf("unknown preset: %v", err)
	}
}

// TestConfinedReadSymlink: a link inside the workspace to a file outside it
// reads as absent.
func TestConfinedReadSymlink(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "ws")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(dir, "secret")
	if err := os.WriteFile(outside, []byte("[*]\nindent_size = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(inner, ".editorconfig")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symbolic links unavailable:", err)
	}
	if _, err := confinedRead(inner)(link); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("link outside: %v", err)
	}
	if _, err := confinedRead(inner)(filepath.Join(inner, "missing", ".editorconfig")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing: %v", err)
	}
}

func TestEditorValuesBounds(t *testing.T) {
	big, ok := 5000, 2
	v := editorValues(editorconfig.Settings{MaxLineLength: &big, IndentSize: &ok})
	if _, set := v["line_length"]; set || v["indent_size"] != 2 {
		t.Errorf("values: %v", v)
	}
}
