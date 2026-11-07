//go:build unix

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `{"runs": 1, "warmup": 0, "timeout": "1m",
  "corpora": [{"name": "c", "repo": "r", "tag": "1.0.0", "paths": ["src"]}],
  "packages": {"vendor/tool": "1.0.0"},
  "tools": [{"id": "t", "label": "T", "kind": "linter", "package": "vendor/tool", "cmd": ["x"], "findings": "phpcs"}]}`

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.json")
	write(t, path, validConfig)
	c, err := loadConfig(path)
	if err != nil || c.Tools[0].ID != "t" {
		t.Fatalf("loadConfig = %v, %v", c, err)
	}
	if _, err := loadConfig(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("missing file: no error")
	}
	for name, content := range map[string]string{
		"not JSON":       "{",
		"no runs":        strings.Replace(validConfig, `"runs": 1`, `"runs": 0`, 1),
		"bad timeout":    strings.Replace(validConfig, `"1m"`, `"soon"`, 1),
		"unknown parser": strings.Replace(validConfig, `"phpcs"`, `"nope"`, 1),
		"unpinned":       strings.Replace(validConfig, `"package": "vendor/tool"`, `"package": "other/tool"`, 1),
	} {
		write(t, path, content)
		if _, err := loadConfig(path); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
