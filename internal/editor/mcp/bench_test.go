package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
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
