package types

import (
	"strings"
	"testing"
)

func TestBindRelative(t *testing.T) {
	for _, src := range []string{
		"self|parent|static|null", "self[][]|parent[]|static[]", "Collection<self, callable(): parent>",
		"array{same: self, next: static, optional?: parent}", "array{same: self, plain: int}[]",
		"non-empty-array{v: self}", "array{plain: int, v: self, next: string}",
		"Collection<int, self, string>", "callable(): Collection<static>", "A&B|null", "array{plain: int}[]", "int", "", "parent",
	} {
		t.Run(src, func(t *testing.T) {
			in := FromDoc(src, nil)
			before := in.DocString()
			bound := in.BindRelative("Base", "Root", "Child")
			want := FromDoc(strings.NewReplacer("self", `\Base`, "parent", `\Root`, "static", `\Child`).Replace(before), nil).DocString()
			if in.IsUnknown() {
				want = before
			}
			if got := bound.DocString(); got != want {
				t.Errorf("got %s want %s", got, want)
			}
			if in.DocString() != before {
				t.Fatal("mutated original")
			}
			if in.HasRelative() != (strings.Contains(src, "self") || strings.Contains(src, "parent") || strings.Contains(src, "static")) {
				t.Errorf("HasRelative %s", src)
			}
		})
	}
	for _, relative := range []string{"self", "parent", "static", "parent[][]"} {
		if !Of(relative).BindRelative("", "", "").IsUnknown() {
			t.Error(relative)
		}
	}
	// Generic arguments keyed by a replaced atom move to its resolved atom.
	g := Of("self", "string").WithTypeArgs("self", []Type{Of("parent"), Int}).WithTypeArgs("string", []Type{Int})
	if got := g.BindRelative("Base", "Root", "Child").TypeArgs(`\Base`); len(got) != 2 || got[0].String() != `\Root` {
		t.Fatalf("arguments: %v", got)
	}
}

func TestBindRelativeBounds(t *testing.T) {
	deep := Of("self")
	for range 70 {
		deep = Of(`\Box`).WithTypeArgs(`\Box`, []Type{deep})
	}
	if !deep.HasRelative() {
		t.Fatal("depth should be conservative")
	}
	bound := deep.BindRelative("Base", "Root", "Child")
	for range 64 {
		args := bound.TypeArgs(`\Box`)
		if len(args) != 1 {
			t.Fatal("lost outer generic contract")
		}
		bound = args[0]
	}
	if !bound.IsUnknown() {
		t.Fatal("over-depth contract must become unknown")
	}
	arr := &arrayInfo{keys: []ShapeKey{{Name: "v", Type: Of("self")}}}
	for range 70 {
		arr = &arrayInfo{elem: arr}
	}
	typ := Type{atoms: []string{"array"}, arr: arr}
	if !typ.HasRelative() {
		t.Fatal("array depth should be conservative")
	}
	boundArray := typ.BindRelative("Base", "Root", "Child").arr
	for range 63 {
		if boundArray == nil {
			t.Fatal("lost outer array facts")
		}
		boundArray = boundArray.elem
	}
	if boundArray != nil {
		t.Fatal("over-depth array facts must be dropped")
	}
	args := make([]Type, 4100)
	for i := range args {
		args[i] = Int
	}
	args[len(args)-1] = Of("self")
	wide := Of(`\Box`).WithTypeArgs(`\Box`, args)
	if !wide.HasRelative() {
		t.Fatal("work bound should be conservative")
	}
	wideBound := wide.BindRelative("Base", "Root", "Child")
	if !wideBound.IsUnknown() || args[len(args)-1].String() != "self" {
		t.Fatal("over-budget contract must be unknown without mutation")
	}
	keys := make([]ShapeKey, 4100)
	for i := range keys {
		keys[i].Type = Int
	}
	keys[len(keys)-1].Type = Of("self")
	typ = Type{atoms: []string{"array"}, arr: &arrayInfo{keys: keys}}
	if !typ.HasRelative() {
		t.Fatal("shape work bound should be conservative")
	}
	shapeBound := typ.BindRelative("Base", "Root", "Child").arr
	if shapeBound != nil || keys[len(keys)-1].Type.String() != "self" {
		t.Fatal("over-budget shape member must become unknown without mutation")
	}
}

func BenchmarkBindRelative(b *testing.B) {
	for _, src := range []string{"int", "self", "array{owner: self, next: callable(): static}"} {
		b.Run(src, func(b *testing.B) {
			t := FromDoc(src, nil)
			b.ReportAllocs()
			for b.Loop() {
				t.BindRelative("Base", "Root", "Child")
			}
		})
	}
}

func BenchmarkHasRelative(b *testing.B) {
	t := Int
	b.ReportAllocs()
	for b.Loop() {
		t.HasRelative()
	}
}
