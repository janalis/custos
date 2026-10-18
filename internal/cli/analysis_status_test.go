package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
)

func TestAnalysisIncompleteDoesNotChangeFindingExitPolicy(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dynamic.php")
	source := `<?php function example($path,$code){$body=file_get_contents($path);eval($code);strlen($body);}`
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := runCLI(t, "analyse", "--rule", "FileReadFailureUnchecked", "--format=json", "--fail-on=never", file)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, stderr)
	}
	var report struct {
		Analysis diagnostic.AnalysisStatus `json:"analysis"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatal(err)
	}
	if report.Analysis.Complete || len(report.Analysis.Issues) == 0 || !strings.Contains(stderr, "dynamic-code") {
		t.Fatalf("status=%+v stderr=%s", report.Analysis, stderr)
	}
}
