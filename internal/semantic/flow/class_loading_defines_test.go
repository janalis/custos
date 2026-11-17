package flow

import "testing"

// Looking for a computed define reads syntax and names only: it infers no
// receiver of a method call, so no inference limit is met on the way and
// what it caches on the tree holds for any later environment.
func TestComputedDefinesInfersNothing(t *testing.T) {
	src := `function cycle() { return cycle(); } cycle()->next(); define($name, 1);`
	e, _ := environment(t, src, nil)
	if !e.computedDefines() {
		t.Fatal("computed define missed")
	}
	if gaps := e.types.Gaps(); len(gaps) != 0 {
		t.Fatalf("gaps %v", gaps)
	}
	// The receiver alone meets the recursion.
	e, f := environment(t, src, nil)
	e.types.TypeOf(callsNamed(f, "cycle")[1])
	if len(e.types.Gaps()) == 0 {
		t.Fatal("the receiver met no inference limit")
	}
}
