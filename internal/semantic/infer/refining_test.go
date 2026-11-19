package infer

import (
	"testing"

	"github.com/janalis/custos/internal/semantic/types"
)

// A stub doc type sharing no member with the declared type refines
// nothing: the declared type stays.
func TestRefiningDisjoint(t *testing.T) {
	if got := refining(types.Of("int"), types.Of("string")).String(); got != "int" {
		t.Errorf("got %s", got)
	}
	env, _ := declEnv(t)
	if got := env.builtinMemberType("bool", "").String(); got != "bool" {
		t.Errorf("no doc: got %s", got)
	}
}
