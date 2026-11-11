package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Proven types.
//
// The native environment (see Native) ignores PHPDoc, but a local read still
// takes the type of the definitions reaching it even where a write the
// definitions do not show may have replaced the value: through a reference,
// a `global` or `static` alias, a by-reference parameter, or a callee that
// may take the variable (or one of its elements) by reference. The proven
// environment adds those writes: such a read is unknown, so every type it
// gives holds at run time (see ProvenTypeOf for what remains assumed).

// Proven returns the native Env whose local reads are unknown wherever a
// write the definitions do not show may have changed the value (see
// ProvenTypeOf). Cached per native Env.
func (e *Env) Proven() *Env {
	n := e.Native()
	if n.proven {
		return n
	}
	if n.provenTwin == nil {
		n.provenTwin = NewEnv(n.File, n.Names, n.Index, n.PHP, n.options)
		n.provenTwin.native = true
		n.provenTwin.proven = true
		n.provenTwin.gapOwner = n.gapRoot()
	}
	return n.provenTwin
}

// IsProven reports whether e is a proven Env (see Proven).
func (e *Env) IsProven() bool { return e.proven }

// ProvenTypeOf types x from native declarations only, as the native Env
// does, except that a local read is unknown when, between the definitions
// reaching it and the read, the value may have been replaced unseen:
//
//   - the local is aliased (bound by reference either way, captured by
//     reference by a closure, an array holding a reference to it or one of
//     its elements, a `global` or `static` local, a by-reference parameter,
//     or a variable of the file's top-level scope, which `global` reaches)
//     and a call that may run project code lies in between; for a reference
//     binding any write does too, and for a global one a write to
//     `$GLOBALS`;
//   - the local, or one of its elements, is passed to a callee that takes it
//     by reference or that cannot be resolved.
//
// A call may run project code unless it resolves to a builtin function
// without a callable parameter. Code run implicitly (magic methods,
// destructors, autoloaders, set hooks) is not considered.
func (e *Env) ProvenTypeOf(x syntax.Expr) types.Type { return e.Proven().TypeOf(x) }

// unproven reports, for the proven Env, whether local name read at use may
// hold a value its reaching definitions fwd and back do not show (see
// ProvenTypeOf).
func (e *Env) unproven(sv *scopeVars, fwd, back []varDef, name string, use, scope syntax.Node) bool {
	from := use.Span().Start
	for _, d := range fwd {
		from = min(from, max(d.pos, d.end))
	}
	if len(back) > 0 {
		if loop := outermostLoop(use, scope); loop != nil {
			from = min(from, loop.Span().Start)
		}
	}
	elems := e.mutations(scope, dimKey(name, anyElem))
	if e.brokenBy(scope, from, use, e.mutations(scope, argKey(name)), kindsOf(elems, mutCallArg)) {
		return true // passed (or an element of it) where it may be bound by reference
	}
	switch {
	case sv.refs[name] || len(kindsOf(e.mutations(scope, name), mutRef)) > 0 || len(kindsOf(elems, mutRef)) > 0:
		// Bound by reference (or holding one): a write through the other
		// name, or a call reaching it, may change it.
		return e.brokenBy(scope, from, use, e.runsCode(scope, sv), e.mutations(scope, anyWrite))
	case scope == nil || sv.shared[name] || sv.statics[name]:
		return e.brokenBy(scope, from, use, e.runsCode(scope, sv), e.mutations(scope, dimKey("GLOBALS", anyElem)))
	}
	return false
}

// kindsOf returns the mutations of muts of kind k, in source order, except
// the `global` and `static` bindings (see mutation.alias).
func kindsOf(muts []mutation, k mutKind) []mutation {
	var out []mutation
	for _, m := range muts {
		if m.kind == k && !m.alias {
			out = append(out, m)
		}
	}
	return out
}

// runsCode lists the calls of scope that may run project code: every call
// except those resolving to a builtin function without a callable
// parameter (cached in sv).
func (e *Env) runsCode(scope syntax.Node, sv *scopeVars) []mutation {
	if sv.runsDone {
		return sv.runs
	}
	for _, m := range e.mutations(scope, anyCall) {
		if c, ok := m.call.(*syntax.FuncCall); ok && e.isBuiltinCall(c) && !e.takesCallable(c) {
			continue
		}
		m.kind = mutRuns
		sv.runs = append(sv.runs, m)
	}
	sv.runsDone = true
	return sv.runs
}

// takesCallable reports whether builtin call c has a parameter that may
// receive a callable, through which it may run project code.
func (e *Env) takesCallable(c *syntax.FuncCall) bool {
	for _, p := range e.ResolveFunction(c).Params {
		t := strings.ToLower(p.Type + "|" + p.DocType)
		if strings.Contains(t, "callable") || strings.Contains(t, "closure") {
			return true
		}
	}
	return false
}
