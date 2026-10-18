package report

import (
	"bytes"
	"encoding/json"
	"testing"

	"custos/internal/diagnostic"
)

func TestMachineReportsAnalysisStatus(t *testing.T) {
	status := diagnostic.AnalysisStatus{Issues: []diagnostic.AnalysisLocation{{Path: "source.php", Reason: "dynamic-code", Line: 2, Column: 3, EndLine: 2, EndColumn: 9}}}
	for _, format := range []string{"json", "sarif"} {
		t.Run(format, func(t *testing.T) {
			var out bytes.Buffer
			if err := WriteWithAnalysis(&out, format, nil, 1, status); err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(out.Bytes(), &document); err != nil {
				t.Fatal(err)
			}
			metadata := document["analysis"]
			if format == "sarif" {
				run := document["runs"].([]any)[0].(map[string]any)
				metadata = run["properties"].(map[string]any)["analysis"]
			}
			analysis := metadata.(map[string]any)
			if analysis["complete"] != false || len(analysis["incomplete"].([]any)) != 1 {
				t.Fatal(analysis)
			}
		})
	}
	var out bytes.Buffer
	if err := WriteWithAnalysis(&out, "text", nil, 0, diagnostic.Status(nil)); err != nil {
		t.Fatal(err)
	}
	if err := WriteWithAnalysis(&out, "unknown", nil, 0, status); err == nil {
		t.Fatal("invalid format accepted")
	}
}

func TestAnalysisMetadataPreservesFindingsAndWriterErrors(t *testing.T) {
	items := []diagnostic.Item{{Rule: "Example", Path: "example.php", Severity: "warning", Message: "example", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}}
	for _, format := range []string{"json", "sarif"} {
		var output bytes.Buffer
		if err := WriteWithAnalysis(&output, format, items, 1, diagnostic.Status(nil)); err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(output.Bytes(), []byte("Example")) || !bytes.Contains(output.Bytes(), []byte(`"complete": true`)) {
			t.Fatalf("lost findings or status: %s", output.Bytes())
		}
		if err := WriteWithAnalysis(&failWriter{}, format, items, 1, diagnostic.Status(nil)); err == nil {
			t.Fatal("writer error lost")
		}
	}
}
