package flow

import "testing"

// Basis answers, for every contract a file consulted, what the snapshot it
// was made from answers: held, ambiguous or absent. It keeps nothing else.
func TestBasis(t *testing.T) {
	a := &File{Path: "a", Summaries: map[string]Summary{"single": {}, "changed": {}, "unread": {}}}
	b := &File{Path: "b", Summaries: map[string]Summary{"shared": {}}}
	c := &File{Path: "c", Summaries: map[string]Summary{"shared": {Generator: true}}}
	before := NewSnapshotWithOptions(Options{}, a, b, c)
	reader := &File{Path: "r", consulted: map[string]bool{"Single": true, "shared": true, "absent": true, "changed": true}}
	basis := reader.Basis(before)
	if basis.held("unread") != nil || len(basis.summaries) != 2 || basis.shared["shared"] != 2 {
		t.Fatalf("basis keeps %v, shared %v", basis.summaries, basis.shared)
	}
	for name := range reader.consulted {
		was, had := before.lookup(name)
		is, has := basis.lookup(name)
		if before.held(name) != basis.held(name) || had != has || had && !sameSummary(&was, &is) {
			t.Errorf("%s: basis answers differently", name)
		}
	}
	after := before.Replace([]*File{a}, []*File{{Path: "a", Summaries: map[string]Summary{"single": {}, "changed": {Generator: true}}}})
	same := before.Replace(nil, nil)
	for _, s := range []*Snapshot{after, same} {
		if reader.Current(basis, s) != reader.Current(before, s) {
			t.Error("Current differs against the basis")
		}
	}
	if reader.Basis(nil) != nil {
		t.Error("basis of no snapshot")
	}
	if overlay := before.WithFile(a); reader.Basis(overlay) != overlay {
		t.Error("an overlay is its own basis")
	}
}
