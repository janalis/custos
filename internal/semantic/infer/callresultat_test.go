package infer_test

import (
	"math"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
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
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
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
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
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
	resolved := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
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

// A caller-assumed type models results the declaration cannot tell apart:
// with null standing for a falsy 0, truthiness rules both out while a
// `!== false` test keeps the 0, also through a variable holding the result.
func TestCallResultAsAt(t *testing.T) {
	src := `<?php
function f($s) {
    if (stream_socket_enable_crypto($s, true)) { u(); }
    if (stream_socket_enable_crypto($s, true) !== false) { u(); }
    $r = stream_socket_enable_crypto($s, true); if ($r !== true) { return; } u();
    $r = stream_socket_enable_crypto($s, true); if ($r === false) { return; } u();
    $r = stream_socket_enable_crypto($s, true); if ($r === 1) { u(); }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	seed := types.Of("true", "false", "null")
	var last *syntax.FuncCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if c.Name.(*syntax.Name).Value == "u" {
				got = append(got, env.CallResultAsAt(last, c, seed).String())
			} else {
				last = c
			}
		}
		return true
	})
	want := []string{"true", "null|true", "true", "null|true", "?unknown"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// Only a plain variable that the assignment alone defines where the result is
// asked for carries the tests of the call; parentheses change nothing.
func TestCallResultStored(t *testing.T) {
	src := "<?php\n" +
		"function f($h) { $ok = (flock($h, LOCK_EX)); if (!$ok) { return; } u(); }\n" +
		"function g($h, $n) { $$n = flock($h, LOCK_EX); if (!$$n) { return; } u(); }\n" +
		"function h($h, $c) { $ok = flock($h, LOCK_EX);" + strings.Repeat(" if ($c > 1) { $ok = 1; }", 513) + " u(); }"
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	var last *syntax.FuncCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if c.Name.(*syntax.Name).Value == "u" {
				got = append(got, env.CallResultAt(last, c).String())
			} else {
				last = c
			}
		}
		return true
	})
	want := []string{"true", "bool", "bool"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// Known int values narrow by truthiness, comparisons with integer literals
// and relational tests: a fork result seeded as int<-1, max>.
func TestCallResultInts(t *testing.T) {
	src := `<?php
function f() {
    $pid = pcntl_fork(); if ($pid) { u(); }
    $pid = pcntl_fork(); if ($pid === -1) { exit(1); } if ($pid) { u(); }
    $pid = pcntl_fork(); if ($pid < 0) { exit(1); } u();
    $pid = pcntl_fork(); if (0 < $pid) { u(); }
    $pid = pcntl_fork(); if (!$pid) { u(); }
    $pid = pcntl_fork(); if ($pid == -1) { return; } elseif ($pid) { u(); } else { u(); }
    $pid = pcntl_fork(); if ($pid != 0) { u(); }
    $pid = pcntl_fork(); if ($pid !== 3) { u(); }
    $pid = pcntl_fork(); if ($pid === -2) { u(); }
    $pid = pcntl_fork(); if ($pid >= 1) { } else { u(); }
    $pid = pcntl_fork(); if ($pid <= 0) { u(); }
    if (pcntl_fork() > 5) { u(); }
    $pid = pcntl_fork(); $failed = -1; if ($failed === $pid) { return; } if ($bound < $pid) {} u();
    $pid = pcntl_fork(); if ($pid == 0) { u(); }
    $pid = pcntl_fork(); if ($pid > 9223372036854775807) { u(); }
    $pid = pcntl_fork(); if ($pid === count($GLOBALS)) { u(); }
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	seed := types.Ints(types.IntRange{Lo: -1, Hi: math.MaxInt64})
	var last *syntax.FuncCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			switch c.Name.(*syntax.Name).Value {
			case "u":
				got = append(got, env.CallResultAsAt(last, c, seed).ShapeString())
			case "pcntl_fork":
				last = c
			}
		}
		return true
	})
	want := []string{
		"int<-1, -1>|int<1, max>", "int<1, max>", "int<0, max>", "int<1, max>", "int<0, 0>",
		"int<1, max>", "int<0, 0>", "int<-1, -1>|int<1, max>", "int<-1, 2>|int<4, max>", "?unknown",
		"int<-1, 0>", "int<-1, 0>", "int<6, max>", "int<0, max>", "int<0, 0>", "?unknown", "int<-1, max>",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A copy of a variable holding the result carries its tests too, while each
// link still holds what it copied.
func TestCallResultCopies(t *testing.T) {
	src := "<?php\n" +
		"function a($h) { $s = flock($h, LOCK_EX); $t = $s; if (!$t) { return; } u(); }\n" +
		"function b($h) { $s = flock($h, LOCK_EX); $t = $s; $u = $t; if ($u !== true) { return; } u(); }\n" +
		"function c($h, $o) { $s = flock($h, LOCK_EX); $s = $o; $t = $s; if (!$t) { return; } u(); }\n" +
		"function d($h, $o) { $s = flock($h, LOCK_EX); $t = $s; $t = $o; if (!$t) { return; } u(); }\n" +
		"function e($h) { $a = flock($h, LOCK_EX); $b = $a; $c = $b; $d = $c; $e = $d; $f = $e; if (!$f) { return; } u(); }\n" +
		"function g($h) { $a = flock($h, LOCK_EX); $f = $a; $b = $a; $c = $b; $d = $c; $e = $d; $f = $e; if (!$f) { return; } u(); }\n"
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	var last *syntax.FuncCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if c.Name.(*syntax.Name).Value == "u" {
				got = append(got, env.CallResultAt(last, c).String())
			} else {
				last = c
			}
		}
		return true
	})
	want := []string{"true", "true", "bool", "bool", "bool", "bool"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A local past the definition budget holds no known integer.
func TestKnownIntBudget(t *testing.T) {
	src := "<?php\nfunction f($c) { $pid = pcntl_fork(); $failed = -1;" + strings.Repeat(" if ($c > 1) { $failed = -1; }", 513) + " if ($pid === $failed) { return; } u(); }"
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	var fork, use *syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if c.Name.(*syntax.Name).Value == "u" {
				use = c
			} else {
				fork = c
			}
		}
		return true
	})
	if got := env.CallResultAsAt(fork, use, types.Ints(types.IntRange{Lo: -1, Hi: math.MaxInt64})).ShapeString(); got != "int<-1, max>" {
		t.Fatalf("got %s", got)
	}
}

