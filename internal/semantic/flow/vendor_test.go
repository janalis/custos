package flow

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// vendorEnvironment analyzes the project source with the given dependency
// files indexed under vendor/ and summarized on demand.
func vendorEnvironment(t testing.TB, project string, files map[string]string) (*Env, *syntax.File, *Vendor, *int) {
	t.Helper()
	opt := syntax.Options{Version: phpversion.Max}
	ix := index.New(stubs.Index())
	for path, src := range files {
		ix.Add(index.Extract(syntax.Parse(path, []byte("<?php "+src), opt)))
	}
	f := syntax.Parse("src/app.php", []byte("<?php "+project), opt)
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix.Add(index.Extract(f))
	reads := new(int)
	var mu sync.Mutex
	read := func(path string) ([]byte, error) {
		mu.Lock()
		*reads++
		mu.Unlock()
		src, ok := files[path]
		if !ok {
			return nil, errors.New("missing")
		}
		return []byte("<?php " + src), nil
	}
	vendor := NewVendor(ix, opt, false, func(path string) bool { return strings.HasPrefix(path, "vendor/") }, read)
	options := Options{Vendor: vendor}
	types := infer.NewEnv(f, names.New(f), ix, phpversion.Max, infer.Options{})
	own := Extract(f, ix, phpversion.Max, nil, options)
	return New(f, types, NewSnapshotWithOptions(options, own), options), f, vendor, reads
}

func TestVendorContracts(t *testing.T) {
	for _, tc := range []struct {
		name, project     string
		files             map[string]string
		resolved, tainted bool
	}{
		{"function", `probe(\Lib\pass($_GET['x']));`, map[string]string{"vendor/lib/f.php": `namespace Lib; function pass($x) { return $x; }`}, true, true},
		{"escaping", `probe(\Lib\clean($_GET['x']));`, map[string]string{"vendor/lib/f.php": `namespace Lib; function clean($x) { return htmlspecialchars($x); }`}, true, false},
		{"final method", `function f(\Lib\Box $b) { probe($b->get($_GET['x'])); }`, map[string]string{"vendor/lib/Box.php": `namespace Lib; final class Box { public function get($x) { return $x; } }`}, true, true},
		{"open method", `function f(\Lib\Box $b) { probe($b->get($_GET['x'])); }`, map[string]string{"vendor/lib/Box.php": `namespace Lib; class Box { public function get($x) { return $x; } }`}, true, true},
		{"ambiguous", `probe(\Lib\pass($_GET['x']));`, map[string]string{"vendor/a/f.php": `namespace Lib; function pass($x) { return $x; }`, "vendor/b/f.php": `namespace Lib; function pass($x) { return $x; }`}, false, false},
		{"outside vendor", `probe(\Lib\pass($_GET['x']));`, map[string]string{"lib/f.php": `namespace Lib; function pass($x) { return $x; }`}, false, false},
		{"unreadable", `probe(\Lib\pass($_GET['x']));`, map[string]string{"vendor/lib/f.php": `namespace Lib; function pass($x) { return $x; }` + strings.Repeat(" ", MaxVendorBytes)}, false, false},
		{"unknown", `probe(\Lib\missing($_GET['x']));`, map[string]string{}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f, _, _ := vendorEnvironment(t, `function probe($x) {} `+tc.project, tc.files)
			x := argument(callsNamed(f, "probe")[0])
			e.Value(x)
			resolved := true
			for _, gap := range e.Gaps() {
				resolved = resolved && gap.Reason != "unresolved-dispatch"
			}
			if resolved != tc.resolved || e.Tainted(x, HTML) != tc.tainted {
				t.Fatalf("resolved=%v tainted=%v gaps=%v", resolved, e.Tainted(x, HTML), e.Gaps())
			}
		})
	}
}

// A project declaration of the same callable is the project's contract; the
// dependency copy makes it ambiguous rather than replacing it.
func TestVendorNeverAnswersProjectNames(t *testing.T) {
	e, f, vendor, reads := vendorEnvironment(t, `function probe($x) {} probe(\Lib\pass($_GET['x']));`, map[string]string{"vendor/lib/f.php": `namespace Lib; function pass($x) { return $x; }`})
	file := Extract(f, e.types.Index, phpversion.Max, nil, Options{Vendor: vendor})
	project := `namespace Lib; function pass($x) { return 'fixed'; }`
	pf := syntax.Parse("src/lib.php", []byte("<?php "+project), syntax.Options{Version: phpversion.Max})
	own := Extract(pf, e.types.Index, phpversion.Max, nil, Options{Vendor: vendor})
	snapshot := NewSnapshotWithOptions(Options{Vendor: vendor}, file, own)
	if summary, ok := snapshot.lookup(`lib\pass`); !ok || len(summary.Return.Sources) != 0 {
		t.Fatalf("project declaration not preferred: %+v %v", summary.Return, ok)
	}
	if *reads != 0 {
		t.Fatal("dependency read for a project name")
	}
	if _, ok := snapshot.WithFile(own).lookup(`lib\other`); ok {
		t.Fatal("unknown name answered")
	}
}

func TestVendorSummarizesEachFileOnce(t *testing.T) {
	files := map[string]string{"vendor/lib/f.php": `namespace Lib; function a($x) { return $x; } function b($x) { return a($x); } final class C { function m() {} }`}
	_, _, vendor, reads := vendorEnvironment(t, ``, files)
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := []string{`lib\a`, `lib\b`, `lib\c::m`}[i%3]
			if _, ok := vendor.lookup(key); !ok {
				t.Errorf("%s unanswered", key)
			}
		}()
	}
	wg.Wait()
	if *reads != 1 {
		t.Fatalf("reads=%d", *reads)
	}
	// Within the file, b composes a: its summary was made with the file's own.
	if summary, _ := vendor.lookup(`lib\b`); summary.Return.Complete != true {
		t.Fatalf("intra-file composition lost: %+v", summary.Return)
	}
	if _, ok := vendor.lookup(sinkClassKey(`Lib\C`)); !ok {
		t.Fatal("class metadata unanswered")
	}
	for _, key := range []string{`lib\c::`, `lib\missing::m`, `lib\c::missing`, sinkClassKey(`Lib\Missing`)} {
		if _, ok := vendor.lookup(key); ok {
			t.Fatalf("%s answered", key)
		}
	}
}

func TestOptionsPrintWithoutProvider(t *testing.T) {
	vendor := NewVendor(index.New(nil), syntax.Options{}, true, nil, nil)
	if got := (Options{Assertions: true, Vendor: vendor}).String(); got != "{true}" {
		t.Fatal(got)
	}
	if (*Snapshot)(nil).Options() != (Options{}) {
		t.Fatal("snapshot options")
	}
}
