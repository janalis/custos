//go:build unix

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	withPath(t, fakeBin(t, "composer"), false)
	work := t.TempDir()
	b := &bench{cfg: &Config{Packages: map[string]string{"v/tool": "1.0.0"}}, work: work}
	tools := []Tool{{ID: "a", Package: "v/tool"}, {ID: "b", Package: "v/tool"}, {ID: "custos"}}
	corpora := []Corpus{
		{Name: "app", Repo: gitRepo(t, true, "1.0.0"), Tag: "1.0.0"},
		{Name: "lib", Repo: gitRepo(t, false, "2.0.0"), Tag: "2.0.0"},
	}
	var log bytes.Buffer
	if err := b.install(tools, corpora, &log); err != nil {
		t.Fatalf("install: %v\n%s", err, log.String())
	}
	for _, p := range []string{"tools/v-tool/vendor", "corpora/app/vendor", "corpora/lib/src/a.php"} {
		if _, err := os.Stat(filepath.Join(work, p)); err != nil {
			t.Error(err)
		}
	}
	if _, err := os.Stat(filepath.Join(work, "corpora/lib/vendor")); err == nil {
		t.Error("composer ran for a corpus without composer.json")
	}
	if n := strings.Count(log.String(), "installing"); n != 3 {
		t.Errorf("%d installs, want 3:\n%s", n, log.String())
	}

	// Up to date: nothing is reinstalled.
	log.Reset()
	if err := b.install(tools, corpora, &log); err != nil || strings.Contains(log.String(), "installing") {
		t.Errorf("second install: %v\n%s", err, log.String())
	}

	// A new version reinstalls, and failures name the package or corpus.
	b.cfg.Packages["v/tool"] = "1.1.0"
	t.Setenv("FAKE_COMPOSER_FAIL", "require")
	if err := b.install(tools, nil, &log); err == nil || !strings.Contains(err.Error(), "v/tool") {
		t.Errorf("failing composer require: %v", err)
	}
	t.Setenv("FAKE_COMPOSER_FAIL", "install")
	corpora[0].Tag = "1.0.0 "
	if err := b.install(nil, corpora[:1], &log); err == nil || !strings.Contains(err.Error(), "app") {
		t.Errorf("failing clone: %v", err)
	}
	corpora[0].Tag = "1.0.0"
	os.Remove(filepath.Join(work, "corpora", "app.version"))
	if err := b.install(nil, corpora[:1], &log); err == nil || !strings.Contains(err.Error(), "app") {
		t.Errorf("failing composer install: %v", err)
	}

	// The directory cannot be created.
	file := filepath.Join(t.TempDir(), "file")
	write(t, file, "")
	b.work = file
	if err := b.install(nil, corpora[:1], &log); err == nil {
		t.Error("work is a file: no error")
	}
}
