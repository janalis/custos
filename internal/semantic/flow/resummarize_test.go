package flow

import (
	"reflect"
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

const resummarizeSource = `
function a() { return ext(); }
function b() { return 'k'; }
function c() { return b(); }
if (PHP_OS) { function d() { return 1; } } else { function d() { return 2; } }
class P { function m() { return 1; } }
class Q extends P { function m() { return $_GET['x']; } }
function u1(P $p) { return $p->m(); }
function u2(P $p) { return $p->m(); }
function w() { global $v; $v = other(); }
function h1() { global $v; return $v; }
function h2() { global $v; return $v; }
`

func resummarizeFiles(t testing.TB) (*index.Index, *syntax.File, *syntax.File, *syntax.File) {
	t.Helper()
	parse := func(path, src string) *syntax.File {
		f := syntax.Parse(path, []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
		if len(f.Errors) > 0 {
			t.Fatalf("parse %s: %v", path, f.Errors)
		}
		return f
	}
	x := parse("x.php", resummarizeSource)
	before := parse("e.php", `function ext() { return 1; } function other() { return 2; }`)
	after := parse("e.php", `function ext() { return $_GET['y']; } function other() { return 2; }`)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(x))
	ix.Add(index.Extract(before))
	return ix, x, before, after
}

// A callable no changed contract reaches keeps its summary; the file comes
// out as a full summary under the new snapshot would.
func TestResummarizeMatchesSummarize(t *testing.T) {
	ix, x, before, after := resummarizeFiles(t)
	php := phpversion.Max
	e1 := Extract(before, ix, php, nil, Options{})
	prior := Extract(x, ix, php, NewSnapshot(e1), Options{})
	earlier := NewSnapshot(e1, prior)
	e2 := Extract(after, ix, php, earlier, Options{})
	now := earlier.Replace([]*File{e1}, []*File{e2})
	got := Reextract(x, ix, php, now, Options{}, prior, earlier)
	want := Extract(x, ix, php, now, Options{})
	assertSameFile(t, got, want)
	same := func(key string) bool {
		return reflect.ValueOf(got.callables[key]).Pointer() == reflect.ValueOf(prior.callables[key]).Pointer()
	}
	for key, kept := range map[string]bool{"a": false, "b": true, "c": true, "d": false, "u1": true, "u2": true, "w": true, "h1": true, "h2": true} {
		if same(key) != kept {
			t.Errorf("%s kept=%v, want %v", key, same(key), kept)
		}
	}
	// The globals that w writes were computed once, reading other(): the
	// second reader of them depends on it through the cached globals.
	if bits := prior.callables["h2"]["other"]; bits != readAny {
		t.Errorf("h2 reads other() as %d, want %d", bits, readAny)
	}
}

// Without usable earlier summaries, the file is summarized whole.
func TestResummarizeWithoutPrior(t *testing.T) {
	ix, x, before, _ := resummarizeFiles(t)
	php := phpversion.Max
	snapshot := NewSnapshot(Extract(before, ix, php, nil, Options{}))
	want := Extract(x, ix, php, snapshot, Options{})
	assertSameFile(t, Reextract(x, ix, php, snapshot, Options{}, nil, snapshot), want)
	other := Extract(x, ix, php, NewSnapshotWithOptions(Options{Assertions: true}), Options{Assertions: true})
	assertSameFile(t, Reextract(x, ix, php, snapshot, Options{}, other, snapshot), want)
}

func assertSameFile(t *testing.T, got, want *File) {
	t.Helper()
	sorted := func(keys []string) []string { out := slices.Clone(keys); slices.Sort(out); return out }
	if !reflect.DeepEqual(got.Summaries, want.Summaries) {
		t.Fatalf("summaries differ:\n%v\n%v", got.Summaries, want.Summaries)
	}
	if !slices.Equal(sorted(got.Consulted()), sorted(want.Consulted())) || !slices.Equal(got.Reads(), want.Reads()) {
		t.Fatalf("reads differ: %v %v / %v %v", got.Consulted(), got.Reads(), want.Consulted(), want.Reads())
	}
}

// A file summarized again after one contract it read changed, whole and
// keeping the callables the change does not reach.
func BenchmarkReextract(b *testing.B) {
	ix, x, before, after := resummarizeFiles(b)
	php := phpversion.Max
	e1 := Extract(before, ix, php, nil, Options{})
	prior := Extract(x, ix, php, NewSnapshot(e1), Options{})
	earlier := NewSnapshot(e1, prior)
	now := earlier.Replace([]*File{e1}, []*File{Extract(after, ix, php, earlier, Options{})})
	b.Run("whole", func(b *testing.B) {
		for b.Loop() {
			Extract(x, ix, php, now, Options{})
		}
	})
	b.Run("again", func(b *testing.B) {
		for b.Loop() {
			Reextract(x, ix, php, now, Options{}, prior, earlier)
		}
	})
}
