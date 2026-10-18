package diagnostic

import (
	"errors"
	"testing"

	"custos/internal/php/syntax"
)

func TestAnalysisStatusPositionsIssues(t *testing.T) {
	status := Status([]FileResult{
		{Path: "clean.php"},
		{Path: "failed.php", Err: errors.New("read failed")},
		{Path: "unicode.php", Src: []byte("<?php\néx"), Incomplete: []AnalysisIssue{{Reason: "dynamic-code", Span: syntax.Span{Start: 8, End: 9}}}},
	})
	if status.Complete || len(status.Issues) != 2 {
		t.Fatal(status)
	}
	issue := status.Issues[1]
	if issue.Path != "unicode.php" || issue.Line != 2 || issue.Column != 2 || issue.EndLine != 2 || issue.EndColumn != 3 {
		t.Fatal(issue)
	}
	if clean := Status(nil); !clean.Complete || clean.Issues == nil {
		t.Fatal(clean)
	}
}
