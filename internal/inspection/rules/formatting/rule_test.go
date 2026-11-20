package formatting

import (
	"testing"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/fixing"
	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func analyze(t *testing.T, settings func(string) (any, bool), src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"Formatting"}, Format: settings})
	if err != nil {
		t.Fatal(err)
	}
	if !e.NeedsFormat() {
		t.Fatal("the rule must declare that it formats code")
	}
	return e.Analyze(syntax.Parse("a.php", []byte(src), syntax.Options{}))
}

func TestExcludedFile(t *testing.T) {
	got := analyze(t, func(path string) (any, bool) {
		if path != "a.php" {
			t.Errorf("settings asked for %q", path)
		}
		return nil, false
	}, "<?php\nfunction f($a,$b){}\n")
	if len(got) != 0 {
		t.Fatalf("an excluded file is reported: %+v", got)
	}
}

func TestLongLineReport(t *testing.T) {
	r, err := format.Resolve(format.DefaultPreset, false, format.Layer{Source: "test", Values: map[string]any{"report_long_lines": true, "line_length": 20}})
	if err != nil {
		t.Fatal(err)
	}
	src := "<?php\n\n$total = $first + $second + $third;\n"
	got := analyze(t, func(string) (any, bool) { return &r.Settings, true }, src)
	if len(got) != 1 || got[0].Message != "Split the line, which exceeds the line length" || len(got[0].Fixes) != 0 || got[0].Span.Start != 7 {
		t.Fatalf("want one long-line report without a fix, got %+v", got)
	}
}

func TestFormatFileFix(t *testing.T) {
	src := []byte("<?php\n\nfunction total($a,$b){\n  return $a+$b;\n}\n")
	want := format.Source("a.php", src, format.Options{})
	got := analyze(t, nil, string(src))
	if len(got) < 2 {
		t.Fatalf("want several findings, got %+v", got)
	}
	for _, f := range got {
		if len(f.Fixes) != 2 || f.Fixes[1].Title != "Format the file" || f.Fixes[0].Title != f.Message {
			t.Fatalf("unexpected fixes: %+v", f)
		}
		single := f.Fixes[0].Edits()
		if len(single) != 1 || single[0].Span != f.Span {
			t.Fatalf("the first fix must apply its change alone: %+v", single)
		}
	}
	edits := got[0].Fixes[1].Edits()
	edits[0].NewText = "changed by the caller"
	edits = got[len(got)-1].Fixes[1].Edits()
	out, applied := fixing.Apply(src, edits)
	if applied != len(edits) || string(out) != string(want.Output) {
		t.Fatalf("Format the file gave %q (%d of %d applied), want %q", out, applied, len(edits), want.Output)
	}
}
