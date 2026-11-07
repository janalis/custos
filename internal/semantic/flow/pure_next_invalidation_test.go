package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureInheritedInstanceEditInvalidates(t *testing.T) {
	defs := `<?php class K{final function leaf($x){return $x+1;}}`
	caller := `<?php final class W extends K{function wrap($x){return $this->leaf($x);}}$w=new W;$cb=$w->wrap(...);probe($cb(2));`
	ix := stubsIndexForResource(t, defs, caller)
	opt := syntax.Options{Version: phpversion.Max}
	makeFile := func(path, source string) *syntax.File { return syntax.Parse(path, []byte(source), opt) }
	first := NewSnapshot(Extract(makeFile("definitions.php", defs), ix, phpversion.Max, nil, Options{}))
	file := makeFile("caller.php", caller)
	_, published := Summarize(file, resourceTypes(file, ix), first, Options{})
	old := NewLocal(file, resourceTypes(file, ix), first, Options{}, nil)
	at := argument(callsNamed(file, "probe")[0])
	scalar, known := knownScalar(old.Value(at))
	if !known || scalar.i != 3 {
		t.Fatal("inherited instance contract missing")
	}
	next := first.WithFile(Extract(makeFile("definitions.php", strings.Replace(defs, "+1", "+2", 1)), ix, phpversion.Max, nil, Options{}))
	if published.Current(first, next) {
		t.Fatal("helper edit kept a stale caller")
	}
	scalar, known = knownScalar(old.Rebase(next).Value(at))
	if !known || scalar.i != 4 {
		t.Fatal("helper edit retained old result")
	}
}
