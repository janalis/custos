package lsp

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/runner"
	"custos/internal/syntax"
)

func TestDeclarationShadowsThroughWorkspaceChanges(t *testing.T) {
	dir := t.TempDir()
	s := whiteBox(t, dir, io.Discard)
	calls := filepath.Join(dir, "calls.php")
	shadow := filepath.Join(dir, "shadow.php")
	s.index = runner.BuildIndex(nil, syntax.Options{Version: s.cfg.PHP})
	s.index.Add(runner.ExtractSymbols(calls, []byte(`<?php namespace App; class Original {} define('FLAG', 7); class_alias(Original::class, 'Alias');`), syntax.Options{Version: s.cfg.PHP}))
	check := func(want bool) {
		t.Helper()
		if (s.index.Constant("FLAG", s.cfg.PHP) != nil) != want || (s.index.Class("Alias", s.cfg.PHP) != nil) != want {
			t.Fatalf("constant and alias availability must both be %v", want)
		}
	}
	check(true)
	if err := os.WriteFile(shadow, []byte(`<?php namespace App; function define() {} function class_alias() {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s.watchedFilesChanged([]fileChange{{URI: "file://" + shadow, Type: 1}})
	check(false)
	if err := os.WriteFile(shadow, []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.watchedFilesChanged([]fileChange{{URI: "file://" + shadow, Type: 2}})
	check(true)
	uri := "file://" + shadow
	s.docs[uri] = &document{path: shadow, text: []byte(`<?php namespace App; function define() {} function class_alias() {}`)}
	if !s.reindexDoc(uri) {
		t.Fatal("buffer was not indexed")
	}
	check(false)
	s.docs[uri].text = []byte("<?php")
	s.reindexDoc(uri)
	check(true)
	s.docs[uri].text = []byte(`<?php namespace App; function define() {} function class_alias() {}`)
	s.reindexDoc(uri)
	check(false)
	delete(s.docs, uri)
	s.watchedFilesChanged([]fileChange{{URI: uri, Type: 3}})
	check(true)
}
