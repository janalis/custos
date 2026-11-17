package flow

import "testing"

// Summaries read another file's contracts when they consulted one they did
// not serve themselves, whatever the case of its name.
func TestReadsOthers(t *testing.T) {
	for _, c := range []struct {
		file *File
		want bool
	}{
		{&File{}, false},
		{&File{consulted: map[string]bool{"Own": true}, served: map[string]bool{"own": true}}, false},
		{&File{consulted: map[string]bool{"Own": true, "other": true}, served: map[string]bool{"own": true}}, true},
	} {
		if got := c.file.ReadsOthers(); got != c.want {
			t.Errorf("%v: ReadsOthers %v", c.file.consulted, got)
		}
	}
}
