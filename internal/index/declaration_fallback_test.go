package index

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/testbudget"
)

func TestDeclarationFallbackChanges(t *testing.T) {
	calls := extract(t, "calls.php", `<?php
namespace App;
class Original {}
define('FLAG', 7);
class_alias(Original::class, 'Alias');
\define('EXPLICIT', 8);
use function define as register;
register('IMPORTED', 9);
\class_alias(Original::class, 'GlobalAlias');
use function class_alias as rename;
rename(Original::class, 'ImportedAlias');
`)
	shadow := extract(t, "shadow.php", `<?php namespace App; function DEFINE() {} function CLASS_ALIAS() {}`)
	for _, reversed := range []bool{false, true} {
		t.Run(fmt.Sprint(reversed), func(t *testing.T) {
			ix := New(nil)
			if reversed {
				ix.Add(shadow)
			}
			ix.Add(calls)
			if !reversed {
				if ix.Constant("FLAG", 0) == nil || len(ix.Ancestors("Alias", 0)) != 1 {
					t.Fatal("unshadowed declarations missing")
				}
				ix.Add(shadow)
			}
			for _, ver := range []phpver.Version{0, phpver.PHP53, phpver.PHP85} {
				if ix.Constant("FLAG", ver) != nil || ix.Class("Alias", ver) != nil || len(ix.Ancestors("Alias", ver)) != 0 {
					t.Fatal("cross-file shadow did not suppress declarations and cached hierarchy")
				}
				if ix.Constant("EXPLICIT", ver) == nil || ix.Constant("IMPORTED", ver) == nil || ix.Class("GlobalAlias", ver) == nil || ix.Class("ImportedAlias", ver) == nil {
					t.Fatal("explicit global or imported builtin was suppressed")
				}
			}
			ix.Add(extract(t, "shadow.php", "<?php"))
			if ix.Constant("FLAG", 0) == nil || ix.Class("Alias", 0) == nil || len(ix.Ancestors("Alias", 0)) != 1 {
				t.Fatal("replacement did not restore declarations")
			}
			ix.Add(shadow)
			ix.Remove("shadow.php")
			if ix.Constant("FLAG", 0) == nil || ix.Class("Alias", 0) == nil {
				t.Fatal("removal did not restore declarations")
			}
			if len(calls.Constants) != 3 || len(calls.ClassAliases) != 3 {
				t.Fatal("lookup mutated original symbols")
			}
		})
	}
}

