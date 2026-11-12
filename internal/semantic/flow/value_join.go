package flow

import (
	"cmp"
)

func merge(a, b Value) Value {
	// A value without request data has nothing to escape: it is neutral for
	// the escaping proof of the other operand, whether or not it is known.
	sa, sb := a.Safe, b.Safe
	ua, ub := a.upload, b.upload
	ca, cb := a.scalars, b.scalars
	switch {
	case len(a.Sources) == 0 && len(b.Sources) == 0:
		sa, sb, ua, ub, ca, cb = allContexts, allContexts, 0, 0, false, false
	case len(a.Sources) == 0:
		sa, ua, ca = allContexts, ub, cb
	case len(b.Sources) == 0:
		sb, ub, cb = allContexts, ua, ca
	}
	v := Value{Complete: a.Complete && b.Complete, Safe: sa & sb, headerMask: a.headerMask & b.headerMask, NonFalse: a.NonFalse && b.NonFalse, NonNull: a.NonNull && b.NonNull, LengthKnown: a.LengthKnown && b.LengthKnown && a.Length == b.Length, Length: a.Length}
	if ta, tb := truthOutcomes(a), truthOutcomes(b); ta != 0 && tb != 0 {
		v.truth = ta | tb
	}
	v.interval = joinInts(a, b)
	if sameScalar(a.literal, b.literal) {
		v.literal = a.literal
	}
	if a.storedIdentity == b.storedIdentity {
		v.storedIdentity = a.storedIdentity
	}
	if sameRecord(a.record, b.record) {
		v.record = a.record
	}
	v.elements = common(a.elements, b.elements)
	v.targets = joinCallableValues(a, b)
	v.returned = joinReturned(a.returned, b.returned)
	// Text of the program joined with a matched label is text of the program.
	v.matched = (a.matched || b.matched) && (a.matched || a.written()) && (b.matched || b.written())
	v.entry = a.entry && b.entry
	v.replaced = a.replaced || b.replaced
	v.plain, v.deep = a.plain && b.plain, a.deep && b.deep
	v.scalars = ca && cb
	if ua == ub {
		v.upload = ua
	}
	if a.Expr == b.Expr {
		v.Expr = a.Expr
	}
	if a.Invalidated == b.Invalidated {
		v.Invalidated = a.Invalidated
	} else {
		v.Invalidated = ^uint32(0)
	}
	if a.Identity == b.Identity {
		v.Identity = a.Identity
	}
	var whole bool
	if v.Sources, whole = union(a.Sources, b.Sources); !whole {
		v.Complete = false
	}
	return v
}

// union joins two source lists, each in the order of before, without
// repeating a source. It returns a itself when b adds nothing, so that a
// value joined again with what it already holds stays recognisably the
// same; the result never shares spare capacity a later append could write
// to. Sources past MaxSources are dropped, which the second result reports.
func union(a, b []Source) ([]Source, bool) {
	switch {
	case len(b) == 0 || len(a) == len(b) && &a[0] == &b[0]:
		return a[:len(a):len(a)], true
	case len(a) == 0:
		return b[:len(b):len(b)], true
	}
	i, j := 0, 0
	for ; j < len(b); j++ {
		for i < len(a) && before(a[i], b[j]) {
			i++
		}
		if i == len(a) || a[i] != b[j] {
			break
		}
	}
	if j == len(b) {
		return a[:len(a):len(a)], true
	}
	out := make([]Source, 0, min(len(a)+len(b), MaxSources))
	i, j = 0, 0
	for i < len(a) || j < len(b) {
		var next Source
		switch {
		case j == len(b) || i < len(a) && before(a[i], b[j]):
			next = a[i]
			i++
		case i == len(a) || before(b[j], a[i]):
			next = b[j]
			j++
		default:
			next = a[i]
			next.ValuesOnly = a[i].ValuesOnly && b[j].ValuesOnly
			next.Keys = a[i].Keys || b[j].Keys
			next.htmlDecoded = a[i].htmlDecoded || b[j].htmlDecoded
			i++
			j++
		}
		if len(out) == MaxSources {
			return out, false
		}
		out = append(out, next)
	}
	return out, true
}

// before orders sources: by position, then by what distinguishes two sources
// at one position (a parameter, another file).
func before(a, b Source) bool {
	if a.Span.Start != b.Span.Start {
		return a.Span.Start < b.Span.Start
	}
	return cmp.Or(cmp.Compare(a.Span.End, b.Span.End), cmp.Compare(a.Parameter, b.Parameter), cmp.Compare(a.Path, b.Path), cmp.Compare(a.Kind, b.Kind)) < 0
}

// equal reports two values that hold the same facts. Element sets are
// compared by identity, which merge keeps when nothing changes.
func equal(v, w Value) bool {
	if v.truth != w.truth || v.storedIdentity != w.storedIdentity || !sameRecord(v.record, w.record) || !sameScalar(v.literal, w.literal) || !sameReturned(v.returned, w.returned) || !sameCallableTargets(v.targets, w.targets) || !sameInts(v.interval, w.interval) || v.tested != w.tested || v.unless != w.unless || v.unlessFalse != w.unlessFalse || v.unlessNull != w.unlessNull || v.skippedFalse != w.skippedFalse || v.skippedNull != w.skippedNull || v.matched != w.matched || v.entry != w.entry || v.replaced != w.replaced || v.plain != w.plain || v.deep != w.deep {
		return false
	}
	if v.Complete != w.Complete || v.Safe != w.Safe || v.headerMask != w.headerMask || v.NonFalse != w.NonFalse || v.NonNull != w.NonNull || v.LengthKnown != w.LengthKnown || v.LengthKnown && v.Length != w.Length || v.Expr != w.Expr || v.elements != w.elements || !sameOrigins(v.origins, w.origins) || v.upload != w.upload || v.scalars != w.scalars || v.Identity != w.Identity || v.Invalidated != w.Invalidated || len(v.Sources) != len(w.Sources) {
		return false
	}
	if len(v.Sources) > 0 && &v.Sources[0] == &w.Sources[0] {
		return true
	}
	for i, s := range v.Sources {
		if s != w.Sources[i] {
			return false
		}
	}
	return true
}
