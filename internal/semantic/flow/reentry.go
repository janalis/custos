package flow

import "custos/internal/php/syntax"

// leaving checks, where the function can be entered again or ends, what is
// assumed of its static variables: each is not false once the function
// wrote it, and nothing else is bound to it. One the function did not write
// holds what it started with or what an invocation that returned left: it
// is left as it was found.
func leaving(f *frame, r *scopeResult) {
	for name, known := range r.statics {
		if known.state == staticHeld {
			r.holds(f, name)
		}
	}
}

// holds checks what is assumed of one static variable.
func (r *scopeResult) holds(f *frame, name string) {
	v, tracked := f.vars.lookup(name)
	if _, shared := f.aliased[name]; shared || tracked && !v.NonFalse && !v.entry {
		if r.broken == nil {
			r.broken = map[string]bool{}
		}
		r.broken[name] = true
	}
}

// idle reports a scope for which code that runs without a call written
// changes nothing more: it is known to run some, holds no static variable
// and matched no request table element.
func (r *scopeResult) idle() bool {
	return r.runs && len(r.statics) == 0 && !r.matching
}

// implicit applies code of the project that runs without a call written: a
// magic or hooked property, an offset of an object, a class that loads on
// first use, an object converted to text, compared with some or iterated.
// Such code can rewrite the request tables and enter the function again.
// Copying a value, testing its identity or its truth and computing with it
// run none.
func (e *Env) implicit(at syntax.Node, f *frame, r *scopeResult) {
	f.unmatchAll(r)
	reenter(at, f, r)
}

// reenter applies a call, or code that runs without one (see implicit), that
// may enter the function again to its static variables: an invocation
// assigns the ones the function assigns. What a
// guard established of them is gone, but that they are not false where
// that is assumed and checked (see leaving). The producer stays, as the
// value such a call leaves is one the function itself assigned; the value
// is marked as possibly another one.
func reenter(at syntax.Node, f *frame, r *scopeResult) {
	r.runs = true
	if len(r.statics) > 0 {
		f.related = nil
	}
	for name, known := range r.statics {
		v, tracked := f.vars.lookup(name)
		if !tracked || known.state == staticKept {
			continue
		}
		held := known.state == staticHeld && v.NonFalse
		v.NonFalse, v.NonNull, v.origins = held, false, v.origins.unguarded(r).excluding(held, false, r)
		v.tested, v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull = nil, nil, false, false, false, false
		v.Invalidated = at.Span().Start + 1
		v.plain, v.deep = v.plain && known.plain, v.deep && known.deep
		f.setVar(name, v)
	}
}
