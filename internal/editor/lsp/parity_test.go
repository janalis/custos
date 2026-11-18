package lsp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/project/config"
)

// The language server indexes and summarizes the workspace its own way; an
// open document must show the findings the command line reports for it,
// cross-file flow through a dependency included.
func TestFindingsMatchCommandLine(t *testing.T) {
	root := t.TempDir()
	for name, code := range map[string]string{
		"custos.json":            `{"rules": {"UntrustedShellCommand": {"enabled": true}}}`,
		"src/Shell.php":          "<?php namespace App; function shell($c) { \\Lib\\run($c); }",
		"src/Entry.php":          "<?php namespace App; shell($_GET['x']); $a = !!!$b; $d = new \\Ext\\Thing(); $d->known();",
		"src/ext.stub.php":       "<?php namespace Ext; final class Thing { public function known(): void {} }",
		"vendor/lib/src/run.php": "<?php namespace Lib; function run($c) { system($c); }",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	render := func(fs []diagnostic.Finding) string {
		var b strings.Builder
		for _, f := range fs {
			fmt.Fprintf(&b, "%s %v %q\n", f.Rule, f.Span, f.Message)
		}
		return b.String()
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	acfg := cfg.Analysis()
	acfg.ProjectPath = project.DiskProjectPaths(cfg.Root)
	entry := filepath.Join(root, "src", "Entry.php")
	p, err := project.Open(catalogue.All(), project.Options{Root: cfg.Root, Paths: []string{root}, Exclude: cfg.Exclude, Analysis: acfg, Parse: cfg.Parse()})
	if err != nil {
		t.Fatal(err)
	}
	want := ""
	for _, r := range p.AnalyzeReport() {
		if r.Path == entry {
			want = render(r.Findings)
		}
	}
	if !strings.Contains(want, "UntrustedShellCommand") {
		t.Fatalf("command line: %s", want)
	}
	s := whiteBox(t, root, io.Discard, catalogue.All()...)
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	src, _ := os.ReadFile(entry)
	uri := "file://" + entry
	s.docs[uri] = &document{uri: uri, path: entry, text: src, lines: syntax.NewLineIndex(src), analyzed: -1}
	s.analyzeNow(uri)
	if got := render(s.docs[uri].findings); got != want {
		t.Fatalf("language server\n%s\ncommand line\n%s", got, want)
	}
}
