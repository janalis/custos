package infer

import (
	"cmp"
	"slices"

	"custos/internal/syntax"
)

// reaching returns the definitions among defs (one variable's, in source
// order) that reach use: fwd are the definitions before use that no later
// kill hides (an unconditional assignment in an enclosing block, an
// if/else chain assigning in every branch, an inline @var), back the
// definitions later in the loop around use that reach it through the back
// edge (only when no kill inside that loop precedes use). from is the
// position of the earliest forward definition. Element writes (defs with a
// non-nil w) take part like definitions but never kill.
func (e *Env) reaching(defs []varDef, use, scope syntax.Node) (fwd, back []varDef, from uint32) {
	pos := use.Span().Start
	var lastKill uint32 // position of the last definition that reset fwd
	for i := 0; i < len(defs); i++ {
		d := defs[i]
		if d.pos > pos {
			break
		}
		if d.end > pos {
			// The use is inside the defining assignment itself (e.g. the
			// right-hand side of `$x = $x ?? null`): it sees earlier definitions only.
			continue
		}
		if d.barrier {
			if d.kill.Start <= pos && pos < d.kill.End {
				fwd = fwd[:0]
				lastKill = d.pos
			}
			continue
		}
		if d.doc && len(fwd) > 0 && !(i+1 < len(defs) && !defs[i+1].doc && defs[i+1].w == nil && defs[i+1].pos <= pos &&
			!e.semicolonBetween(d.docEnd, defs[i+1].pos)) && !e.docFits(d, fwd) {
			// A standalone `/** @var T $x */` contradicting what $x holds
			// (`@var Widget $class` on a class-name string) is a hint about
			// something else: the definitions keep their type.
			continue
		}
		if d.doc || (d.kill.Len() > 0 && d.kill.Start <= pos && pos < d.kill.End) {
			// An inline annotation states the type from here on; so does an
			// unconditional assignment earlier in an enclosing block.
			fwd = fwd[:0]
			lastKill = d.pos
		}
		if d.w == nil && !d.doc && d.kill.Len() > 0 && len(fwd) > 0 {
			// d runs unconditionally in its block: earlier definitions made
			// inside that block are overwritten for every later use, even
			// one after the block (`if (…) { $x = a(); $x = b($x); } use($x)`).
			kept := fwd[:0]
			for _, f := range fwd {
				if f.pos >= d.kill.Start && f.pos < d.pos && max(f.pos, f.end) <= d.kill.End {
					continue
				}
				kept = append(kept, f)
			}
			fwd = kept
		}
		if len(fwd) == 0 {
			from = max(d.pos, d.end)
		}
		fwd = append(fwd, d)
		// `/** @var T $x */ $x = ...;` — the annotation replaces the
		// assigned value's type. Only the statement the annotation is
		// attached to: no `;` between the comment and that definition.
		if d.doc && i+1 < len(defs) && !defs[i+1].doc && defs[i+1].w == nil && defs[i+1].pos <= pos && !e.semicolonBetween(d.docEnd, defs[i+1].pos) {
			i++
			if len(fwd) == 1 {
				// The annotated assignment is where the value is defined:
				// it is no mutation of the annotated type (non-empty facts).
				from = max(defs[i].pos, defs[i].end)
			}
		}
	}
	fwd = dropOtherCases(fwd, use, scope)
	fwd = dropExclusive(fwd, use, scope)
	fwd, via, loops := e.dropExited(fwd, use, scope, false)
	// Definitions followed by continue come back through their loop's
	// head, unless a definition after that head hides them.
	for i, d := range via {
		if lastKill < loops[i] {
			back = append(back, d)
		}
	}
	if len(fwd) == 0 && len(back) == 0 {
		return nil, nil, 0
	}
	// Inside a loop, definitions later in the loop reach the use through
	// the back edge (`$prev = null; foreach (…) { f($prev); $prev = $x; }`).
	// Not when an unconditional definition inside the loop precedes the
	// use: every path from the back edge passes through it.
	if loop := outermostLoop(use, scope); loop != nil && lastKill < loop.Span().Start {
		end := loop.Span().End
		for _, d := range defs {
			if d.pos > pos && d.pos < end && !d.doc && !d.barrier {
				back = append(back, d)
			}
		}
		back, _, _ = e.dropExited(back, use, scope, true)
	}
	return fwd, back, from
}

