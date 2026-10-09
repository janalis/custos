package infer

import (
	"testing"

	"custos/internal/semantic/types"
)

// A stub doc type sharing no member with the declared type refines
// nothing: the declared type stays.
func TestRefiningDisjoint(t *testing.T) {
	if got := refining(types.Of("int"), types.Of("string")).String(); got != "int" {
		t.Errorf("got %s", got)
	}
	if got := builtinMemberType("bool", "").String(); got != "bool" {
		t.Errorf("no doc: got %s", got)
	}
}
