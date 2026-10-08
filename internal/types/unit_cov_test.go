package types

import "testing"

// Statements once covered only by FuzzFromDoc seeds, now by unit tests:
// literal types, empty generic members, unknown arguments and shape keys
// in DocString, intersections against other alternatives.
func TestDocEdgesUnitOnly(t *testing.T) {
	if got := FromDoc("1.5", nil).String(); got != "float" {
		t.Errorf("1.5: %s", got)
	}
	if got := FromDoc("array<int, >", nil).String(); got != "array" {
		t.Errorf("array<int, >: %s", got)
	}
	if got := Of(`\Box`).WithTypeArgs(`\Box`, []Type{Unknown, Int}).DocString(); got != `\Box<mixed, int>` {
		t.Errorf("unknown argument: %s", got)
	}
	if Of("int").ShapeKeys() != nil {
		t.Error("ShapeKeys without a shape")
	}
	sh := Array.WithShape([]ShapeKey{{Name: "a", Type: Unknown}}, true)
	if got := sh.DocString(); got != "array{a:mixed}" {
		t.Errorf("unknown key type: %s", got)
	}
	if got := docAtom("int[]", &arrayInfo{}); got != "int[]" {
		t.Errorf("plain element: %s", got)
	}
	// An intersection does not survive a union with other intersections
	// or with another class.
	ab, cd := FromDoc(`\A&\B`, nil), FromDoc(`\C&\D`, nil)
	if Union(ab, cd).Intersection() != nil || Union(ab, Of(`\C`)).Intersection() != nil {
		t.Error("intersection kept across alternatives")
	}
	if got := Union(ab, Null).String(); got != `(\A&\B)|null` {
		t.Errorf("nullable intersection: %s", got)
	}
	if FromDoc(`(\A&\B)|(\C&\D)`, nil).Intersection() != nil {
		t.Error("DNF with two intersections")
	}
	if got := FromDoc(`\A&\B`, nil).Without(`\A`).Intersection(); got != nil {
		t.Errorf("one side removed: %v", got)
	}
	if !ab.Equal(FromDoc(`\B&\A`, nil)) || ab.Equal(Of(`\A`, `\B`)) {
		t.Error("Equal on intersections")
	}
	if got := FromDoc(`\A&int`, nil).Intersection(); got != nil {
		t.Errorf("non-class part: %v", got)
	}
}