// A method call's result is narrowed like a function's: directly, silenced,
// assigned in place or through the local holding it.
func TestCallResultAtMethod(t *testing.T) {
	src := `<?php
function f(SplQueue $q) {
    if (!$q->isEmpty()) { u(); }
    if ($q->isEmpty()) { return; } u();
    if (!$e = @$q->isEmpty()) { u(); }
    $e = $q->isEmpty(); if ($e === false) { u(); }
    if ($q->isEmpty()) { u(); }
    $q->isEmpty(); u();
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	var last *syntax.MethodCall
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch c := n.(type) {
		case *syntax.MethodCall:
			last = c
		case *syntax.FuncCall:
			got = append(got, env.CallResultAt(last, c).String())
		}
		return true
	})
	want := []string{"false", "false", "false", "false", "true", "bool"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}

// A comparison's result is narrowed like a call's: tested directly, through
// `!`, compared with a boolean or kept in a local.
func TestCallResultAtComparison(t *testing.T) {
	src := `<?php
function f($a, $b) {
    if ($a === $b) { u(); }
    if ($a !== $b) { return; } u();
    $same = $a == $b; if (!$same) { u(); }
    if (($a === $b) === false) { u(); }
    $a === $b && u();
    if ($a === $b) {} u();
}`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{}).Native()
	var last *syntax.Binary
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch c := n.(type) {
		case *syntax.Binary:
			switch c.Op.Kind {
			case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual:
				if _, inner := c.Left.(*syntax.Paren); !inner {
					last = c
				}
			}
		case *syntax.FuncCall:
			got = append(got, env.CallResultAsAt(last, c, types.Bool).String())
		}
		return true
	})
	want := []string{"true", "false", "false", "false", "true", "bool"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("use %d: got %s want %s", i+1, got[i], want[i])
		}
	}
}
