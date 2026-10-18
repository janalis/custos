package lsp

import (
	"bytes"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/rules/filereadfailureunchecked"
)

func TestAnalysisStatusReasonsAreStable(t *testing.T) {
	issues := []diagnostic.AnalysisIssue{{Reason: "flow-budget"}, {Reason: "dynamic-code"}, {Reason: "flow-budget"}}
	if got := analysisStatus(issues); got != "dynamic-code, flow-budget" {
		t.Fatal(got)
	}
	if got := analysisStatus(nil); got != "" {
		t.Fatal(got)
	}
}

func TestDocumentStatusChangesArePublished(t *testing.T) {
	useRules(t, filereadfailureunchecked.New())
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/custos-incomplete.php"
	cl.open(uri, `<?php function example($path,$code){$body=file_get_contents($path);eval($code);strlen($body);}`)
	cl.logContaining("analysis incomplete")
	cl.diagsFor(uri)
	cl.change(uri, 2, `<?php function example(){return 1;}`)
	cl.logContaining("analysis complete")
	cl.diagsFor(uri)
}

func TestAnalysisPanicCannotPublishCompleteStatus(t *testing.T) {
	var log bytes.Buffer
	s := whiteBox(t, t.TempDir(), &log, testRule{})
	result := s.guardedAnalysis("broken.php", func() diagnostic.FileResult { panic("unexpected engine failure") })
	if got := analysisStatus(result.Incomplete); got != "internal-failure" || !strings.Contains(log.String(), "unexpected engine failure") {
		t.Fatalf("status=%q log=%s", got, log.String())
	}
	clean := s.guardedAnalysis("clean.php", func() diagnostic.FileResult { return diagnostic.FileResult{Path: "clean.php"} })
	if clean.Path != "clean.php" || len(clean.Incomplete) != 0 {
		t.Fatalf("clean=%+v", clean)
	}
}
