package types

import (
	"math"
	"slices"
	"strconv"
	"strings"
)

// IntRange is a closed interval of integer values; math.MinInt64 and
// math.MaxInt64 stand for unbounded ends.
type IntRange struct{ Lo, Hi int64 }

// maxIntRanges bounds the intervals a type keeps for its int member: past
// it, the gaps between the outermost ones are given up.
const maxIntRanges = 4

// Ints builds the int type holding exactly the values of ranges (unknown
// when they hold none). The values ride on the int atom like array facts:
// String and Equal ignore them, ShapeString shows them.
func Ints(ranges ...IntRange) Type {
	rs := normalizeInts(ranges)
	if len(rs) == 0 {
		return Unknown
	}
	t := Of("int")
	t.ints = rs
	return t
}

// IntRanges returns the values known for the int member, false when it is
// unconstrained (or absent).
func (t Type) IntRanges() ([]IntRange, bool) {
	if t.ints == nil || !t.Has("int") {
		return nil, false
	}
	return t.ints, true
}

// IntsOnly reports whether t is an int whose values all lie in [lo, hi].
func (t Type) IntsOnly(lo, hi int64) bool {
	rs, ok := t.IntRanges()
	return ok && len(t.atoms) == 1 && rs[0].Lo >= lo && rs[len(rs)-1].Hi <= hi
}

// HasInt reports whether t may hold the integer v: its int member is
// unconstrained or holds v.
func (t Type) HasInt(v int64) bool {
	if !t.Has("int") {
		return false
	}
	if t.ints == nil {
		return true
	}
	for _, r := range t.ints {
		if r.Lo <= v && v <= r.Hi {
			return true
		}
	}
	return false
}

// IntsWithin keeps the values of a constrained int member that lie in
// [lo, hi], dropping the member when none does; other members and an
// unconstrained int are kept.
func (t Type) IntsWithin(lo, hi int64) Type {
	if _, ok := t.IntRanges(); !ok {
		return t
	}
	return t.withInts(rangesWithin(t.ints, lo, hi))
}

// IntsWithout removes v from a constrained int member, dropping the member
// when nothing remains; other members and an unconstrained int are kept.
func (t Type) IntsWithout(v int64) Type {
	if _, ok := t.IntRanges(); !ok {
		return t
	}
	return t.withInts(rangesWithout(t.ints, v))
}

// IntsAlso keeps the values of t's int member that o's int member may also
// hold: all of them when o's is unconstrained, none when o has no int.
func (t Type) IntsAlso(o Type) Type {
	if !t.Has("int") {
		return t
	}
	if !o.Has("int") {
		return t.Without("int")
	}
	if o.ints == nil {
		return t
	}
	mine := t.ints
	if mine == nil {
		mine = []IntRange{{math.MinInt64, math.MaxInt64}}
	}
	return t.withInts(rangesMeet(mine, o.ints))
}

// rangesWithin keeps the parts of rs inside [lo, hi].
func rangesWithin(rs []IntRange, lo, hi int64) []IntRange {
	var out []IntRange
	for _, r := range rs {
		r.Lo, r.Hi = max(r.Lo, lo), min(r.Hi, hi)
		if r.Lo <= r.Hi {
			out = append(out, r)
		}
	}
	return out
}

// rangesWithout removes v from rs.
func rangesWithout(rs []IntRange, v int64) []IntRange {
	var out []IntRange
	for _, r := range rs {
		if v < r.Lo || v > r.Hi {
			out = append(out, r)
			continue
		}
		if v > r.Lo {
			out = append(out, IntRange{r.Lo, v - 1})
		}
		if v < r.Hi {
			out = append(out, IntRange{v + 1, r.Hi})
		}
	}
	return out
}

// rangesMeet keeps the values both a and b hold.
func rangesMeet(a, b []IntRange) []IntRange {
	var out []IntRange
	for _, x := range a {
		for _, y := range b {
			if lo, hi := max(x.Lo, y.Lo), min(x.Hi, y.Hi); lo <= hi {
				out = append(out, IntRange{lo, hi})
			}
		}
	}
	return out
}

// withInts replaces the values of the int member; none drops the member.
func (t Type) withInts(rs []IntRange) Type {
	rs = normalizeInts(rs)
	if len(rs) == 0 {
		return t.Without("int")
	}
	t.ints = rs
	return t
}

// normalizeInts sorts, merges overlapping or adjacent intervals and caps
// their number (see maxIntRanges).
func normalizeInts(rs []IntRange) []IntRange {
	var out []IntRange
	for _, r := range rs {
		if r.Lo <= r.Hi {
			out = append(out, r)
		}
	}
	return normalizeOwned(out)
}