// elemWrite is a write into an element of a variable: `$x[k] = v`,
// `$x[] = v`, `$x[k] ??= v` (a set, dim non-nil unless appending), a
// destructuring target `[$x[k]] = …` (a nil: unknown value), or a nested
// write `$x[k][…] = v` (nested set, key = k).
type elemWrite struct {
	a      *syntax.Assign
	nested bool
	key    syntax.Expr // nested writes: the first-level key (nil: `$x[][…]`)
}

// dim returns the key expression of a direct write (nil for appends and
// unknown writes).
func (w *elemWrite) dim() syntax.Expr {
	if w.a == nil || w.nested {
		return nil
	}
	return w.a.Var.(*syntax.ArrayDimFetch).Dim
}

// reachingWrites returns the element writes into variable v that can reach
// the read at v: those after the last whole-variable definition hiding them
// (same kill rules as variables), plus the writes later in an enclosing loop
// (back edge). back flags the latter.
func (e *Env) reachingWrites(v *syntax.Variable) (ws []*elemWrite, back []bool) {
	scope := syntax.EnclosingFuncLike(v)
	sv := e.scopeVars(scope)
	if _, ok := scope.(*syntax.ArrowFunction); ok && len(sv.defs[v.Name]) == 0 {
		// Arrow functions capture the enclosing scope by value.
		scope = syntax.EnclosingFuncLike(scope)
		sv = e.scopeVars(scope)
	}
	defs := sv.elemDefs(v.Name)
	if defs == nil {
		return nil, nil
	}
	if len(defs) > maxVarDefs {
		// Too many to walk per read (see maxVarDefs): one unknown write.
		return []*elemWrite{{}}, []bool{false}
	}
	fwd, bk, _ := e.reaching(defs, v, scope)
	for _, d := range fwd {
		if d.w != nil {
			ws = append(ws, d.w)
			back = append(back, false)
		}
	}
	for _, d := range bk {
		if d.w != nil {
			ws = append(ws, d.w)
			back = append(back, true)
		}
	}
	return ws, back
}

// elemDefs merges the definitions of variable name with the writes into its
// elements, in source order (nil when the variable has no element writes).
func (sv *scopeVars) elemDefs(name string) []varDef {
	if d, ok := sv.edefs[name]; ok {
		return d
	}
	ws := sv.elemWrites[name]
	var out []varDef
	if len(ws) > 0 {
		out = make([]varDef, 0, len(sv.defs[name])+len(ws))
		out = append(out, sv.defs[name]...)
		out = append(out, ws...)
		sortDefs(out)
	}
	if sv.edefs == nil {
		sv.edefs = map[string][]varDef{}
	}
	sv.edefs[name] = out
	return out
}

// writeDominates reports whether an element write into variable v runs on
// every path to the read v after the variable's last definition: a
// `$v[k] = x;` statement after every definition reaching v, directly in a
// block enclosing v, with no mutation of the variable (assignment,
// reference, by-reference argument) in between.
func (e *Env) writeDominates(v *syntax.Variable) bool {
	scope := syntax.EnclosingFuncLike(v)
	sv := e.scopeVars(scope)
	if _, ok := scope.(*syntax.ArrowFunction); ok && len(sv.defs[v.Name]) == 0 {
		scope = syntax.EnclosingFuncLike(scope)
		sv = e.scopeVars(scope)
	}
	defs := sv.elemDefs(v.Name)
	if len(defs) > maxVarDefs {
		return false
	}
	fwd, _, _ := e.reaching(defs, v, scope)
	last := uint32(0)
	for _, d := range fwd {
		if d.w == nil {
			last = max(last, d.pos, d.end)
		}
	}
	at := v.Span().Start
	for _, d := range fwd {
		if d.w == nil || d.w.a == nil || d.pos < last {
			continue
		}
		es, ok := d.w.a.Parent().(*syntax.ExprStmt)
		if !ok {
			continue
		}
		blk, ok := es.Parent().(*syntax.Block)
		if !ok || at < blk.Span().Start || at >= blk.Span().End {
			continue
		}
		if !e.nonEmptyBroken(scope, v.Name, d.w.a.Span().End, v) {
			return true
		}
	}
	return false
}

