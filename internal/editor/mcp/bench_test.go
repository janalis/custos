package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
	"custos/internal/project"
)

func BenchmarkSnapshotReuse(b *testing.B) {
	s, ctx, _ := fixture(b)
	cp := invoke(ctx, b, s, "analyze", Input{})["checkpointId"].(string)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := s.call(ctx, "findings", Input{CheckpointID: cp}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPagination(b *testing.B) {
	items := make([]any, 1000)
	for i := range items {
		items[i] = Output{"id": i, "message": "Use the simpler expression."}
	}
	ctx := context.WithValue(context.Background(), identityKey{}, "benchmark")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := page(ctx, "benchmark", Input{}, items); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNavigation(b *testing.B) {
	s, ctx, _ := fixture(b)
	cp := invoke(ctx, b, s, "analyze", Input{})["checkpointId"].(string)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := s.call(ctx, "symbols", Input{CheckpointID: cp, Query: "sample"}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFixPreviewPage(b *testing.B) {
	ctx := context.WithValue(context.Background(), identityKey{}, "benchmark")
	text := strings.Repeat("é", 1<<19)
	p := &plan{ID: randomID(), Expires: time.Now().Add(time.Hour), Changes: []change{{Path: "a.php", Edits: []diagnostic.TextEdit{{Span: syntax.Span{End: 1}, NewText: text}}}}}
	s := &Server{}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := s.planPage(ctx, p, Input{}); err != nil {
			b.Fatal(err)
		}
	}
}

// vendorFixture adds a dependency tree that dominates the project inputs.
func vendorFixture(b *testing.B) (*Server, context.Context, string) {
	s, ctx, root := fixture(b)
	old := time.Now().Add(-time.Hour)
	for pkg := range 20 {
		dir := filepath.Join(root, "vendor", fmt.Sprintf("pkg%02d", pkg), "src")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			b.Fatal(err)
		}
		for file := range 25 {
			p := filepath.Join(dir, fmt.Sprintf("Type%02d.php", file))
			writeFile(b, p, fmt.Sprintf("<?php\nnamespace Pkg%02d;\n\nclass Type%02d { public function value(): int { return %d; } }\n", pkg, file, file))
			if err := os.Chtimes(p, old, old); err != nil {
				b.Fatal(err)
			}
		}
	}
	return s, ctx, root
}

func BenchmarkVendorAnalyzeCold(b *testing.B) {
	s, ctx, _ := vendorFixture(b)
	w := s.workspaces["default"]
	b.ReportAllocs()
	for b.Loop() {
		w.cache = &project.VendorCache{}
		invoke(ctx, b, s, "analyze", Input{})
	}
}

// Unchanged dependencies are neither read nor parsed by a later analysis.
func BenchmarkVendorAnalyzeWarm(b *testing.B) {
	s, ctx, _ := vendorFixture(b)
	invoke(ctx, b, s, "analyze", Input{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		invoke(ctx, b, s, "analyze", Input{})
	}
}

// Paging validates the checkpoint by size and modification time.
func BenchmarkVendorValidation(b *testing.B) {
	s, ctx, _ := vendorFixture(b)
	cp := invoke(ctx, b, s, "analyze", Input{})["checkpointId"].(string)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := s.call(ctx, "findings", Input{CheckpointID: cp}); err != nil {
			b.Fatal(err)
		}
	}
}
