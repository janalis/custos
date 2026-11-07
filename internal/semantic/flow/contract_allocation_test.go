package flow

import "testing"

func TestPublishedShapeCompactionDoesNotAllocate(t *testing.T) {
	e, file := environment(t, `function row(){return ['a'=>'safe','b'=>['nested'=>'text']];}`, nil)
	published := Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	original := published.Summaries["row"].Return.elements
	var compact *elementSet
	allocations := testing.AllocsPerRun(100, func() {
		fuel := MaxSources
		compact = compactElements(original, MaxDepth, &fuel)
	})
	if compact != original || allocations != 0 {
		t.Fatalf("unchanged published shape: reused=%v allocations=%v", compact == original, allocations)
	}
}