// dropOtherCases removes from fwd the definitions made in an earlier case
// of a switch enclosing use that cannot fall through to use's case: a
// case between them (the defining one included) ends by leaving the
// switch (`case 1: $d = []; break; case 2: use($d);`). An enclosing loop
// still brings them back through the back edge (see reaching).
func dropOtherCases(fwd []varDef, use, scope syntax.Node) []varDef {
	var child syntax.Node = use
	for p := use.Parent(); p != nil && p != scope && len(fwd) > 0; child, p = p, p.Parent() {
		c, ok := p.(*syntax.Case)
		if !ok || child == syntax.Node(c.Cond) {
			continue
		}
		sw := c.Parent().(*syntax.Switch)
		if inLoop(sw, scope) {
			continue // the next iteration enters any case again
		}
		j := nodeIndex(sw.Cases, c)
		last := -1 // last case before j that leaves the switch
		for k := j - 1; k >= 0 && k >= j-maxBranchScan; k-- {
			if caseExits(sw.Cases[k]) {
				last = k
				break
			}
		}
		if last < 0 {
			continue
		}
		lo, hi := sw.Cases[0].Span().Start, sw.Cases[last].Span().End
		kept := fwd[:0]
		for _, d := range fwd {
			if d.pos >= lo && d.pos < hi {
				continue
			}
			kept = append(kept, d)
		}
		fwd = kept
	}
	return fwd
}

// caseExits reports whether case c ends with break, continue, return,
// throw or exit (not goto, which may jump into a later case).
func caseExits(c *syntax.Case) bool {
	if len(c.Stmts) == 0 {
		return false
	}
	s := c.Stmts[len(c.Stmts)-1]
	for {
		b, ok := s.(*syntax.Block)
		if !ok || len(b.Stmts) == 0 {
			break
		}
		s = b.Stmts[len(b.Stmts)-1]
	}
	if _, ok := s.(*syntax.Goto); ok {
		return false
	}
	return terminates(s)
}

// dropExclusive removes from fwd the definitions made in a branch that
// excludes the one holding use (`if (is_int($c)) { $c = []; } else {
// use($c); }`): another if/elseif/else body, the other ternary branch,
// another match arm. Only for branchings outside every loop around use:
// inside a loop the other branch runs on an earlier iteration.
func dropExclusive(fwd []varDef, use, scope syntax.Node) []varDef {
	var excl []syntax.Span
	var child syntax.Node = use
	for p := use.Parent(); p != nil && p != scope && len(fwd) > 0; child, p = p, p.Parent() {
		switch n := p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			excl = excl[:0] // branchings below a loop do not exclude
		case *syntax.If:
			if child == syntax.Node(n.Cond) {
				continue
			}
			add := func(b syntax.Node) {
				if b != nil && b != child {
					excl = append(excl, b.Span())
				}
			}
			if len(n.ElseIfs) >= maxBranchScan {
				continue // hostile chains: keep the definitions (sound)
			}
			add(n.Body)
			for _, ei := range n.ElseIfs {
				add(ei.Body)
			}
			if n.Else != nil {
				add(n.Else)
			}
		case *syntax.Ternary:
			switch {
			case n.Then != nil && child == syntax.Node(n.Then):
				excl = append(excl, n.Else.Span())
			case n.Then != nil && child == syntax.Node(n.Else):
				excl = append(excl, n.Then.Span())
			}
		case *syntax.MatchArm:
			if m := n.Parent().(*syntax.Match); child == syntax.Node(n.Body) && len(m.Arms) <= maxBranchScan {
				for _, a := range m.Arms {
					if a != n {
						excl = append(excl, a.Body.Span())
					}
				}
			}
		}
	}
	if len(excl) == 0 {
		return fwd
	}
	kept := fwd[:0]
	for _, d := range fwd {
		in := false
		for _, sp := range excl {
			if d.pos >= sp.Start && d.pos < sp.End {
				in = true
				break
			}
		}
		if !in {
			kept = append(kept, d)
		}
	}
	return kept
}

