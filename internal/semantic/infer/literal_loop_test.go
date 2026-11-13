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

// A counter set from a literal and incremented in a loop is not read as
// its first value, even where an outer loop sets it again: comparing it
// with that value leaves the other branch a type, whatever order the reads
// are typed in.
func TestLiteralCounterInLoop(t *testing.T) {
	src := `<?php
class S {
    public $io;
    public function run(int $rulesCount, bool $more) {
      while ($more) {
        $pass = 1;
        for ($i = 0, $n = 0; $n < $rulesCount; $i++, $n++) {
            if ($i === $rulesCount) {
                if (1 === $pass) {
                    $this->io->writeError("again (pass #$pass)", false);
                } else {
                    $this->io->overwriteError("again (pass #$pass)", false, null);
                }
                $i = 0;
                $pass++;
            }
        }
      }
    }
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		x, ok := n.(syntax.Expr)
		if !ok {
			return true
		}
		t := env.TypeOf(x) // every expression, outer ones first
		if v, ok := n.(*syntax.Variable); ok && v.Name == "pass" {
			got = append(got, t.ShapeString())
		}
		return true
	})
	// The assignment target, the condition, both branches and the
	// incremented operand (which misses the float its own previous
	// iteration may store, but claims no value).
	want := []string{"?unknown", "float|int", "int", "float|int", "int"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("read %d: got %s want %s (all %v)", i, got[i], want[i], got)
		}
	}
}
