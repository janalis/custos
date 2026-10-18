package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

func TestDescendantMethods(t *testing.T) {
	src := `<?php
namespace N;
interface I { public function run(); }
class A implements I { public function run() {} public function name() {} public function only() {} }
class B extends A { public function name() {} }
class C extends B { public function run() {} }
class D extends A {}
class E extends D {}
interface K extends I {}
class G implements I, K {}
`
	f := syntax.Parse("x.php", []byte(src), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	for _, c := range []struct {
		fqn, name string
		want      bool
	}{
		{`N\A`, "name", true},
		{`N\A`, "run", true},
		{`\N\I`, "run", true},
		{`N\A`, "only", false},
		{`N\B`, "name", false},
		{`N\D`, "run", false},
		{`N\Missing`, "run", false},
		{`N\I`, "absent", false},
	} {
		if got := DescendantMethods(ix, c.fqn, 0).MayOverride(c.name); got != c.want {
			t.Errorf("DescendantMethods(%s, %s) = %v", c.fqn, c.name, got)
		}
	}
}

func TestDescendantTraitMethodsAndAliases(t *testing.T) {
	f := syntax.Parse("traits.php", []byte(`<?php
trait Body { function run(){} }
trait Nested { use Body {run as alias;} }
class Base { function run(){} function alias(){} function stable(){} }
class Child extends Base {use Nested;}
`), syntax.Options{})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	methods := DescendantMethods(ix, "Base", 0)
	if !methods.Complete || !methods.MayOverride("RUN") || !methods.MayOverride("alias") || methods.MayOverride("stable") {
		t.Fatalf("trait overrides=%+v", methods)
	}
	if exhausted := descendantMethods(ix, "Base", 0, 1); exhausted.Complete || !exhausted.MayOverride("stable") {
		t.Fatal("exhausted hierarchy query established absence")
	}
}

func TestDescendantMissingTraitPreventsAbsenceProof(t *testing.T) {
	f := syntax.Parse("missing.php", []byte(`<?php class Base{} class Child extends Base{use Missing;}`), syntax.Options{})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	methods := DescendantMethods(ix, "Base", 0)
	if methods.Complete || !methods.MayOverride("unknown") {
		t.Fatal("missing trait established absence of overrides")
	}
}

func TestInheritedTraitDoesNotCreateNewOverride(t *testing.T) {
	f := syntax.Parse("inherited.php", []byte(`<?php trait Body{function run(){}} class Base{use Body;} class Child extends Base{}`), syntax.Options{})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	methods := DescendantMethods(ix, "Base", 0)
	if !methods.Complete || methods.MayOverride("run") {
		t.Fatal("inherited trait counted as a replacement")
	}
}

func BenchmarkDescendantTraitMethods(b *testing.B) {
	var source strings.Builder
	source.WriteString(`<?php trait Body{function run(){}} class Base{function run(){}}`)
	for i := range 100 {
		fmt.Fprintf(&source, "class Child%d extends Base{use Body {run as alias%d;}}", i, i)
	}
	f := syntax.Parse("traits.php", []byte(source.String()), syntax.Options{})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	b.ReportAllocs()
	for b.Loop() {
		DescendantMethods(ix, "Base", 0)
	}
}

func TestDescendantMethodsCycle(t *testing.T) {
	f := syntax.Parse("x.php", []byte(`<?php class A extends B { function m() {} } class B extends A {}`), syntax.Options{})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !DescendantMethods(ix, "A", 0).MayOverride("m") {
		t.Error("a class in a cycle is its own descendant")
	}
}

func TestOverrideBudgetNeverEstablishesAbsence(t *testing.T) {
	f := syntax.Parse("bounded.php", []byte(`<?php trait Body{function run(){}} class Base{} class Child extends Base{use Body{run as renamed;}function direct(){}}`), syntax.Options{})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	for budget := range 32 {
		methods := descendantMethods(ix, "Base", 0, budget)
		if !methods.Complete {
			if !methods.MayOverride("unknown") {
				t.Fatalf("budget %d established absence", budget)
			}
		} else if !methods.MayOverride("run") || !methods.MayOverride("renamed") || !methods.MayOverride("direct") || methods.MayOverride("unknown") {
			t.Fatalf("budget %d completed with wrong contracts: %+v", budget, methods)
		}
	}
}

func TestDuplicateDescendantCannotProveAbsence(t *testing.T) {
	ix := index.New(nil)
	for i, source := range []string{`<?php class Base{} class Child extends Base{}`, `<?php class Child extends Base{function hidden(){}}`} {
		ix.Add(index.Extract(syntax.Parse(fmt.Sprintf("copy%d.php", i), []byte(source), syntax.Options{})))
	}
	methods := DescendantMethods(ix, "Base", 0)
	if methods.Complete || !methods.MayOverride("unknown") {
		t.Fatal("duplicate descendants established absence")
	}
}

func TestUnavailableDescendantAndTruncatedAncestors(t *testing.T) {
	f := syntax.Parse("versioned.php", []byte(`<?php class Base{} class Child extends Base{}`), syntax.Options{})
	symbols := index.Extract(f)
	symbols.Classes[1].Avail.From = phpversion.PHP80
	ix := index.New(nil)
	ix.Add(symbols)
	if methods := DescendantMethods(ix, "Base", phpversion.PHP74); methods.Complete || !methods.MayOverride("unknown") {
		t.Fatal("unavailable descendant established absence")
	}
	var source strings.Builder
	source.WriteString("<?php class Base{} ")
	var interfaces []string
	for i := range index.MaxAncestors + 1 {
		name := fmt.Sprintf("I%d", i)
		interfaces = append(interfaces, name)
		fmt.Fprintf(&source, "interface %s{} ", name)
	}
	fmt.Fprintf(&source, "class Child extends Base implements %s {}", strings.Join(interfaces, ","))
	ix = index.New(nil)
	ix.Add(index.Extract(syntax.Parse("wide.php", []byte(source.String()), syntax.Options{})))
	if methods := DescendantMethods(ix, "Base", 0); methods.Complete || !methods.MayOverride("unknown") {
		t.Fatal("truncated ancestor graph established absence")
	}
}