// exitRegion is a statement list prefix that always ends by leaving: from
// the list's start to the end of its first statement that always exits
// (return, throw, exit; a break or an if/else whose branches all exit).
// A definition inside it reaches no read outside it within limit (the
// function, or the loop/switch a break leaves).
type exitRegion struct {
	span, limit syntax.Span
	list        syntax.Span // the whole statement list (code after the exit is dead)
	kind        int         // exitContinue, exitBreak or exitFunc
	parent      int         // index of the innermost enclosing region, -1 when none
}

const (
	exitNone = iota
	exitContinue
	exitBreak
	exitFunc
)

// exitKind classifies a statement that always leaves: exitFunc (return,
// throw, exit), exitBreak (a plain break, or a mix with exitFunc).
func exitKind(s syntax.Stmt) int {
	switch n := s.(type) {
	case *syntax.Return:
		return exitFunc
	case *syntax.Break:
		if n.Num == nil {
			return exitBreak
		}
	case *syntax.Continue:
		if n.Num == nil {
			return exitContinue
		}
	case *syntax.ExprStmt:
		switch n.Expr.(type) {
		case *syntax.Exit, *syntax.Throw:
			return exitFunc
		}
	case *syntax.Block:
		for _, st := range n.Stmts {
			if k := exitKind(st); k != exitNone || jumps(st) {
				return k
			}
		}
	case *syntax.If:
		if n.Else == nil {
			return exitNone
		}
		k := exitKind(n.Body)
		for _, ei := range n.ElseIfs {
			k = min(k, exitKind(ei.Body))
		}
		return min(k, exitKind(n.Else.Body))
	}
	return exitNone
}

// jumps reports a statement that leaves to another point of the function
// (goto, a multi-level break or continue): what follows it in its list is
// unreachable, but definitions before it still reach code elsewhere.
func jumps(s syntax.Stmt) bool {
	switch n := s.(type) {
	case *syntax.Goto:
		return true
	case *syntax.Break:
		return n.Num != nil
	case *syntax.Continue:
		return n.Num != nil
	}
	return false
}

// exitRegions returns the exit regions of scope, sorted by start (outer
// first among equal starts), each with its parent region.
func (e *Env) exitRegions(scope syntax.Node) []exitRegion {
	sv := e.scopeVars(scope)
	if sv.exitsDone {
		return sv.exits
	}
	sv.exitsDone = true
	whole := syntax.Span{Start: 0, End: uint32(len(e.File.Src))}
	if scope != nil {
		whole = scope.Span()
	}
	var out []exitRegion
	list := func(ls syntax.Span, stmts []syntax.Stmt) {
		start := ls.Start
		for _, st := range stmts {
			k := exitKind(st)
			if jumps(st) {
				return // the rest of the list never runs
			}
			if k == exitNone {
				continue
			}
			limit := whole
			for p := st.Parent(); p != nil && p != scope; p = p.Parent() {
				switch p.(type) {
				case *syntax.Try:
					return // catch/finally may resume: no region
				case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile, *syntax.Switch:
					if k == exitContinue && limit == whole {
						if _, sw := p.(*syntax.Switch); sw {
							k = exitBreak // continue in a switch acts as break
						} else {
							limit = p.Span() // later iterations: see dropExited
							continue
						}
					}
					if k != exitBreak {
						break
					}
					if limit != whole {
						// A loop around the construct the break leaves
						// comes back to it: no region.
						if _, sw := p.(*syntax.Switch); !sw {
							return
						}
						break
					}
					limit = p.Span()
				}
			}
			out = append(out, exitRegion{span: syntax.Span{Start: start, End: st.Span().End}, limit: limit, list: ls, parent: -1, kind: k})
			return
		}
	}
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.Method, *syntax.ClassLike:
			return false
		case *syntax.Block:
			list(n.Span(), n.Stmts)
		case *syntax.Case:
			list(n.Span(), n.Stmts)
		}
		return true
	}
	if scope == nil {
		list(whole, e.File.Stmts)
		for _, st := range e.File.Stmts {
			syntax.Inspect(st, visit)
		}
	} else if body := syntax.FuncLikeBody(scope); body != nil {
		syntax.Inspect(body, visit)
	}
	// Statement lists start at distinct positions (their `{` or `case`).
	slices.SortFunc(out, func(a, b exitRegion) int { return cmp.Compare(a.span.Start, b.span.Start) })
	// Regions nest or are disjoint: a stack sweep finds each parent.
	var stack []int
	for i := range out {
		for len(stack) > 0 && out[stack[len(stack)-1]].span.End <= out[i].span.Start {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			out[i].parent = stack[len(stack)-1]
		}
		stack = append(stack, i)
	}
	sv.exits = out
	return out
}

