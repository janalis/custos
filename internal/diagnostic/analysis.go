package diagnostic

import "custos/internal/php/syntax"

// AnalysisIssue describes a proof the enabled checks could not complete.
// Reasons are stable identifiers; the span identifies the affected expression
// or scope. Issues are separate from findings and never determine exit severity.
type AnalysisIssue struct {
	Reason string      `json:"reason"`
	Span   syntax.Span `json:"span"`
}

// AnalysisResult retains findings and analysis status without a syntax tree.
type AnalysisResult struct {
	Findings   []Finding
	Incomplete []AnalysisIssue
}

// AnalysisLocation positions an incomplete proof for machine reports.
type AnalysisLocation struct {
	Path      string `json:"path"`
	Reason    string `json:"reason"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"endLine"`
	EndColumn int    `json:"endColumn"`
}

// AnalysisStatus describes completion of enabled checks, independently of whether
// they found a problem. Complete means no instrumented query reported a gap;
// it does not mean the program or the runtime model is complete.
type AnalysisStatus struct {
	Complete bool               `json:"complete"`
	Issues   []AnalysisLocation `json:"incomplete"`
}

// Status positions analysis issues in file order.
func Status(results []FileResult) AnalysisStatus {
	status := AnalysisStatus{Complete: true, Issues: []AnalysisLocation{}}
	for _, result := range results {
		if result.Err != nil {
			status.Issues = append(status.Issues, AnalysisLocation{Path: result.Path, Reason: "io-error", Line: 1, Column: 1, EndLine: 1, EndColumn: 1})
			continue
		}
		if len(result.Incomplete) == 0 {
			continue
		}
		lines := syntax.NewLineIndex(result.Src)
		for _, issue := range result.Incomplete {
			line, column := lines.RuneColumn(issue.Span.Start)
			endLine, endColumn := lines.RuneColumn(issue.Span.End)
			status.Issues = append(status.Issues, AnalysisLocation{Path: result.Path, Reason: issue.Reason, Line: line + 1, Column: column + 1, EndLine: endLine + 1, EndColumn: endColumn + 1})
		}
	}
	status.Complete = len(status.Issues) == 0
	return status
}
