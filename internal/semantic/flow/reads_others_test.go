package flow

import "testing"

// Summaries read another file's contracts when they consulted one they did
// not serve themselves, whatever the case of its name, or one they served
// that another file of the snapshot declares too.
func TestReadsOthers(t *testing.T) {
	shared := NewSnapshot(&File{Path: "a.php", Summaries: map[string]Summary{"own": {}}}, &File{Path: "b.php", Summaries: map[string]Summary{"own": {}}})
	for _, c := range []struct {
		file     *File
		snapshot *Snapshot
		want     bool
	}{
		{&File{}, nil, false},
		{&File{Path: "a.php", consulted: map[string]bool{"Own": true}, served: map[string]bool{"own": true}}, nil, false},
		{&File{Path: "a.php", consulted: map[string]bool{"Own": true, "other": true}, served: map[string]bool{"own": true}}, nil, true},
		{&File{Path: "a.php", consulted: map[string]bool{"Own": true}, served: map[string]bool{"own": true}}, shared, true},
	} {
		if got := c.file.ReadsOthers(c.snapshot); got != c.want {
			t.Errorf("%v: ReadsOthers %v", c.file.consulted, got)
		}
	}
}
