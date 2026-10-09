package infer

import (
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

func TestRelativeMissingContext(t *testing.T) {
	e := &Env{Index: index.New(nil)}
	e.Index.Add(&index.FileSymbols{Path: "relative.php", Classes: []*index.Class{{FQN: "Trait", Kind: syntax.KindTrait}, {FQN: "Base"}}})
	for _, owner := range []string{"Missing", "Trait", "Base"} {
		if !e.bindMember(types.Of("parent"), owner, "", "Child").IsUnknown() {
			t.Fatal(owner)
		}
	}
	for _, owner := range []string{"Missing", "Trait"} {
		if !e.bindMember(types.Of("self"), owner, "", "Child").IsUnknown() {
			t.Fatal(owner)
		}
	}
	if got := e.bindMember(types.Of("self"), "Trait", "Base", "Child").String(); got != `\Base` {
		t.Fatal(got)
	}
	if got := e.bindMember(types.Int, "Missing", "", "Child").String(); got != "int" {
		t.Fatal(got)
	}
}

func BenchmarkBindMemberOrdinary(b *testing.B) {
	// No index is needed: ordinary types must avoid hierarchy lookups entirely.
	e := &Env{}
	b.ReportAllocs()
	for b.Loop() {
		e.bindMember(types.Int, "Base", "", "Child")
	}
}
