package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// TypeOfVarAt types a variable where a later call consumes a value computed
// from it, by the conditions guarding that call; ReachesUnchanged tells
// whether the variable still holds the value read earlier.
func TestTypeOfVarAt(t *testing.T) {
	src := `<?php
function f($x, $o) {
    $d = realpath($x); $p = $d . '/a';
    if ($d === false) { return; }
    u($p);
    $d = realpath($o); $p = $d . '/b';
    if ($o) { u($p); }
    $p = $d . '/c'; $d = 'x';
    u($p);
    $p = $d . '/e';
    while ($o) { if ($d) { u($p); } $d = realpath($o); }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var reads []*syntax.Variable
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if b, ok := n.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			reads = append(reads, b.Left.(*syntax.Variable))
		}
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "u" {
				v := reads[len(got)]
				got = append(got, env.TypeOfVarAt("d", c).String())
				if same := env.ReachesUnchanged(v, c); same != (len(got) < 3) {
					t.Errorf("call %d: unchanged %v", len(got), same)
				}
			}
		}
		return true
	})
	want := []string{"string", "false|string", "string", "string"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call %d: got %s want %s", i+1, got[i], want[i])
		}
	}
	if !env.TypeOfVarAt("this", reads[0]).IsUnknown() || !env.TypeOfVarAt("", reads[0]).IsUnknown() {
		t.Error("$this and unnamed variables are not typed")
	}
	if env.ReachesUnchanged(reads[1], reads[0]) {
		t.Error("an earlier place is not reached")
	}
}
