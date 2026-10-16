package project

import (
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/unescapedhtmloutput"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/infer"
)

func TestProjectAssertionOptions(t *testing.T) {
	root := t.TempDir()
	wrapper := filepath.Join(root, "wrapper.php")
	main := filepath.Join(root, "main.php")
	writeFile(t, wrapper, `<?php function escaped($v){assert(($v=htmlspecialchars($v))!=='');return $v;} function checked(?string $v){assert(is_string($v));return $v;}`)
	writeFile(t, main, `<?php echo '<p>'.escaped($_GET['x']);`)
	for _, enabled := range []bool{false, true} {
		opt := syntax.Options{Version: phpversion.PHP84}
		ix := BuildIndex([]string{wrapper, main}, opt, infer.Options{Assertions: enabled})
		want := "null|string"
		if enabled {
			want = "string"
		}
		if got := ix.Function("checked", phpversion.PHP84).Inferred; got != want {
			t.Fatalf("index enabled=%v return=%s want=%s", enabled, got, want)
		}
		p, err := Open([]analysis.Rule{unescapedhtmloutput.New()}, Options{Root: root, Paths: []string{root}, Parse: opt, Analysis: analysis.Config{PHP: phpversion.PHP84, Flow: flow.Options{Assertions: enabled}}})
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, result := range p.AnalyzeReport() {
			count += len(result.Findings)
		}
		expected := 1
		if enabled {
			expected = 0
		}
		if count != expected {
			t.Fatalf("project enabled=%v findings=%d", enabled, count)
		}
		for _, prepared := range p.PrepareFixes() {
			if prepared.Err != nil {
				t.Fatal(prepared.Err)
			}
		}
	}
}
