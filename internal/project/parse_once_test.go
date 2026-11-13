package project

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// Keeping the trees between phases changes no result: the same findings,
// including for a file that cannot be read and one whose declarations
// crash extraction (it is still analysed).
func TestParseOnceMatches(t *testing.T) {
	root, paths := oracleFixture(t)
	writeFile(t, filepath.Join(root, "src", "c.php"), `<?php function c($v) { exec($_GET['x']); }`)
	unreadable := filepath.Join(root, "src", "d.php")
	writeFile(t, unreadable, `<?php function d() {}`)
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o644) })
	old := extractTree
	t.Cleanup(func() { extractTree = old })
	extractTree = func(f *syntax.File, opt syntax.Options, options infer.Options) *index.FileSymbols {
		if strings.HasSuffix(f.Path, "c.php") {
			panic("simulated extraction crash")
		}
		return old(f, opt, options)
	}
	run := func(once bool) []string {
		p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: paths, ParseOnce: once, Parse: syntax.Options{Version: phpversion.PHP84}, Analysis: analysis.Config{PHP: phpversion.PHP84}})
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, r := range p.AnalyzeReport() {
			var rules []string
			for _, f := range r.Findings {
				rules = append(rules, f.Rule)
			}
			out = append(out, filepath.Base(r.Path)+":"+strings.Join(rules, ","))
			if r.Err != nil {
				out = append(out, filepath.Base(r.Path)+": unreadable")
			}
		}
		return out
	}
	without, with := run(false), run(true)
	if !reflect.DeepEqual(without, with) || len(without) < 4 {
		t.Fatalf("without %v, with %v", without, with)
	}
}
