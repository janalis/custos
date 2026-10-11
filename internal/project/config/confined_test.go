package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfinedSettings(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	cfg, err := LoadConfined(root)
	if err != nil || cfg.Root != root.Name() {
		t.Fatal(cfg, err)
	}
	sources := map[string][]byte{FileName: []byte(`{"php":"8.2"}`), "composer.json": []byte(`{"require-dev":{"phpunit/phpunit":"^9.2"}}`)}
	for name, b := range sources {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err = LoadConfined(root)
	if err != nil || cfg.PHP.String() != "8.2" {
		t.Fatal(cfg, err)
	}
	cfg, err = ResolveCaptured(dir, sources)
	if err != nil || cfg.PHP.String() != "8.2" {
		t.Fatal(cfg, err)
	}
	sources[FileName] = []byte("bad")
	if _, err := ResolveCaptured(dir, sources); err == nil {
		t.Fatal("invalid captured config")
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfined(root); err == nil {
		t.Fatal("invalid config")
	}
	if err := os.Remove(filepath.Join(dir, FileName)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, FileName), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfined(root); err == nil {
		t.Fatal("config directory accepted")
	}
	if err := os.Remove(filepath.Join(dir, FileName)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "composer.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(dir, "composer.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfined(root); err == nil {
		t.Fatal("unconfined composer accepted")
	}
}
