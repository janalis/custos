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

// CallResultAt narrows the result of one call by the conditions on that very
// call guarding a later place: a test of the same text on another call, or of
// a boolean that no longer holds it, proves nothing.
func TestCallResultAt(t *testing.T) {
	src := `<?php
function f($h, $c) {
    if (flock($h, LOCK_EX)) { u(); }
    flock($h, LOCK_EX) or exit(1); u();
    if (!@flock($h, LOCK_EX)) { return; } u();
    if (($ok = flock($h, LOCK_EX)) === false) { return; } u();
    $ok = @flock($h, LOCK_EX); if (!$ok) { return; } u();
    $ok = flock($h, LOCK_EX); $ok = $c; if (!$ok) { return; } u();
    flock($h, LOCK_EX); if (flock($h, LOCK_SH)) { u(); }
    if (flock($h, LOCK_EX)) { echo 1; } u();
    if (flock($h, LOCK_EX) !== true) { return; } u();
    $r = flock($h, LOCK_EX) ? u() : 0;
    if ($c && flock($h, LOCK_EX)) { echo 1; } elseif ($c) { u(); }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var last *syntax.FuncCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		c, ok := n.(*syntax.FuncCall)
		if !ok {
			return true
		}
		switch c.Name.(*syntax.Name).Value {
		case "flock": // the target takes LOCK_EX, a decoy LOCK_SH
			if c.Args.Args[1].(*syntax.Arg).Value.(*syntax.ConstFetch).Name.Value == "LOCK_EX" {
				last = c
			}
		case "u":
			got = append(got, env.CallResultAt(last, c).String())
		}
		return true
	})
	want := []string{"true", "true", "true", "true", "true", "bool", "bool", "bool", "true", "true", "bool|false"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A place before the call, or in another scope, knows nothing of its result.
func TestCallResultAtElsewhere(t *testing.T) {
	src := `<?php
function f($h) { u(); if (flock($h, LOCK_EX)) { $g = function () { u(); }; } }`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	var lock *syntax.FuncCall
	var uses []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if c.Name.(*syntax.Name).Value == "flock" {
				lock = c
			} else {
				uses = append(uses, c)
			}
		}
		return true
	})
	for i, u := range uses {
		if r := env.CallResultAt(lock, u); !r.IsUnknown() {
			t.Errorf("use %d: got %s", i+1, r)
		}
	}
}

// Identical to a constant of a type its declaration rules out, a value
// reaches no code in the native environment; the resolved one keeps the
// constant's type, as documented types may be wrong.
func TestDisjointIdentity(t *testing.T) {
	src := `<?php
function f(bool $b) { if ($b === 1) { u($b); } }`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	resolved := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84)
	for _, tc := range []struct {
		env  *infer.Env
		want string
	}{{resolved.Native(), "?unknown"}, {resolved, "int"}} {
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok && c.Name.(*syntax.Name).Value == "u" {
				if got := tc.env.TypeOf(c.Args.Args[0].(*syntax.Arg).Value).String(); got != tc.want {
					t.Errorf("got %s want %s", got, tc.want)
				}
			}
			return true
		})
	}
}
