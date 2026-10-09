package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"reflect"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
)

func TestFormats(t *testing.T) {
	items := []diagnostic.Item{{Path: "a.php", Line: 2, Column: 3, EndLine: 2, EndColumn: 5, Rule: "UnnecessarySemicolon", Severity: "info", Message: "m", Fixable: true}}
	for _, f := range Formats {
		var b bytes.Buffer
		if err := Write(&b, f, items, 1); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if !strings.Contains(b.String(), "a.php") {
			t.Errorf("%s output lacks path:\n%s", f, b.String())
		}
		if f == "json" || f == "sarif" {
			var v any
			if err := json.Unmarshal(b.Bytes(), &v); err != nil {
				t.Errorf("%s: invalid JSON: %v", f, err)
			}
		}
	}
	if err := Write(&bytes.Buffer{}, "nope", nil, 0); err == nil {
		t.Error("unknown format must fail")
	}
}

func TestItems(t *testing.T) {
	src := []byte("<?php\n$é = 1;;\n")
	results := []diagnostic.FileResult{
		{Path: "io.php", Err: errors.New("boom")},
		{Path: "clean.php", Src: src},
		{
			Path:   "a.php",
			Src:    src,
			Errors: []syntax.Error{{Span: syntax.Span{Start: 6, End: 9}, Msg: "bad"}},
			Findings: []diagnostic.Finding{
				{Rule: "R", Severity: diagnostic.SeverityInfo, Message: "m", Span: syntax.Span{Start: 14, End: 15}, Fixes: []diagnostic.Fix{{}}},
				{Rule: "S", Severity: diagnostic.SeverityWarning, Message: "n", Span: syntax.Span{Start: 6, End: 7}},
			},
		},
	}
	got := diagnostic.Items(results)
	want := []diagnostic.Item{
		{Path: "io.php", Line: 1, Column: 1, EndLine: 1, EndColumn: 1, Rule: "io", Severity: "error", Message: "boom"},
		{Path: "a.php", Line: 2, Column: 1, EndLine: 2, EndColumn: 3, Rule: "syntax", Severity: "error", Message: "bad"},
		{Path: "a.php", Line: 2, Column: 8, EndLine: 2, EndColumn: 9, Rule: "R", Severity: "info", Message: "m", Fixable: true},
		{Path: "a.php", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Rule: "S", Severity: "warning", Message: "n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("items:\n got %+v\nwant %+v", got, want)
	}
}

func TestFormatDetails(t *testing.T) {
	items := []diagnostic.Item{
		{Path: "./a.php", Line: 1, Column: 2, EndLine: 1, EndColumn: 3, Rule: "UnnecessarySemicolon", Severity: "error", Message: "50%\nx"},
		{Path: `b\c.php`, Line: 4, Column: 1, EndLine: 4, EndColumn: 2, Rule: "Custom", Severity: "odd", Message: "m"},
		{Path: `b\c.php`, Line: 5, Column: 1, EndLine: 5, EndColumn: 2, Rule: "Custom", Severity: "warning", Message: "w"},
	}
	out := func(format string) string {
		var b bytes.Buffer
		if err := Write(&b, format, items, 2); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		return b.String()
	}
	if s := out("github"); !strings.Contains(s, "::error file=./a.php,line=1,col=2,endLine=1,endColumn=3,title=UnnecessarySemicolon::50%25%0Ax\n") ||
		!strings.Contains(s, "::notice file=b\\c.php") || !strings.Contains(s, "::warning file=b\\c.php,line=5") {
		t.Errorf("github:\n%s", s)
	}
	if s := out("checkstyle"); strings.Count(s, "<file ") != 2 || !strings.Contains(s, `severity="info"`) || !strings.Contains(s, `severity="error"`) || !strings.Contains(s, `source="custos.Custom"`) {
		t.Errorf("checkstyle:\n%s", s)
	}
	if s := out("text"); !strings.HasSuffix(s, "2 file(s) analysed: 1 error(s), 1 warning(s), 0 info\n") {
		t.Errorf("text:\n%s", s)
	}
	var sarif struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Rules []map[string]any `json:"rules"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				RuleIndex int    `json:"ruleIndex"`
				Level     string `json:"level"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal([]byte(out("sarif")), &sarif); err != nil {
		t.Fatal(err)
	}
	run := sarif.Runs[0]
	if len(run.Tool.Driver.Rules) != 2 || run.Tool.Driver.Rules[0]["shortDescription"] == nil || run.Tool.Driver.Rules[1]["properties"] != nil {
		t.Errorf("sarif rules: %+v", run.Tool.Driver.Rules)
	}
	if r := run.Results; len(r) != 3 || r[0].Level != "error" || r[1].Level != "warning" || r[1].RuleIndex != 1 ||
		r[0].Locations[0].PhysicalLocation.ArtifactLocation.URI != "a.php" || r[1].Locations[0].PhysicalLocation.ArtifactLocation.URI != "b/c.php" {
		t.Errorf("sarif results: %+v", r)
	}
	if s := func() string {
		var b bytes.Buffer
		_ = Write(&b, "json", nil, 0)
		return b.String()
	}(); !strings.Contains(s, `"findings": []`) {
		t.Errorf("json without findings: %s", s)
	}
}

// failWriter accepts n bytes then fails, like a closed pipe.
type failWriter struct{ n int }

func (w *failWriter) Write(p []byte) (int, error) {
	if len(p) > w.n {
		k := w.n
		w.n = 0
		return k, errors.New("broken pipe")
	}
	w.n -= len(p)
	return len(p), nil
}

// TestWriteErrors: every format reports a failing output instead of
// claiming success.
func TestWriteErrors(t *testing.T) {
	items := []diagnostic.Item{{Path: "a.php", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Rule: "R", Severity: "info", Message: "m"}}
	for _, f := range Formats {
		for _, n := range []int{0, len(xml.Header)} {
			if err := Write(&failWriter{n: n}, f, items, 1); err == nil {
				t.Errorf("%s (fail after %d bytes): error not reported", f, n)
			}
		}
	}
}
