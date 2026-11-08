package project

import (
	"fmt"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A chain of wrappers across more files than compose checks one by one
// settles as a short one does: request data still reaches the sink through
// every link.
func TestFlowComposesLongCrossFileChains(t *testing.T) {
	for _, links := range []int{2, minParallelChecks + 8} {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "g0.php"), "<?php function g0() { return $_GET['value']; }")
		for i := 1; i < links; i++ {
			writeFile(t, filepath.Join(root, fmt.Sprintf("g%d.php", i)), fmt.Sprintf("<?php function g%d() { return g%d(); }", i, i-1))
		}
		writeFile(t, filepath.Join(root, "main.php"), fmt.Sprintf("<?php system(g%d());", links-1))
		p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: syntax.Options{Version: phpversion.Default}})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, result := range p.AnalyzeReport() {
			count += len(result.Findings)
		}
		if count != 1 {
			t.Fatalf("%d links: %d findings, want 1", links, count)
		}
	}
}