// dropExited removes from defs those that cannot reach use because every
// path from them leaves first (`$x = 5; if ($c) { $x = 'a'; return; }
// use($x)`: only 5 reaches): the innermost exit region holding the
// definition must not hold use, use being within its limit (else the
// enclosing regions are tried). A definition followed by continue reaches
// the reads of its loop only through the loop's head: for forward
// definitions (back false) it is returned in via with that loop's start,
// for the caller to treat as a back-edge definition; back-edge definitions
// (back true) are kept.
func (e *Env) dropExited(defs []varDef, use, scope syntax.Node, back bool) (kept, via []varDef, loops []uint32) {
	if len(defs) == 0 {
		return defs, nil, nil
	}
	rs := e.exitRegions(scope)
	if len(rs) == 0 {
		return defs, nil, nil
	}
	at := use.Span().Start
	kept = defs[:0]
	for _, d := range defs {
		switch kind := exitedKind(rs, d.pos, at); {
		case kind == exitNone:
			kept = append(kept, d)
		case kind != exitContinue:
		case !back:
			via, loops = append(via, d), append(loops, continueLoop(rs, d.pos, at))
		default:
			kept = append(kept, d)
		}
	}
	return kept, via, loops
}

// exitedKind returns the kind of the exit region that keeps a definition
// at p from reaching at (exitNone when none does): the innermost region
// holding p whose statement list does not hold at, at being within its
// limit (else the enclosing regions are tried).
func exitedKind(rs []exitRegion, p, at uint32) int {
	if i := deadRegion(rs, p, at); i >= 0 {
		return rs[i].kind
	}
	return exitNone
}

// continueLoop returns the start of the loop a continue region keeping p
// from at leaves to.
func continueLoop(rs []exitRegion, p, at uint32) uint32 {
	return rs[deadRegion(rs, p, at)].limit.Start
}

// deadRegion is the index of the region exitedKind finds, -1 when none.
func deadRegion(rs []exitRegion, p, at uint32) int {
	in := func(sp syntax.Span, p uint32) bool { return p >= sp.Start && p < sp.End }
	i, _ := slices.BinarySearchFunc(rs, p, func(r exitRegion, p uint32) int {
		if r.span.Start <= p {
			return -1
		}
		return 1
	})
	i--
	for i >= 0 && !in(rs[i].span, p) {
		i = rs[i].parent
	}
	for ; i >= 0; i = rs[i].parent {
		if in(rs[i].list, at) {
			return -1
		}
		if in(rs[i].limit, at) {
			return i
		}
	}
	return -1
}

// exited reports whether a definition at p cannot reach use because every
// path from it leaves first (return, throw, exit, break; not continue,
// which comes back through the loop head). The SpecOnly T-rules typer,
// which ignores reaching definitions otherwise, applies it.
func (e *Env) exited(p uint32, use, scope syntax.Node) bool {
	k := exitedKind(e.exitRegions(scope), p, use.Span().Start)
	return k != exitNone && k != exitContinue
}

// inLoop reports whether a loop of scope encloses n.
func inLoop(n, scope syntax.Node) bool {
	for p := n.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			return true
		}
	}
	return false
}

// docFits reports whether inline annotation d may override the
// definitions fwd reaching it. It may, except in the class-name idiom: a
// standalone `/** @var Widget $class */` over a variable holding strings
// (a class name used for `$class::widget()`) describes the class, not the
// value, so the strings stay (PhpStorm uses such hints for completion;
// typing the string as an instance would be wrong).
func (e *Env) docFits(d varDef, fwd []varDef) bool {
	dt := d.typ()
	if dt.IsUnknown() || len(dt.Classes()) != len(dt.Atoms()) {
		return true
	}
	for _, f := range fwd {
		if ft := f.typ(); f.w != nil || !ft.OnlyOf("string") {
			return true
		}
	}
	return false
}
