package syntax

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	phpversion "github.com/janalis/custos/internal/php/version"
)

// largeCorpusFile is a large real-world file the walker tests and benchmark
// add to the rule fixtures when the bench corpora are checked out
// (make compare).
const largeCorpusFile = "../../../.cache/bench/corpora/laravel/src/Illuminate/Database/Query/Builder.php"

var walkFixtures = sync.OnceValues(func() ([]*File, error) {
	var paths []string
	err := filepath.WalkDir("../../inspection/rules", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".php") {
			paths = append(paths, p)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	paths = append(paths, "testdata/cov/allkinds.php")
	if _, err := os.Stat(largeCorpusFile); err == nil {
		paths = append(paths, largeCorpusFile)
	}
	files := make([]*File, 0, len(paths))
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		files = append(files, Parse(p, src, Options{Version: phpversion.PHP85}))
	}
	return files, nil
})

func loadWalkFixtures(tb testing.TB) []*File {
	tb.Helper()
	files, err := walkFixtures()
	if err != nil {
		tb.Fatal(err)
	}
	return files
}

var nodeType = reflect.TypeFor[Node]()

// reachable collects every non-nil node reachable from v through exported
// fields (node pointers, node interfaces and slices of them).
func reachable(v reflect.Value, seen map[Node]bool) {
	switch v.Kind() {
	case reflect.Interface:
		if !v.IsNil() {
			reachable(v.Elem(), seen)
		}
	case reflect.Slice:
		for i := range v.Len() {
			reachable(v.Index(i), seen)
		}
	case reflect.Pointer:
		if v.IsNil() || !v.Type().Implements(nodeType) {
			return
		}
		n := v.Interface().(Node)
		if seen[n] {
			return
		}
		seen[n] = true
		s := v.Elem()
		for i := range s.NumField() {
			if s.Type().Field(i).IsExported() {
				reachable(s.Field(i), seen)
			}
		}
	}
}

// TestChildrenVisitsEveryField checks that the walk reaches every node held
// in a field of a parsed node, so a field added to a node struct cannot be
// forgotten by Children (and, through TestInspectMatchesChildren, Inspect).
func TestChildrenVisitsEveryField(t *testing.T) {
	for _, f := range loadWalkFixtures(t) {
		want := map[Node]bool{}
		reachable(reflect.ValueOf(f), want)
		got := map[Node]bool{}
		Inspect(f, func(n Node) bool {
			if !reflect.ValueOf(n).IsNil() {
				got[n] = true
			}
			return true
		})
		for n := range want {
			if !got[n] {
				t.Fatalf("%s: %s at %v is held by a field but never walked", f.Path, n.Kind(), n.Span())
			}
		}
		for n := range got {
			if !want[n] {
				t.Fatalf("%s: %s at %v walked but not held by a field", f.Path, n.Kind(), n.Span())
			}
		}
	}
}

func TestChildrenSkipsTypedNils(t *testing.T) {
	name := &Name{Value: "f"}
	nodes := []Node{
		&ArrayDimFetch{Var: (*Variable)(nil), Dim: (*Name)(nil)},
		&PropertyFetch{Var: (*ClassLike)(nil), Name: (*Identifier)(nil)},
		&If{Cond: name, Body: (*Block)(nil)},
		&Block{Stmts: []Stmt{(*ClassLike)(nil), (*Block)(nil)}},
		&PropertyHook{Name: &Identifier{}, Body: (*Block)(nil)},
		&PropertyHook{Name: &Identifier{}, Body: Expr((*Name)(nil))},
	}
	want := []int{0, 0, 1, 0, 1, 1}
	for i, n := range nodes {
		got := 0
		Children(n, func(c Node) {
			if reflect.ValueOf(c).IsNil() {
				t.Fatalf("%s: visited typed nil %s", n.Kind(), c.Kind())
			}
			got++
		})
		if got != want[i] {
			t.Fatalf("%s #%d: %d children, want %d", n.Kind(), i, got, want[i])
		}
	}
	var body []Node
	Children(&PropertyHook{Name: &Identifier{}, Body: name}, func(c Node) { body = append(body, c) })
	if len(body) != 2 || body[1] != name {
		t.Fatalf("hook expression body not walked: %v", body)
	}
}

// BenchmarkInspect measures the walk machinery alone (fn only counts) over
// the rule fixtures and, when the bench corpora are checked out, one large
// real file.
func BenchmarkInspect(b *testing.B) {
	files := loadWalkFixtures(b)
	sets := []struct {
		name  string
		files []*File
	}{{"fixtures", files}}
	if last := files[len(files)-1]; last.Path == largeCorpusFile {
		sets[0].files = files[:len(files)-1]
		sets = append(sets, struct {
			name  string
			files []*File
		}{"laravel-query-builder", files[len(files)-1:]})
	}
	for _, set := range sets {
		b.Run(set.name, func(b *testing.B) {
			nodes := 0
			b.ReportAllocs()
			for b.Loop() {
				nodes = 0
				for _, f := range set.files {
					InspectFile(f, func(Node) bool { nodes++; return true })
				}
			}
			b.ReportMetric(float64(nodes), "nodes/op")
		})
	}
}
