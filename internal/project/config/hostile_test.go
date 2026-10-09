package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// TestHostileConfig checks that malicious custos.json / composer.json files
// (they come with the analysed repository) give errors or defaults, never
// crashes, hangs or unbounded reads.
func TestHostileConfig(t *testing.T) {
	deep := strings.Repeat("[", 200000) + strings.Repeat("]", 200000)
	for name, tc := range map[string]struct {
		custos, composer string
		wantErr          bool
	}{
		"deep custos":          {custos: `{"paths": ` + deep + `}`, wantErr: true},
		"deep composer":        {composer: `{"require": ` + deep + `}`},
		"deep options":         {custos: `{"rules": {"NestedNotOperators": {"options": {"x": ` + deep + `}}}}`, wantErr: true},
		"huge array":           {custos: `{"exclude": [` + strings.Repeat(`"a",`, 500000) + `"a"]}`},
		"wrong types":          {custos: `{"php": 8.1, "paths": "src", "rules": []}`, wantErr: true},
		"wrong rule shape":     {custos: `{"rules": {"NestedNotOperators": "off"}}`, wantErr: true},
		"bad severity":         {custos: `{"rules": {"NestedNotOperators": {"severity": "fatal"}}}`, wantErr: true},
		"unknown rule":         {custos: `{"rules": {"Nope": {}}}`, wantErr: true},
		"bad php":              {custos: `{"php": "` + strings.Repeat("9", 100000) + `"}`, wantErr: true},
		"composer wrong types": {composer: `{"require": {"php": 7}, "config": {"platform": []}}`},
		"composer garbage":     {composer: "\xff\xfe{"},
		"composer constraint":  {composer: `{"require": {"php": "` + strings.Repeat(">=8.1 || ", 100000) + `"}}`},
		"paths escape":         {custos: `{"paths": ["src", "../../.."]}`, wantErr: true},
		"absolute path":        {custos: `{"paths": ["/etc"]}`, wantErr: true},
		"baseline escape":      {custos: `{"baseline": "../../../../dev/zero"}`, wantErr: true},
		"inner dotdot ok":      {custos: `{"paths": ["src/../lib"], "baseline": "a/../b.json"}`},
		"null everything":      {custos: `{"php": null, "paths": null, "rules": null}`},
		"options any types":    {custos: `{"rules": {"MultipleReturnStatements": {"options": {"COMPLAIN_THRESHOLD": 1e308, "SCREAM_THRESHOLD": {"a": [1]}}}}}`},
	} {
		dir := t.TempDir()
		if tc.custos != "" {
			if err := os.WriteFile(filepath.Join(dir, FileName), []byte(tc.custos), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if tc.composer != "" {
			if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(tc.composer), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		start := time.Now()
		c, err := Load(dir)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, want error %v", name, err, tc.wantErr)
		}
		if err == nil {
			c.Analysis()
		}
		if d := time.Since(start); d > testbudget.Of(3*time.Second) {
			t.Errorf("%s: %v", name, d)
		}
	}
}

func TestConfigSpecialFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no /dev/zero")
	}
	dir := t.TempDir()
	if err := os.Symlink("/dev/zero", filepath.Join(dir, FileName)); err != nil {
		t.Skip(err)
	}
	done := make(chan error, 1)
	go func() { _, err := Load(dir); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("custos.json -> /dev/zero accepted")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reading custos.json -> /dev/zero did not return")
	}
	dir = t.TempDir()
	if err := os.Symlink("/dev/zero", filepath.Join(dir, "composer.json")); err != nil {
		t.Skip(err)
	}
	if _, err := Load(dir); err != nil {
		t.Fatalf("composer.json -> /dev/zero: %v (want defaults)", err)
	}
}
