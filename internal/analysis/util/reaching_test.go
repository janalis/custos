package util

import (
	"strings"
	"testing"

	"custos/internal/syntax"
)

func TestReachingAssignments(t *testing.T) {
	src := `<?php
	function f1($v = 'p') { at($v); $v = 'after'; }
	function f2($v = 'p') { $v = 'a'; $v = 'b'; at($v); }
	function f3($c) { $v = 'a'; if ($c) { $v = 'b'; } at($v); }
	function f4($xs) { $v = 'a'; foreach ($xs as $x) { at($v); $v = 'loop'; } $v = 'end'; }
	function f5($xs) { $v = 'a'; foreach ($xs as $v) { at($v); } }
	function f6() { $v = 'a'; return; $v = 'dead'; }
	function f7() { $v = 'a'; $g = function () { $v = 'inner'; }; $h = function () use (&$v) { $v = 'ref'; }; at($v); }
	function f8($xs) { while ($xs) { $v = 'w1'; at($v); $v = 'w2'; } }
	function f9() { $v = 'a'; ($v = 'b'); at($v); }
	function f10($c) { switch ($c) { case 1: $v = 'a'; at($v); break; default: $v = 'b'; } }`
	f := parse(t, src)
	want := map[string]string{
		"f1":  "entry",
		"f2":  "'b'",
		"f3":  "'a' 'b'",
		"f4":  "'a' 'loop'",
		"f5":  "",
		"f6":  "",
		"f7":  "'a' 'ref'",
		"f8":  "'w1'",
		"f9":  "'b'",
		"f10": "'a'",
	}
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		c, ok := n.(*syntax.FuncCall)
		if !ok || CallLastName(c) != "at" {
			return true
		}
		scope := EnclosingFuncLike(c)
		fn := scope.(*syntax.Function).Name.Value
		defs, entry := ReachingAssignments(scope, c, "v")
		var parts []string
		if entry {
			parts = append(parts, "entry")
		}
		for _, d := range defs {
			s := d.Value.Span()
			parts = append(parts, src[s.Start:s.End])
		}
		got[fn] = strings.Join(parts, " ")
		return true
	})
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q, want %q", k, got[k], v)
		}
	}
}
