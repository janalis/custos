package project

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type unknownRule struct{}

func (unknownRule) ID() string                           { return "not-in-catalogue" }
func (unknownRule) Kinds() []syntax.NodeKind             { return nil }
func (unknownRule) Check(*analysis.Context, syntax.Node) {}

func TestOpenErrors(t *testing.T) {
	if _, err := Open([]analysis.Rule{unknownRule{}}, Options{}); err == nil {
		t.Fatal("accepted unknown inspection")
	}
	if _, err := Open(nil, Options{Paths: []string{filepath.Join(t.TempDir(), "missing")}}); err == nil {
		t.Fatal("accepted missing project")
	}
}

func TestAnalyzeBufferDoesNotReadDisk(t *testing.T) {
	engine, err := analysis.NewEngine(nil, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "missing.php")
	result := AnalyzeBuffer(engine, path, []byte("<?php $x = 1;"), syntax.Options{})
	if result.Err != nil || len(result.Errors) != 0 || result.Path != path {
		t.Fatalf("%+v", result)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("buffer created a file")
	}
	result = AnalyzeBuffer(engine, path, []byte("<?php function {"), syntax.Options{})
	if len(result.Errors) == 0 || len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "syntax-recovery" {
		t.Fatal("syntax errors missing")
	}
}

func TestEmptyProject(t *testing.T) {
	p, err := Open(nil, Options{Paths: []string{t.TempDir()}})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.AnalyzeReport()) != 0 || len(p.PrepareFixes()) != 0 {
		t.Fatal("empty project produced outcomes")
	}
}

type legacyBufferAnalyzer struct{}

func (legacyBufferAnalyzer) Analyze(*syntax.File) []diagnostic.Finding {
	return []diagnostic.Finding{{Rule: "legacy", Message: "legacy result"}}
}

func TestBufferSupportsLegacyAnalyzers(t *testing.T) {
	result := AnalyzeBuffer(legacyBufferAnalyzer{}, "legacy.php", []byte("<?php ;"), syntax.Options{})
	if len(result.Findings) != 1 || result.Findings[0].Rule != "legacy" || len(result.Incomplete) != 0 {
		t.Fatalf("result=%+v", result)
	}
}

func TestLegacyBufferParserStatus(t *testing.T) {
	result := AnalyzeBuffer(legacyBufferAnalyzer{}, "broken.php", []byte("<?php function {"), syntax.Options{})
	if len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "syntax-recovery" {
		t.Fatal(result)
	}
}
