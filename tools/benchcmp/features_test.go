//go:build unix

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleFeatures() *Features {
	return &Features{
		Checked: "2026-11-08",
		Tools:   []FTool{{ID: "custos", Label: "custos", URL: "https://example.com/custos"}, {ID: "other", Label: "Other | tool", URL: "https://example.com/other"}},
		Features: []Feature{
			{Name: "Written in", Cells: map[string]Cell{"custos": {Note: "Go"}, "other": {Note: "PHP"}}},
			{Name: "Fixes", Cells: map[string]Cell{"custos": {Value: "yes", Note: "`custos fix`"}, "other": {Value: "partial", Note: "some {{rules}}", Source: "https://example.com/fix"}}},
			{Name: "LSP", Cells: map[string]Cell{"custos": {Value: "yes"}, "other": {Value: "no"}}},
		},
	}
}

func TestFeaturesTable(t *testing.T) {
	got := sampleFeatures().table()
	for _, want := range []string{
		"| Feature | [custos](https://example.com/custos) | [Other \\| tool](https://example.com/other) |",
		"| Written in | Go | PHP |",
		"| Fixes | ✓ `custos fix` | [partial: some &#123;&#123;rules&#125;&#125;](https://example.com/fix) |",
		"| LSP | ✓ | — |",
		"on 2026-11-08",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	f := sampleFeatures()
	f.Features[0].Cells["other"] = Cell{Note: "PHP", Source: "https://example.com/php"}
	got = f.compact()
	for _, want := range []string{"| Written in | Go | PHP |", "| Fixes | ✓ | ◐ |", "| LSP | ✓ | — |"} {
		if !strings.Contains(got, want) {
			t.Errorf("compact: missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Checked") {
		t.Errorf("compact has the checked line:\n%s", got)
	}
}

func TestLoadFeatures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.json")
	if err := writeJSON(path, sampleFeatures()); err != nil {
		t.Fatal(err)
	}
	if f, err := loadFeatures(path); err != nil || f.table() != sampleFeatures().table() {
		t.Fatalf("round trip: %v", err)
	}
	for name, body := range map[string]string{
		"syntax":        `{`,
		"unknown field": `{"checked": "x", "tools": [{"id": "a"}], "bogus": 1}`,
		"no tools":      `{"checked": "x"}`,
		"no date":       `{"tools": [{"id": "a"}]}`,
		"unknown tool":  `{"checked": "x", "tools": [{"id": "a"}], "features": [{"name": "f", "cells": {"a": {"value": "yes"}, "b": {"value": "yes"}}}]}`,
		"missing cell":  `{"checked": "x", "tools": [{"id": "a"}, {"id": "b"}], "features": [{"name": "f", "cells": {"a": {"value": "yes"}}}]}`,
		"bad value":     `{"checked": "x", "tools": [{"id": "a"}], "features": [{"name": "f", "cells": {"a": {"value": "maybe"}}}]}`,
		"empty cell":    `{"checked": "x", "tools": [{"id": "a"}], "features": [{"name": "f", "cells": {"a": {}}}]}`,
	} {
		write(t, path, body)
		if _, err := loadFeatures(path); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if _, err := loadFeatures(filepath.Join(dir, "missing.json")); !os.IsNotExist(err) {
		t.Errorf("missing file: %v", err)
	}
}
