package flow

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"reflect"
	"testing"
)

// Replace builds the snapshot NewSnapshotWithOptions builds from the files
// with the replaced ones swapped, through any sequence of replacements:
// names declared by one, two or more files, files dropped, files carrying
// other options, and an overlay as the receiver.
func TestSnapshotReplaceMatchesRebuild(t *testing.T) {
	opt, other := Options{}, Options{Assertions: true}
	names := []string{"a", "b", "c", "d", "e", "f::g"}
	rng := rand.New(rand.NewPCG(1, 2))
	serial := 0
	random := func(path string) *File {
		switch rng.IntN(6) {
		case 0:
			return nil
		case 1:
			return &File{options: other, Path: path, Summaries: map[string]Summary{"a": {}}}
		}
		f := &File{options: opt, Path: path, Summaries: map[string]Summary{}}
		for _, name := range names {
			if rng.IntN(3) == 0 {
				serial++
				f.Summaries[name] = Summary{defaults: make([]Value, serial%5)}
			}
		}
		return f
	}
	for trial := range 200 {
		results := make([]*File, 8)
		for i := range results {
			results[i] = random(fmt.Sprint(i))
		}
		snapshot := NewSnapshotWithOptions(opt, results...)
		if trial%4 == 0 && results[0] != nil && results[0].options == opt {
			snapshot = snapshot.WithFile(results[0])
		}
		for round := range 6 {
			var prev, fresh []*File
			for i := range results {
				if rng.IntN(3) == 0 {
					prev = append(prev, results[i])
					results[i] = random(fmt.Sprint(i))
					fresh = append(fresh, results[i])
				}
			}
			snapshot = snapshot.Replace(prev, fresh)
			assertSameSnapshot(t, fmt.Sprintf("trial %d round %d", trial, round), snapshot, NewSnapshotWithOptions(opt, results...))
		}
	}
}

// A file replacing one the snapshot holds at its path, without that file
// among the replaced ones, still takes its place.
func TestSnapshotReplaceSamePath(t *testing.T) {
	old := &File{Path: "p", Summaries: map[string]Summary{"a": {}, "b": {}}}
	keep := &File{Path: "q", Summaries: map[string]Summary{"b": {}}}
	fresh := &File{Path: "p", Summaries: map[string]Summary{"c": {}}}
	got := NewSnapshotWithOptions(Options{}, old, keep).Replace(nil, []*File{fresh})
	assertSameSnapshot(t, "same path", got, NewSnapshotWithOptions(Options{}, keep, fresh))
	// One with other options removes the file at its path, as a rebuild
	// leaves it out.
	foreign := &File{options: Options{Assertions: true}, Path: "p", Summaries: map[string]Summary{"c": {}}}
	got = NewSnapshotWithOptions(Options{}, old, keep).Replace(nil, []*File{foreign})
	assertSameSnapshot(t, "other options", got, NewSnapshotWithOptions(Options{}, keep, foreign))
	// A nil snapshot is the empty one, as for lookups.
	var none *Snapshot
	assertSameSnapshot(t, "nil", none.Replace(nil, []*File{keep}), NewSnapshotWithOptions(Options{}, keep))
}

func assertSameSnapshot(t *testing.T, label string, got, want *Snapshot) {
	t.Helper()
	nonEmpty := func(m map[string]int) map[string]int {
		if len(m) == 0 {
			return nil
		}
		return maps.Clone(m)
	}
	if !reflect.DeepEqual(got.files, want.files) || !reflect.DeepEqual(got.summaries, want.summaries) || !reflect.DeepEqual(nonEmpty(got.shared), nonEmpty(want.shared)) || got.options != want.options {
		t.Fatalf("%s:\nfiles %v / %v\nsummaries %v / %v\nshared %v / %v", label, got.files, want.files, got.summaries, want.summaries, got.shared, want.shared)
	}
}

// A laravel-sized round of compose: 1746 files of about a dozen contracts,
// 100 of them summarized again.
func BenchmarkSnapshotRound(b *testing.B) {
	files := make([]*File, 1746)
	for i := range files {
		f := &File{Path: fmt.Sprint("src/", i, ".php"), Summaries: map[string]Summary{}}
		for j := range 12 {
			f.Summaries[fmt.Sprint("ns\\class", i, "::m", j)] = Summary{}
		}
		files[i] = f
	}
	snapshot := NewSnapshotWithOptions(Options{}, files...)
	prev, fresh := files[:100], make([]*File, 100)
	for i, f := range prev {
		fresh[i] = &File{Path: f.Path, Summaries: maps.Clone(f.Summaries)}
	}
	results := append(append([]*File{}, fresh...), files[100:]...)
	b.Run("rebuild", func(b *testing.B) {
		for b.Loop() {
			NewSnapshotWithOptions(Options{}, results...)
		}
	})
	b.Run("replace", func(b *testing.B) {
		for b.Loop() {
			snapshot.Replace(prev, fresh)
		}
	})
}

// After Replace, a file reading a contract that was kept stays current and
// one reading a contract that changed does not.
func TestSnapshotReplaceCurrent(t *testing.T) {
	a := &File{Path: "a", Summaries: map[string]Summary{"kept": {}, "changed": {}}}
	before := NewSnapshotWithOptions(Options{}, a)
	after := before.Replace([]*File{a}, []*File{{Path: "a", Summaries: map[string]Summary{"kept": {}, "changed": {Generator: true}}}})
	after = after.Replace(nil, nil)
	for name, want := range map[string]bool{"kept": true, "changed": false} {
		reader := &File{Path: "r", consulted: map[string]bool{name: true}}
		if got := reader.Current(before, after); got != want {
			t.Errorf("reader of %s: Current=%v want %v", name, got, want)
		}
	}
	kept := &File{Path: "r", consulted: map[string]bool{"changed": true}}
	if !kept.Current(after, after.Replace(nil, nil)) {
		t.Error("a contract Replace kept is current")
	}
}

// A reader of 40 contracts checked against the next round's snapshot, as
// compose's outdated check does: Replace keeps each contract's pointer.
func BenchmarkCurrentAfterReplace(b *testing.B) {
	files := make([]*File, 200)
	consulted := map[string]bool{}
	for i := range files {
		f := &File{Path: fmt.Sprint(i), Summaries: map[string]Summary{}}
		for j := range 10 {
			name := fmt.Sprint("ns\\c", i, "::m", j)
			f.Summaries[name] = Summary{defaults: make([]Value, 2), Sinks: make([]Sink, 1)}
			if len(consulted) < 40 && j == 0 {
				consulted[name] = true
			}
		}
		files[i] = f
	}
	before := NewSnapshotWithOptions(Options{}, files...)
	after := before.Replace(files[199:], []*File{{Path: "199"}})
	reader := &File{Path: "r", consulted: consulted}
	for b.Loop() {
		if !reader.Current(before, after) {
			b.Fatal("stale")
		}
	}
}
