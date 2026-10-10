package gdjpegqualityoutsidesupportedrange

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/testing/conformance"
)

func TestContracts(t *testing.T) {
	r := New()
	if semantic, ok := r.(analysis.SemanticRule); ok {
		semantic.Semantic()
	}
	if flow, ok := r.(analysis.FlowRule); ok {
		flow.Flow()
	}
	fixtures, err := filepath.Glob("../../../../testdata/rules/GdJpegQualityOutsideSupportedRange/*.php")
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	for _, path := range fixtures {
		if strings.HasSuffix(path, ".fixed.php") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			src, want, err := conformance.ParseMarkup(raw)
			if err != nil {
				t.Fatal(err)
			}
			cfg := analysis.Config{Only: []string{"GdJpegQualityOutsideSupportedRange"}}
			var side struct {
				PHP     string         `json:"php"`
				Options map[string]any `json:"options"`
			}
			settings, err := os.ReadFile(strings.TrimSuffix(path, ".php") + ".json")
			if err == nil {
				if err = json.Unmarshal(settings, &side); err != nil {
					t.Fatal(err)
				}
				if side.PHP != "" {
					cfg.PHP = phpversion.MustParse(side.PHP)
				}
				cfg.Rules = map[string]analysis.RuleConfig{"GdJpegQualityOutsideSupportedRange": {Options: side.Options}}
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			e, err := analysis.NewEngine([]analysis.Rule{r}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse(path, src, syntax.Options{})
			if !strings.HasPrefix(filepath.Base(path), "broken") && len(f.Errors) > 0 {
				t.Fatalf("invalid fixture: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != len(want) {
				t.Fatalf("got %d want %d: %+v", len(got), len(want), got)
			}
			var edits []diagnostic.TextEdit
			for i, g := range got {
				w := want[i]
				if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity {
					t.Fatalf("got %+v want %+v", g, w)
				}
				if len(g.Fixes) > 0 {
					edits = append(edits, g.Fixes[0].Edits()...)
				}
			}
			expected, err := os.ReadFile(strings.TrimSuffix(path, ".php") + ".fixed.php")
			if os.IsNotExist(err) {
				if len(edits) > 0 {
					t.Fatal("fix without expected output")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			fixed, _ := fixing.Apply(src, edits)
			if !bytes.Equal(fixed, expected) {
				t.Fatalf("fixed:\n%s\nwant:\n%s", fixed, expected)
			}
			parsed := syntax.Parse(path, fixed, syntax.Options{})
			if len(parsed.Errors) > 0 {
				t.Fatalf("invalid fix: %+v", parsed.Errors)
			}
			if again := e.Analyze(parsed); len(again) > 0 {
				t.Fatalf("fix not stable: %+v", again)
			}
		})
	}
}
