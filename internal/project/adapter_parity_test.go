package project

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/catalogue"
	"github.com/janalis/custos/internal/project/config"
)

// The command line (Open, AnalyzeReport) and the MCP server (a snapshot
// of the root) build the index and the flow summaries their own way; a
// project must report the same findings through both. The fixture has
// cross-file flow, a vendor library, a declaration-only file and an
// excluded directory.
func TestAdapterParity(t *testing.T) {
	root := t.TempDir()
	for name, code := range map[string]string{
		"custos.json":                    `{"exclude": ["generated"], "rules": {"UntrustedShellCommand": {"enabled": true}}}`,
		"composer.json":                  `{"autoload": {"psr-4": {"App\\": "src/"}}}`,
		"src/Shell.php":                  "<?php namespace App; function shell($c) { \\Lib\\run($c); }",
		"src/Entry.php":                  "<?php namespace App; shell($_GET['x']); $a = !!!$b; $d = new \\Ext\\Thing(); $d->missing();",
		"src/ext.stub.php":               "<?php namespace Ext; final class Thing { public function known(): void {} }",
		"src/generated/G.php":            "<?php $z = !!!$y;",
		"vendor/lib/src/run.php":         "<?php namespace Lib; function run($c) { system($c); }",
		"vendor/composer/installed.json": `{"packages": []}`,
	} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(name)), code)
	}
	render := func(results []diagnostic.FileResult, rel func(string) string) map[string]string {
		out := map[string]string{}
		for _, r := range results {
			out[rel(r.Path)] = renderResults([]diagnostic.FileResult{{Findings: r.Findings, Incomplete: r.Incomplete, Err: r.Err}})
		}
		return out
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	acfg := cfg.Analysis()
	acfg.ProjectPath = DiskProjectPaths(cfg.Root)
	var paths []string
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	if len(paths) == 0 {
		paths = []string{cfg.Root}
	}
	p, err := Open(catalogue.All(), Options{Root: cfg.Root, Paths: paths, Exclude: cfg.Exclude, Analysis: acfg, Parse: cfg.Parse()})
	if err != nil {
		t.Fatal(err)
	}
	cli := render(p.AnalyzeReport(), func(path string) string {
		rel, _ := filepath.Rel(cfg.Root, path)
		return filepath.ToSlash(rel)
	})
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	s, err := BuildSnapshot(context.Background(), r, catalogue.All(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mcp := render(s.Results, func(path string) string { return path })
	if len(cli) == 0 || len(cli) != len(mcp) {
		t.Fatalf("files: command line %v, snapshot %v", cli, mcp)
	}
	for path, want := range cli {
		if mcp[path] != want {
			t.Fatalf("%s: command line\n%s\nsnapshot\n%s", path, want, mcp[path])
		}
	}
}
