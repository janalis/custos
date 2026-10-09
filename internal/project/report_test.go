package project

import (
	"path/filepath"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type fixRule struct{}

func (fixRule) ID() string               { return "NestedNotOperators" }
func (fixRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (fixRule) Check(ctx *analysis.Context, n syntax.Node) {
	ctx.ReportNode(n, "found", diagnostic.Fix{Title: "drop", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: n.Span()}}
	}})
}

// TestRunReportDropsEdits: report mode keeps findings and fix titles but
// releases the edit closures; RunSources keeps them for fixing.
func TestRunReportDropsEdits(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{fixRule{}}, analysis.Config{PHP: phpversion.Default, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "a.php")
	writeFile(t, p, "<?php\n$a = !$b;\n")
	opt := syntax.Options{Version: phpversion.Default}
	rep := RunReport(e, []string{p}, nil, opt)
	if f := rep[0].Findings; len(f) != 1 || len(f[0].Fixes) != 1 || f[0].Fixes[0].Title != "drop" || f[0].Fixes[0].Edits != nil {
		t.Fatalf("report findings: %+v", f)
	}
	full := RunSources(e, []string{p}, nil, opt)
	if f := full[0].Findings; len(f) != 1 || f[0].Fixes[0].Edits == nil {
		t.Fatalf("fix findings: %+v", f)
	}
}
