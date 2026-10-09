package infer

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// guards applies early-exit guards among the statements preceding child
// in stmts (the statements of owner; nil: the file). Only the statements
// that can matter for name are visited (see guardIndex).
func (e *Env) guards(use syntax.Expr, scope syntax.Node, t types.Type, owner syntax.Node, stmts []syntax.Stmt, child syntax.Node, name string) types.Type {
	gi := e.guardIndexOf(owner, stmts)
	end, ok := gi.pos[child]
	if !ok {
		return t
	}
	// Only the statements after the last one assigning name (or its array)
	// matter: an assignment resets the guards seen before it. No statement
	// scanned below assigns name (or, for an element key, its array or an
	// element of it): gi.resets records every such top-level assignment.
	from := gi.lastReset(name, end)
	if from > 0 && isPropKey(name) && !isDimKey(name) {
		t = e.afterPropertyWrite(use, scope, t, stmts[from-1], name)
	}
	cands, ok := gi.candidates(name, from, end)
	if !ok {
		return types.Unknown // hostile statement lists: see maxGuardScan
	}
	for _, i := range cands {
		s := stmts[i]
		if es, ok := s.(*syntax.ExprStmt); ok {
			// `Assert::string($x);` (@phpstan-assert on the callee).
			if ca := e.assertsOf(es.Expr); ca != nil {
				t, _ = e.applyAsserts(ca, index.AssertAlways, name, t, false)
			}
			// `assert($x instanceof Foo);`: the code after it relies on
			// the condition (with assertions disabled it is not checked,
			// but then the author's claim is all there is).
			if c := e.assertCall(es.Expr); c != nil {
				t = e.condAt(use, scope, t, c, name, true, es.Span().End)
			}
			continue
		}
		if g, ok := s.(*syntax.If); ok && g.Else == nil && len(g.ElseIfs) == 0 {
			if terminates(g.Body) {
				t = e.condAt(use, scope, t, g.Cond, name, false, g.Span().End)
			} else if val := e.overwrites(g.Body, name); val != nil && !(isDimKey(name) && e.dimBroken(scope, name, g.Span().End, use)) &&
				!(isPropKey(name) && e.nonEmptyBroken(scope, name, g.Span().End, use)) {
				// `if (false === $x) { $x = $default; }`: past the if, the
				// condition no longer holds (unless the new value matches it).
				vt := e.TypeOf(val)
				if !vt.IsUnknown() && e.applyCond(vt, g.Cond, name, false).String() == vt.String() {
					t = e.applyCond(t, g.Cond, name, false)
				}
			}
		} else if ok && len(g.ElseIfs) < maxBranchScan {
			t = e.chainJoin(use, scope, t, g, name)
		}
	}
	return t
}

// chainJoin types name past an if/elseif(/else) chain g preceding use: the
// union, over the paths through the chain, of the value each path leaves —
// the value a branch assigns last, or the incoming type narrowed by the
// conditions that path saw (earlier ones false, its own true; all false
// without else). Branches that always leave contribute nothing. t is kept
// when a branch writes name otherwise, or a written value is unknown.
// (`if ($c instanceof Lang) {…} elseif ($c instanceof Code) { $c = 'x'; }`
// leaves a string|Code $c a string.)
func (e *Env) chainJoin(use syntax.Expr, scope syntax.Node, t types.Type, g *syntax.If, name string) types.Type {
	from := g.Span().End
	var parts []types.Type
	var prefix []syntax.Expr // conditions found false so far
	path := func(body syntax.Stmt, last syntax.Expr) bool {
		if terminates(body) {
			return true
		}
		if val := e.overwrites(body, name); val != nil {
			vt := e.TypeOf(val)
			parts = append(parts, vt)
			return !vt.IsUnknown()
		}
		sp := body.Span()
		for _, m := range e.mutations(scope, name) {
			if m.span.Start >= sp.Start && m.span.Start < sp.End {
				return false
			}
		}
		pt := t
		for _, c := range prefix {
			pt = e.condAt(use, scope, pt, c, name, false, from)
		}
		if last != nil {
			pt = e.condAt(use, scope, pt, last, name, true, from)
		}
		if !pt.IsUnknown() { // unknown: an impossible path
			parts = append(parts, pt)
		}
		return true
	}
	prefix = make([]syntax.Expr, 0, 1+len(g.ElseIfs))
	parts = make([]types.Type, 0, 2+len(g.ElseIfs))
	for i := -1; i < len(g.ElseIfs); i++ {
		c, body := g.Cond, g.Body
		if i >= 0 {
			c, body = g.ElseIfs[i].Cond, g.ElseIfs[i].Body
		}
		if !path(body, c) {
			return t
		}
		prefix = append(prefix, c)
	}
	var tail syntax.Stmt = emptyBlock // all conditions false, nothing runs
	if g.Else != nil {
		tail = g.Else.Body
	}
	if !path(tail, nil) || len(parts) == 0 {
		return t
	}
	return types.Union(parts...)
}