func TestDeclarationFallbackDuplicatePrecedence(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php class First {} define('FLAG', 1); class_alias('First', 'Alias');`))
	ix.Add(extract(t, "b.php", `<?php namespace { class Last {} } namespace App { define('FLAG', 2); const OTHER = 1; class_alias('Last', 'Alias'); class_alias('Last', 'Alias'); }`))
	if ix.Constant("FLAG", 0).Value != "1" || ix.Class("Alias", 0).FQN != "Last" {
		t.Fatal("initial constant/alias precedence changed")
	}
	ix.Add(extract(t, "shadow.php", `<?php namespace App; function class_alias() {}`))
	if ix.Class("Alias", 0).FQN != "First" {
		t.Fatal("shadowed later alias hid earlier declaration")
	}
	ix.Remove("a.php")
	if ix.Constant("FLAG", 0).Value != "2" || ix.Class("Alias", 0) != nil {
		t.Fatal("removal did not retain correct candidates")
	}
	ix.Remove("shadow.php")
	if ix.Class("Alias", 0).FQN != "Last" {
		t.Fatal("later alias was not restored")
	}
	ix.Remove("b.php")
	if ix.Class("Alias", 0) != nil || ix.Constant("FLAG", 0) != nil {
		t.Fatal("removal left stale candidates")
	}
}

func TestDeclarationFallbackLayers(t *testing.T) {
	project := New(nil)
	project.Add(extract(t, "calls.php", `<?php namespace App; class Original {} define('FLAG', 7); class_alias(Original::class, 'Alias');`))
	project.Add(extract(t, "shadow.php", `<?php namespace App; function define() {} function class_alias() {}`))
	buffer := New(project)
	buffer.Add(extract(t, "shadow.php", "<?php"))
	if buffer.Constant("FLAG", 0) == nil || buffer.Class("Alias", 0) == nil {
		t.Fatal("buffer did not hide saved shadow functions")
	}
	buffer.Add(extract(t, "calls.php", "<?php"))
	if buffer.Constant("FLAG", 0) != nil || buffer.Class("Alias", 0) != nil {
		t.Fatal("buffer did not hide saved declaration calls")
	}
	buffer.Remove("calls.php")
	buffer.Remove("shadow.php")
	if buffer.Constant("FLAG", 0) != nil || buffer.Class("Alias", 0) != nil {
		t.Fatal("removing buffer did not restore saved shadow functions")
	}
	project.Remove("shadow.php")
	if buffer.Constant("FLAG", 0) == nil || buffer.Class("Alias", 0) == nil {
		t.Fatal("changes in base layer were ignored")
	}
}

func TestWithoutProvisionalDeclarations(t *testing.T) {
	plain := extract(t, "plain.php", `<?php define('FLAG', 1); class_alias('First', 'Alias');`)
	if plain.WithoutProvisionalDeclarations() != plain {
		t.Fatal("unnecessary view allocation")
	}
	fs := extract(t, "calls.php", `<?php namespace App; class Original {} function read() {} define('FLAG', 1); class_alias(Original::class, 'Alias'); \define('GLOBAL', 2); \class_alias(Original::class, 'GlobalAlias');`)
	view := fs.WithoutProvisionalDeclarations()
	if view == fs || len(view.Constants) != 1 || view.Constants[0].FQN != "GLOBAL" || len(view.ClassAliases) != 1 || view.ClassAliases[0][0] != "GlobalAlias" || len(view.ClassAliasFallbacks) != 0 {
		t.Fatalf("incorrect annotation view: %+v", view)
	}
	if view.Functions[0] != fs.Functions[0] || view.Classes[0] != fs.Classes[0] || len(fs.Constants) != 2 || len(fs.ClassAliases) != 2 {
		t.Fatal("view changed declarations or detached annotation targets")
	}
}

func TestDeclarationFallbackConcurrentChanges(t *testing.T) {
	ix := New(nil)
	calls := extract(t, "calls.php", `<?php namespace App; class Original {} define('FLAG', 7); class_alias(Original::class, 'Alias');`)
	ix.Add(calls)
	shadow := extract(t, "shadow.php", `<?php namespace App; function define() {} function class_alias() {}`)
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 200 {
				ix.Constant("FLAG", 0)
				ix.Class("Alias", 0)
			}
		})
	}
	for range 200 {
		ix.Add(shadow)
		ix.Add(calls)
		ix.Remove("shadow.php")
		ix.Remove("calls.php")
		ix.Add(calls)
	}
	wg.Wait()
}

func BenchmarkDeclarationFallbackLookup(b *testing.B) {
	for _, guarded := range []bool{false, true} {
		for _, shadowed := range []bool{false, true} {
			b.Run(fmt.Sprintf("guarded=%v/shadowed=%v", guarded, shadowed), func(b *testing.B) {
				src := `<?php class Original {} define('FLAG', 7); class_alias('Original', 'Alias');`
				if guarded {
					src = `<?php namespace App; class Original {} define('FLAG', 7); class_alias(Original::class, 'Alias');`
				}
				ix := New(nil)
				ix.Add(Extract(syntax.Parse("calls.php", []byte(src), syntax.Options{Version: phpver.PHP85})))
				if shadowed {
					ix.Add(Extract(syntax.Parse("shadow.php", []byte(`<?php namespace App; function define() {} function class_alias() {}`), syntax.Options{Version: phpver.PHP85})))
				}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					ix.Constant("FLAG", phpver.PHP85)
					ix.Class("Alias", phpver.PHP85)
				}
			})
		}
	}
}

func TestConstantAvailabilityFallbackWithDeclarationGuards(t *testing.T) {
	ix := New(nil)
	first := &Constant{FQN: "FLAG", Value: "1", Avail: Avail{From: phpver.PHP85}}
	second := &Constant{FQN: "FLAG", Value: "2", Avail: Avail{From: phpver.PHP84}, DeclarationFallback: `App\define`}
	ix.Add(&FileSymbols{Path: "constants.php", Constants: []*Constant{first, second}})
	if ix.Constant("FLAG", phpver.PHP53) != first {
		t.Fatal("permissive lookup must retain first eligible declaration")
	}
	if ix.Constant("FLAG", phpver.PHP84) != second {
		t.Fatal("available later declaration must win over unavailable first")
	}
	ix.Add(extract(t, "shadow.php", `<?php namespace App; function define() {}`))
	if ix.Constant("FLAG", phpver.PHP84) != first {
		t.Fatal("suppressed available declaration must reveal permissive fallback")
	}
}

// Repeated declarations of one name must not trigger repeated filtering of
// the same slice. This also leaves another file's candidates intact.
func TestDeclarationRemovalBounded(t *testing.T) {
	const count = 20000
	ix := New(nil)
	keeper := &Constant{FQN: "FLAG", Value: "1"}
	ix.Add(&FileSymbols{Path: "keep.php", Constants: []*Constant{keeper}, ClassAliases: [][2]string{{"alias", "Original"}}})
	repeated := repeatedDeclarations(count)
	ix.Add(repeated)
	start := time.Now()
	ix.Remove(repeated.Path)
	if elapsed := time.Since(start); elapsed > testbudget.Of(2*time.Second) {
		t.Fatalf("removing %d repeated declarations took %v", count, elapsed)
	}
	if ix.Constant("FLAG", 0) != keeper || ix.aliasOf("alias") != "Original" {
		t.Fatal("removal discarded another file's candidates")
	}
}

func repeatedDeclarations(count int) *FileSymbols {
	fs := &FileSymbols{Path: "repeated.php"}
	for range count {
		fs.Constants = append(fs.Constants, &Constant{FQN: "FLAG", Value: "2"})
		fs.ClassAliases = append(fs.ClassAliases, [2]string{"alias", "Other"})
	}
	return fs
}

func BenchmarkDeclarationRemoval(b *testing.B) {
	for _, count := range []int{1000, 10000, 20000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			fs := repeatedDeclarations(count)
			ix := New(nil)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				b.StopTimer()
				ix.Add(fs)
				b.StartTimer()
				ix.Remove(fs.Path)
			}
		})
	}
}
