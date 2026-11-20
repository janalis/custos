package format

import (
	"sort"
	"testing"
)

// apply replays changes on src.
func apply(src []byte, changes []Change) []byte {
	cs := append([]Change(nil), changes...)
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].Start < cs[j].Start })
	var out []byte
	pos := uint32(0)
	for _, c := range cs {
		if c.Report {
			continue
		}
		out = append(out, src[pos:c.Start]...)
		out = append(out, c.New...)
		pos = c.End
	}
	return append(out, src[pos:]...)
}

func TestComposeReplays(t *testing.T) {
	src := []byte("abcdef")
	// first: "abcdef" -> "aXcdf" (b->X, delete e), second: "aXcdf" -> "aXYcd!f"
	first := []Change{{Rule: "r1", Start: 1, End: 2, New: "X"}, {Rule: "r2", Start: 4, End: 5, New: ""}}
	mid := apply(src, first)
	second := []Change{{Rule: "r3", Start: 2, End: 2, New: "Y"}, {Rule: "r4", Start: 4, End: 4, New: "!"}}
	out := apply(mid, second)
	got := apply(src, compose(first, second, out))
	if string(got) != string(out) {
		t.Fatalf("compose replays %q, want %q (mid %q)", got, out, mid)
	}
}