// emptyBlock stands for the missing else of a chain (see chainJoin).
var emptyBlock = &syntax.Block{}

// afterPropertyWrite drops null from the type t of `$this->prop` read use
// when s, a statement preceding it in an enclosing block (so it ran on
// every path to use), is `$this->prop = v;` or `$this->prop ??= v;` with a
// non-null v, and nothing that may reset the property (an assignment, a
// reference, a non-builtin call) lies between them (nonEmptyBroken).
func (e *Env) afterPropertyWrite(use syntax.Expr, scope syntax.Node, t types.Type, s syntax.Stmt, name string) types.Type {
	a := s.(*syntax.ExprStmt).Expr.(*syntax.Assign) // gi.resets lists only such statements
	if narrowKey(a.Var) != name {
		return t // `$v = …` resetting `$v->p`: the property's value is not known
	}
	if a.ByRef || (a.Op.Kind != syntax.TEqual && a.Op.Kind != syntax.TCoalesceEqual) {
		return t
	}
	vt := e.TypeOf(a.Value)
	if e.nonEmptyBroken(scope, name, a.Span().End, use) {
		return t
	}
	if a.Op.Kind == syntax.TEqual && (vt.IsUnknown() || vt.Has("mixed")) && !e.nativeProp(a.Var) {
		// The value just stored is not known and nothing enforces the
		// property's documented type: it is no more than a hint.
		return types.Unknown
	}
	if !t.IsNullable() || vt.IsUnknown() || vt.IsNullable() || vt.Has("mixed") {
		return t
	}
	if nt := t.Without("null"); len(nt.Atoms()) > 0 {
		return nt
	}
	return t
}

// guardIndex lists, for one statement list, the statements guards may use
// for each narrowing key: expression statements (assignments, assertion
// calls) and else-less ifs, by the keys (narrowKey) occurring in them.
// guards used to test every preceding statement for each read, which is
// quadratic in long statement lists.
type guardIndex struct {
	pos  map[syntax.Node]int // statement -> position
	keys map[string][]int    // key -> positions, ascending
	// resets lists, per key, the positions of the statements assigning it
	// (assigns) and, under baseKey(k), those assigning an element of it
	// or it.
	resets map[string][]int
}

// maxGuardScan caps the statements guards examines for one read (those
// after the last assignment of the name that mention it). Beyond it the
// read is unknown: real code stays far below; hostile code with thousands
// of guards or calls on one variable would make each read linear.
const maxGuardScan = 512

func baseKey(k string) string { return "\x02" + k }

// lastReset returns the position just after the last statement before end
// that assigns name (for an element key, also its array): guards before
// it no longer apply (0 when none).
func (gi *guardIndex) lastReset(name string, end int) int {
	from := 0
	last := func(l []int) {
		if i, _ := slices.BinarySearch(l, end); i > 0 && l[i-1]+1 > from {
			from = l[i-1] + 1
		}
	}
	last(gi.resets[name])
	if b := propBase(name); b != "" {
		last(gi.resets[b])
	}
	if isDimKey(name) {
		base, _ := splitDimKey(name)
		last(gi.resets[baseKey(base)])
	}
	return from
}

// aliasGuardKey lists the else-less ifs whose condition is a bare variable
// (possibly negated or combined): a boolean alias (see aliasCond).
const aliasGuardKey = "\x04alias"

// bareVarCond reports a condition made of plain variables under `!`, `&&`
// and `||` (at least one).
func bareVarCond(c syntax.Expr) bool {
	switch n := syntax.UnwrapParens(c).(type) {
	case *syntax.Variable:
		return n.NameExpr == nil
	case *syntax.Unary:
		return n.Op.Kind == syntax.TExclaim && bareVarCond(n.Expr)
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return bareVarCond(n.Left) || bareVarCond(n.Right)
		}
	}
	return false
}

// thisCallKey lists the statements with a call on `$this` (assertions on
// `$this->prop` targets).
const thisCallKey = "this->*"

