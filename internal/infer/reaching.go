package infer

import (
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
	if len(fwd) == 0 {
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
