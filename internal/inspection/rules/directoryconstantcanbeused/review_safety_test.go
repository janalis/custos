package directoryconstantcanbeused

import (
	"testing"

	"custos/internal/inspection/analysis"
)

// The project index is only built for semantic rules; without the marker a
// namespaced dirname() declared in another file would not be seen (D1).
func TestReviewSafetySemanticMarker(t *testing.T) {
	r, ok := New().(analysis.SemanticRule)
	if !ok {
		t.Fatal("rule must be semantic to resolve dirname() across files")
	}
	r.Semantic()
}
