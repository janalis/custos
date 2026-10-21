package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// A function of another file may bind a global and replace what file-level
// code proved of it when called; a PHP function that runs no project code
// cannot.
func TestGlobalsBoundInAnotherFile(t *testing.T) {
	for _, tc := range []struct {
		other, src string
		kept       bool
	}{
		{`<?php function refill() { global $u; $u = false; }`, `$u = parse_url($a); if ($u === false) { return; } refill(); probe($u);`, false},
		{`<?php function refill() { global $u; $u = false; }`, `$u = parse_url($a); if ($u === false) { return; } strlen('x'); probe($u);`, true},
		{`<?php function refill() { global $v; $v = false; }`, `$u = parse_url($a); if ($u === false) { return; } refill(); probe($u);`, true},
		{`<?php function refill($k) { $GLOBALS[$k] = false; }`, `$u = parse_url($a); if ($u === false) { return; } refill('u'); probe($u);`, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{Version: phpversion.Max})
			other := syntax.Parse("other.php", []byte(tc.other), syntax.Options{Version: phpversion.Max})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(f))
			ix.Add(index.Extract(other))
			e := New(f, infer.NewEnv(f, names.New(f), ix, phpversion.Max, infer.Options{}), nil, Options{})
			if got := e.Excludes(argument(callsNamed(f, "probe")[0]), "false"); got != tc.kept {
				t.Fatalf("non-false=%v", got)
			}
		})
	}
}