func (e *Env) guardIndexOf(owner syntax.Node, stmts []syntax.Stmt) *guardIndex {
	if gi, ok := e.guardIdx[owner]; ok {
		return gi
	}
	gi := &guardIndex{pos: make(map[syntax.Node]int, len(stmts)), keys: map[string][]int{}, resets: map[string][]int{}}
	for i, st := range stmts {
		gi.pos[st] = i
		seen := map[string]bool{}
		add := func(k string) {
			if k != "" && !seen[k] {
				seen[k] = true
				gi.keys[k] = append(gi.keys[k], i)
			}
		}
		collect := func(x syntax.Node) {
			syntax.Inspect(x, func(n syntax.Node) bool {
				switch n := n.(type) {
				case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
					return false
				case *syntax.Variable, *syntax.PropertyFetch, *syntax.ArrayDimFetch, *syntax.StaticPropertyFetch:
					add(narrowKey(n.(syntax.Expr)))
				case *syntax.MethodCall:
					if narrowKey(syntax.UnwrapParens(n.Var)) == "this" {
						add(thisCallKey)
					}
				case *syntax.StaticCall:
					if nm, ok := n.Class.(*syntax.Name); ok {
						switch strings.ToLower(nm.Value) {
						case "self", "static", "parent":
							add(thisCallKey)
						}
					}
				}
				return true
			})
		}
		switch st := st.(type) {
		case *syntax.ExprStmt:
			collect(st.Expr)
			if a, ok := st.Expr.(*syntax.Assign); ok {
				// The writes assigns matches, and under baseKey those of
				// the array or one of its elements.
				if k := narrowKey(a.Var); k != "" {
					gi.resets[k] = append(gi.resets[k], i)
				}
				x := a.Var
				for {
					d, ok := x.(*syntax.ArrayDimFetch)
					if !ok {
						break
					}
					x = d.Var
				}
				if k := narrowKey(x); k != "" {
					bk := baseKey(k)
					if l := gi.resets[bk]; len(l) == 0 || l[len(l)-1] != i {
						gi.resets[bk] = append(gi.resets[bk], i)
					}
				}
			}
		case *syntax.If:
			if st.Else == nil && len(st.ElseIfs) == 0 {
				collect(st.Cond)
				if bareVarCond(st.Cond) {
					add(aliasGuardKey) // may stand for a condition on another variable
				}
				last := st.Body
				if b, ok := last.(*syntax.Block); ok && len(b.Stmts) > 0 {
					last = b.Stmts[len(b.Stmts)-1]
				}
				if es, ok := last.(*syntax.ExprStmt); ok {
					if a, ok := es.Expr.(*syntax.Assign); ok {
						add(narrowKey(a.Var))
					}
				}
			} else if len(st.ElseIfs) < maxBranchScan {
				// Chains: their conditions narrow the join (chainJoin).
				collect(st.Cond)
				for _, ei := range st.ElseIfs {
					collect(ei.Cond)
				}
			}
		}
	}
	if e.guardIdx == nil {
		e.guardIdx = map[syntax.Node]*guardIndex{}
	}
	e.guardIdx[owner] = gi
	return gi
}

// candidates returns the positions in [from, end) of the statements that
// may concern name, ascending: those mentioning it, its array for an
// element key, calls on `$this` for a `$this->prop` key, boolean-alias
// guards for a variable. ok is false beyond maxGuardScan of them (each
// list is cut by binary search before merging, so a read costs no more
// than the statements it may scan).
func (gi *guardIndex) candidates(name string, from, end int) ([]int, bool) {
	var buf [3][]int
	lists := append(buf[:0], gi.keys[name])
	if isDimKey(name) {
		base, _ := splitDimKey(name)
		lists = append(lists, gi.keys[base])
		if strings.HasPrefix(base, "this->") {
			lists = append(lists, gi.keys[thisCallKey])
		}
	} else if strings.HasPrefix(name, "this->") {
		lists = append(lists, gi.keys[thisCallKey])
	} else if !isPropKey(name) {
		lists = append(lists, gi.keys[aliasGuardKey])
	}
	total, n := 0, 0
	var only []int
	for i, l := range lists {
		lo, _ := slices.BinarySearch(l, from)
		hi, _ := slices.BinarySearch(l, end)
		lists[i] = l[lo:hi]
		if hi > lo {
			total += hi - lo
			n++
			only = lists[i]
		}
	}
	if n <= 1 {
		return only, len(only) <= maxGuardScan
	}
	if total > len(lists)*maxGuardScan {
		return nil, false // more than maxGuardScan even without duplicates
	}
	out := make([]int, 0, total)
	for _, l := range lists {
		out = append(out, l...)
	}
	slices.Sort(out)
	out = slices.Compact(out)
	return out, len(out) <= maxGuardScan
}
