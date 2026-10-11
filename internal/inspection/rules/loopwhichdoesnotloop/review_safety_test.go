package loopwhichdoesnotloop

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/stubs"
)

// TestReviewSafety covers generator methods declared in another file: the
// rule relies on the return type the index derived from their bodies.
func TestReviewSafety(t *testing.T) {
	opt := syntax.Options{Version: phpversion.PHP85}
	feed := syntax.ParseBest("RemoteFeed.php", []byte(`<?php
class RemoteFeed {
    public function items() { yield 'first'; yield 'second'; }
    public function titles() { return ['first', 'second']; }
    public function batches() { return [$this->items()]; }
}`), opt)
	fs := index.Extract(feed)
	infer.AnnotateReturns(feed, fs, stubs.Index(), opt.Version)
	ix := index.New(stubs.Index())
	ix.Add(fs)
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	if !e.NeedsIndex() {
		t.Fatal("the rule must request the project index")
	}
	e = e.WithIndex(ix)
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"remote generator", "<?php function f(RemoteFeed $r) { foreach ($r->items() as $i) { return $i; } }", 0},
		{"remote generator break", "<?php function f(RemoteFeed $r) { foreach ($r->items() as $i) { break; } }", 0},
		{"remote array", "<?php function f(RemoteFeed $r) { foreach ($r->titles() as $i) { return $i; } }", 1},
		{"remote array of generators", "<?php function f(RemoteFeed $r) { foreach ($r->batches() as $i) { return $i; } }", 1},
		{"remote generator throw", "<?php function f(RemoteFeed $r) { foreach ($r->items() as $i) { throw new \\Exception(); } }", 1},
		{"union receiver", "<?php class Local { function items() { return [1]; } } function f(Local|RemoteFeed $r) { foreach ($r->items() as $i) { break; } }", 0},
		{"local array first", "<?php class Local { function items() { return [1]; } } function f(Local $r) { foreach ($r->items() as $i) { break; } }", 1},
		{"unknown method", "<?php function f(RemoteFeed $r) { foreach ($r->nope() as $i) { break; } }", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
