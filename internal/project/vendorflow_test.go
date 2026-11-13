package project

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A dependency wrapper reaching a sink is summarized on demand: the project
// call passing request data is reported, as with a project wrapper.
func TestVendorWrapperSinks(t *testing.T) {
	for _, tc := range []struct {
		name, composer, dir string
		want                int
	}{
		{"vendor", ``, "vendor", 1},
		{"configured", `{"config":{"vendor-dir":"lib/deps"}}`, "lib/deps", 1},
		{"project path", ``, "other", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.composer != "" {
				writeFile(t, filepath.Join(root, "composer.json"), tc.composer)
			}
			writeFile(t, filepath.Join(root, tc.dir, "acme", "run", "Runner.php"), `<?php namespace Acme; final class Runner { public function run($command) { system($command); } }`)
			main := filepath.Join(root, "src", "main.php")
			writeFile(t, main, `<?php function go(\Acme\Runner $r) { $r->run($_GET['c']); }`)
			opt := syntax.Options{Version: phpversion.PHP84}
			paths := []string{filepath.Join(root, "src")}
			if tc.dir == "other" {
				paths = append(paths, filepath.Join(root, "other"))
			}
			p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: paths, Scope: []string{filepath.Join(root, "src")}, Parse: opt, Analysis: analysis.Config{PHP: phpversion.PHP84}})
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, result := range p.AnalyzeReport() {
				if result.Path == main {
					count += len(result.Findings)
				}
			}
			if count != tc.want {
				t.Fatalf("findings=%d want %d", count, tc.want)
			}
		})
	}
}

// Relative paths, as the CLI passes them, resolve against the working
// directory read once, the way filepath.Abs resolves them on each call.
func TestVendorWrapperSinksRelative(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "vendor", "acme", "run", "Runner.php"), `<?php namespace Acme; final class Runner { public function run($command) { system($command); } }`)
	writeFile(t, filepath.Join(root, "src", "main.php"), `<?php function go(\Acme\Runner $r) { $r->run($_GET['c']); }`)
	t.Chdir(root)
	opt := syntax.Options{Version: phpversion.PHP84}
	p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: ".", Paths: []string{"src"}, Scope: []string{"src"}, Parse: opt, Analysis: analysis.Config{PHP: phpversion.PHP84}})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, result := range p.AnalyzeReport() {
		count += len(result.Findings)
	}
	if count != 1 {
		t.Fatalf("findings=%d want 1", count)
	}
}

func TestAbsoluteFrom(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"a/b.php", "./a/../b.php", "..", filepath.Join(cwd, "x", "..", "y.php")} {
		if got, want := absoluteFrom(cwd, p), absolute(p); got != want {
			t.Errorf("absoluteFrom(%q)=%q want %q", p, got, want)
		}
	}
}
