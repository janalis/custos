package index

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEngineDeprecationAndAvailability(t *testing.T) {
	src := []byte(`<?php
 namespace Demo;
 use JetBrains\PhpStorm\Deprecated as Old;
 /** @since 8.0
  * @removed 8.5 */
 #[\Deprecated] function old() {}
 #[Old] function vendor() {}
 #[Other\Deprecated] function unrelated() {}
 /** @deprecated */ function documented() {}
 class C {
 #[\Deprecated] function old() {}
 #[\Deprecated] const OLD = 1;
 }
 enum E { #[\Deprecated] case Old; }
 `)
	f := syntax.Parse("test.php", src, syntax.Options{Version: phpversion.PHP84})
	user, stub := Extract(f), ExtractStubs(f)
	if !user.Functions[0].Deprecated || user.Functions[0].Avail != (Avail{}) || !stub.Functions[0].Avail.In(phpversion.PHP84) || stub.Functions[0].Avail.In(phpversion.PHP74) {
		t.Fatal("native attribute or stub availability")
	}
	if user.Functions[1].Deprecated || !stub.Functions[1].Deprecated || stub.Functions[2].Deprecated || !user.Functions[3].Deprecated {
		t.Fatal("attribute identity")
	}
	if !user.Classes[0].Methods["old"].Deprecated || !user.Classes[0].Consts["OLD"].Deprecated || !user.Classes[1].Consts["Old"].Deprecated {
		t.Fatal("member attributes")
	}
}

func TestStubSignatureVariants(t *testing.T) {
	f := syntax.Parse("stub.php", []byte(`<?php
 class C {
 #[PhpStormStubsElementAvailable(to: '7.4')] function query($old) {}
 #[PhpStormStubsElementAvailable(from: '8.0')] function query($new, $extra = 0) {}
 }
 function changed(#[PhpStormStubsElementAvailable(to: '7.4')] $old, #[PhpStormStubsElementAvailable(from: '8.0')] $new) {}
 `), syntax.Options{Version: phpversion.PHP84})
	ix := New(nil)
	ix.Add(ExtractStubs(f))
	for _, tc := range []struct {
		v    phpversion.Version
		name string
	}{{phpversion.PHP74, "old"}, {phpversion.PHP84, "new"}, {0, "new"}} {
		if got := ix.FindMethod("C", "query", tc.v); got == nil || got.Params[0].Name != tc.name {
			t.Fatalf("method %v: %+v", tc.v, got)
		}
		if got := ix.Function("changed", tc.v); len(got.Params) != 1 || got.Params[0].Name != tc.name || ix.Function("changed", tc.v) != got {
			t.Fatalf("parameter %v: %+v", tc.v, got)
		}
	}
	if ix.FindMethod("C", "query", phpversion.PHP85).Params[0].Name != "new" {
		t.Fatal("latest")
	}
	if (&Method{Avail: Avail{From: phpversion.PHP80}}).available(phpversion.PHP74) != nil {
		t.Fatal("unavailable")
	}
	if got := paramsAt([]Param{{Name: "a"}}, phpversion.PHP84); len(got) != 1 {
		t.Fatal(got)
	}
}

func TestTraitUnqualifiedAliasAfterSelection(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "traits.php", `<?php
 trait A { function run(): int {} }
 trait B { function run(): string {} }
 class C { use A, B { A::run insteadof B; run as private chosen; B::run as protected other; } }
 `))
	if m := ix.FindMethod("C", "chosen", 0); m == nil || m.Return != "int" || m.Visibility != Private {
		t.Fatalf("chosen %+v", m)
	}
	if m := ix.FindMethod("C", "other", 0); m == nil || m.Return != "string" || m.Visibility != Protected {
		t.Fatalf("other %+v", m)
	}
}

func TestVersionedInterfaces(t *testing.T) {
	c := &Class{FQN: "C", Interfaces: []string{"Old", "New"}, InterfacesVer: []VerInterfaces{{Until: phpversion.PHP80, Interfaces: []string{"Old"}}}}
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "a.php", Classes: []*Class{c}})
	old := ix.Class("C", phpversion.PHP74)
	if len(old.Interfaces) != 1 || ix.Class("C", phpversion.PHP74) != old || ix.ClassDecls("C", phpversion.PHP74)[0] != old {
		t.Fatal("old interfaces")
	}
	if ix.Class("C", 0) != c || ix.Class("C", phpversion.PHP84) != c {
		t.Fatal("latest interfaces")
	}
}

func BenchmarkVersionedSignature(b *testing.B) {
	f := &Function{Params: []Param{{Name: "old", Avail: Avail{To: phpversion.PHP74}}, {Name: "new", Avail: Avail{From: phpversion.PHP80}}}}
	f.at(phpversion.PHP74)
	b.ReportAllocs()
	for b.Loop() {
		f.at(phpversion.PHP74)
	}
}

func TestLatestVersionParameterNames(t *testing.T) {
	f := &Function{Params: []Param{{Name: "old", Avail: Avail{To: phpversion.PHP74}}, {Name: "new", Avail: Avail{From: phpversion.PHP80}}}}
	if p := f.at(0).Params; len(p) != 1 || p[0].Name != "new" {
		t.Fatal(p)
	}
}

func TestMethodAvailabilityIncludesReceiverAndParents(t *testing.T) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "versions.php", Classes: []*Class{
		{FQN: "Old", Methods: map[string]*Method{"run": {Name: "run", Class: "Old"}}},
		{FQN: "New", Parent: "Old", Avail: Avail{From: phpversion.PHP80}},
		{FQN: "Child", Parent: "New"},
	}})
	if ix.Class("New", phpversion.PHP74) == nil {
		t.Fatal("compatibility metadata lost")
	}
	if ix.FindMethod("New", "run", phpversion.PHP74) != nil || ix.FindMethod("Child", "run", phpversion.PHP74) != nil {
		t.Fatal("method from unavailable class")
	}
	if ix.FindMethod("New", "run", phpversion.PHP80) == nil || ix.FindMethod("New", "run", 0) == nil {
		t.Fatal("available inherited method lost")
	}
}
