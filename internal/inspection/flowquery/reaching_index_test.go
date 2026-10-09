package flowquery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestReachingAssignmentsIndexMatches checks the indexed implementation
// against the original per-query walk (kept below as the reference) for
// every variable use of every own fixture.
func TestReachingAssignmentsIndexMatches(t *testing.T) {
	paths, _ := filepath.Glob("../../../testdata/rules/*/*.php")
	if len(paths) == 0 {
		t.Skip("no fixtures")
	}
	extra := "<?php function f($xs) { $v = 1; $g = function () use (&$v, $w) { $h = function () use (&$v) { $v = 2; }; $v = 3; $w = 4; }; foreach ($xs as $k => $v) { at($v); } foreach ($xs as $v => $v) { at($v); } $v = 5; at($v); }"
	for _, p := range append(paths, "") {
		src := []byte(extra)
		if p != "" {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			src = b
		}
		f := syntax.Parse(p, src, syntax.Options{Version: phpversion.PHP84, Permissive: true})
		syntax.InspectFile(f, func(n syntax.Node) bool {
			v, ok := n.(*syntax.Variable)
			if !ok || v.NameExpr != nil || v.Name == "" {
				return true
			}
			scope := syntax.EnclosingFuncLike(v)
			if scope == nil {
				return true
			}
			gotDefs, gotEntry := ReachingAssignmentsIn(f, scope, v, v.Name)
			wantDefs, wantEntry := reachingAssignmentsRef(scope, v, v.Name)
			if gotEntry != wantEntry || !(len(gotDefs) == 0 && len(wantDefs) == 0) && !reflect.DeepEqual(gotDefs, wantDefs) {
				t.Fatalf("%s: $%s at %d: got %d defs entry=%v, want %d defs entry=%v", p, v.Name, v.Span().Start, len(gotDefs), gotEntry, len(wantDefs), wantEntry)
			}
			return true
		})
	}
}

func reachingAssignmentsRef(scope, at syntax.Node, name string) (defs []*syntax.Assign, entry bool) {
	_, body := astquery.ScopeParts(scope)
	if body == nil || name == "" {
		return nil, true
	}
	atStart := at.Span().Start
	var killers []syntax.Node // unconditional writes dominating at
	var all []*syntax.Assign
	var always []*syntax.Assign // from by-reference closures
	var walk func(n syntax.Node, byRefClosure bool)
	walk = func(n syntax.Node, byRefClosure bool) {
		syntax.Children(n, func(c syntax.Node) {
			switch x := c.(type) {
			case *syntax.Function, *syntax.ClassLike, *syntax.ArrowFunction:
				return
			case *syntax.Closure:
				for _, u := range x.Uses {
					if u.ByRef && u.Var != nil && u.Var.Name == name {
						walk(x, true)
						return
					}
				}
				return
			case *syntax.Assign:
				if t, ok := syntax.UnwrapParens(x.Var).(*syntax.Variable); ok && t.NameExpr == nil && t.Name == name {
					if x.Op.Kind == syntax.TEqual && !x.ByRef {
						if byRefClosure {
							always = append(always, x)
						} else {
							all = append(all, x)
						}
					}
					if !byRefClosure && reachDominates(x, at, atStart) {
						killers = append(killers, x)
					}
				}
			case *syntax.Foreach:
				if !byRefClosure && x.Body != nil && NodeContains(x.Body, at) && (reachIsVarNamed(x.Value, name) || reachIsVarNamed(x.Key, name)) {
					killers = append(killers, x)
				}
			}
			walk(c, byRefClosure)
		})
	}
	walk(body, false)

	lastKill := func(within syntax.Node) uint32 {
		var pos uint32
		found := false
		for _, k := range killers {
			if within != nil && !NodeContains(within, k) {
				continue
			}
			if p := reachKillPos(k); !found || p > pos {
				pos, found = p, true
			}
		}
		if !found {
			return 0
		}
		return pos
	}
	hasKill := func(within syntax.Node) bool {
		for _, k := range killers {
			if within == nil || NodeContains(within, k) {
				return true
			}
		}
		return false
	}

	entry = !hasKill(nil)
	kill := lastKill(nil)
	for _, a := range all {
		if !Reachable(a, scope) {
			continue
		}
		if a.Span().End <= atStart { // written before at
			if !hasKill(nil) || a.Span().Start >= kill {
				defs = append(defs, a)
			}
			continue
		}
		// written after at: only through an enclosing loop's back edge
		for l := a.Parent(); l != nil && l != scope; l = l.Parent() {
			if !reachIsLoop(l) || !NodeContains(l, at) {
				continue
			}
			if !hasKill(l) {
				defs = append(defs, a)
			}
			break
		}
	}
	return append(defs, always...), entry
}
