package types

import "math"

// StringLengths builds the string type whose byte lengths lie in ranges
// (unknown when they hold none). Like int values (see ints.go), lengths ride
// on the string atom: String and Equal ignore them, ShapeString shows them.
func StringLengths(ranges ...IntRange) Type {
	rs := rangesWithin(ranges, 0, math.MaxInt64)
	if len(rs) == 0 {
		return Unknown
	}
	return Of("string").withLens(rs)
}

// StringLengthRanges returns the byte lengths known for the string member,
// false when they are unconstrained (or the member is absent).
func (t Type) StringLengthRanges() ([]IntRange, bool) {
	if t.lens == nil || !t.Has("string") {
		return nil, false
	}
	return t.lens, true
}

// LengthsOnly reports whether t is a string whose lengths all lie in
// [lo, hi].
func (t Type) LengthsOnly(lo, hi int64) bool {
	rs, ok := t.StringLengthRanges()
	return ok && len(t.atoms) == 1 && rs[0].Lo >= lo && rs[len(rs)-1].Hi <= hi
}

// LengthsWithin keeps the lengths of the string member in [lo, hi] (from
// any length when unconstrained), dropping the member when none remains.
func (t Type) LengthsWithin(lo, hi int64) Type {
	if !t.Has("string") {
		return t
	}
	return t.withLens(rangesWithin(t.lengths(), lo, hi))
}

// LengthsWithout removes length v from the string member (from any length
// when unconstrained), dropping the member when none remains.
func (t Type) LengthsWithout(v int64) Type {
	if !t.Has("string") {
		return t
	}
	return t.withLens(rangesWithout(t.lengths(), v))
}

// LengthsAlso keeps the lengths of t's string member that o's may also
// have: all of them when o's are unconstrained, none when o has no string.
func (t Type) LengthsAlso(o Type) Type {
	if !t.Has("string") {
		return t
	}
	if !o.Has("string") {
		return t.Without("string")
	}
	if o.lens == nil {
		return t
	}
	return t.withLens(rangesMeet(t.lengths(), o.lens))
}

// lengths returns the string member's lengths, any length when unconstrained.
func (t Type) lengths() []IntRange {
	if t.lens == nil {
		return []IntRange{{0, math.MaxInt64}}
	}
	return t.lens
}

// withLens replaces the lengths of the string member; none drops it, and
// every length leaves it unconstrained.
func (t Type) withLens(rs []IntRange) Type {
	rs = normalizeInts(rs)
	if len(rs) == 0 {
		return t.Without("string")
	}
	if len(rs) == 1 && rs[0] == (IntRange{0, math.MaxInt64}) {
		rs = nil
	}
	t.lens = rs
	return t
}

// unionLens computes the string lengths of a union, as unionInts.
func unionLens(ts []Type) []IntRange {
	return unionRanges(ts, "string", func(t Type) []IntRange { return t.lens })
}