// normalizeOwned is normalizeInts over non-empty intervals the caller owns:
// it reorders and merges them in place.
func normalizeOwned(out []IntRange) []IntRange {
	slices.SortFunc(out, func(a, b IntRange) int {
		switch {
		case a.Lo < b.Lo:
			return -1
		case a.Lo > b.Lo:
			return 1
		}
		return 0
	})
	k := 0
	for _, r := range out {
		if k > 0 && (out[k-1].Hi == math.MaxInt64 || r.Lo <= out[k-1].Hi+1) {
			out[k-1].Hi = max(out[k-1].Hi, r.Hi)
			continue
		}
		out[k] = r
		k++
	}
	out = out[:k]
	for len(out) > maxIntRanges {
		// Give up the smallest gap first.
		at, gap := 1, uint64(math.MaxUint64)
		for i := 1; i < len(out); i++ {
			if g := uint64(out[i].Lo - out[i-1].Hi); g < gap {
				at, gap = i, g
			}
		}
		out[at-1].Hi = out[at].Hi
		out = slices.Delete(out, at, at+1)
	}
	return out
}

// unionInts computes the int values of a union: every member with an int
// atom must constrain it, members without one do not count.
func unionInts(ts []Type) []IntRange {
	return unionRanges(ts, "int", func(t Type) []IntRange { return t.ints })
}

// unionRanges merges the value sets field gives for atom over a union.
// Members all holding the same (normalized) set share it.
func unionRanges(ts []Type, atom string, field func(Type) []IntRange) []IntRange {
	var first []IntRange
	same := true
	for _, t := range ts {
		if !t.Has(atom) {
			continue
		}
		rs := field(t)
		switch {
		case rs == nil:
			return nil
		case first == nil:
			first = rs
		default:
			same = same && slices.Equal(rs, first)
		}
	}
	if same {
		return first
	}
	n := 0
	for _, t := range ts {
		if t.Has(atom) {
			n += len(field(t))
		}
	}
	all := make([]IntRange, 0, n)
	for _, t := range ts {
		if t.Has(atom) {
			all = append(all, field(t)...)
		}
	}
	return normalizeOwned(all)
}

// intsString renders the int member with its values: `int<-1, 0>`, joined
// with `|` for several intervals.
func intsString(rs []IntRange) string { return rangesString("int<", rs) }

// rangesString renders value sets as `prefix lo, hi>` joined with `|`.
func rangesString(prefix string, rs []IntRange) string {
	bound := func(v int64, low bool) string {
		switch {
		case low && v == math.MinInt64:
			return "min"
		case !low && v == math.MaxInt64:
			return "max"
		}
		return strconv.FormatInt(v, 10)
	}
	parts := make([]string, len(rs))
	for i, r := range rs {
		parts[i] = prefix + bound(r.Lo, true) + ", " + bound(r.Hi, false) + ">"
	}
	return strings.Join(parts, "|")
}

// smallValues backs the value sets of IntValue and StringLength for values
// in [smallLo, smallHi], so typing a literal allocates nothing: value sets
// are never written in place.
var smallValues = func() []IntRange {
	rs := make([]IntRange, smallHi-smallLo+1)
	for i := range rs {
		v := int64(i) + smallLo
		rs[i] = IntRange{v, v}
	}
	return rs
}()

const smallLo, smallHi = -64, 1024

// single is the value set {v}.
func single(v int64) []IntRange {
	if v < smallLo || v > smallHi {
		return []IntRange{{v, v}}
	}
	i := v - smallLo
	return smallValues[i : i+1 : i+1]
}

// IntValue is the int type holding exactly v: Ints(IntRange{v, v}).
func IntValue(v int64) Type {
	t := Int
	t.ints = single(v)
	return t
}

// HasValues reports whether t carries int values, string lengths or string
// values (see WithoutValues).
func (t Type) HasValues() bool {
	return t.ints != nil || t.HasStringValues()
}

// HasStringValues reports whether t carries string lengths or string
// values (see WithoutStringValues).
func (t Type) HasStringValues() bool {
	return t.lens != nil || t.strs != nil && (t.strs.only != nil || t.strs.not != nil)
}

// WithoutValues drops the int values, string lengths and string values of
// t, keeping its atoms and other facts (a tracked string stays tracked).
func (t Type) WithoutValues() Type {
	t.ints = nil
	return t.WithoutStringValues()
}

// WithoutStringValues drops the string lengths and string values of t,
// keeping its int values, atoms and other facts (a tracked string stays
// tracked).
func (t Type) WithoutStringValues() Type {
	t.lens = nil
	if t.strs != nil {
		t.strs = trackedAny
	}
	return t
}
