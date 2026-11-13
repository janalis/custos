package types

import (
	"slices"
	"strconv"
	"strings"
)

// strFacts are the values the string member may hold. A string literal
// (StringLiteral) and a PHPDoc literal string type carry its value; other
// strings track values only where a caller seeded them (TrackedStrings).
// Like int values (see ints.go) they ride on the atom: String and Equal
// ignore them, ShapeString shows them.
type strFacts struct {
	only []string // sorted, unique; nil when not limited to a list
	not  []string // sorted, unique values ruled out (only when only is nil)
}

// maxStringValues bounds the values listed: past it, exclusions are given
// up (a list of possible values is never cut, the type gives it up whole).
const maxStringValues = 16

// TrackedStrings is the string type whose values comparisons and searches
// with literal strings refine.
func TrackedStrings() Type {
	t := Of("string")
	t.strs = &strFacts{}
	return t
}

// stringValues builds the string type holding exactly values (unknown when
// there are none, a tracked string past maxStringValues).
func stringValues(values ...string) Type {
	if len(values) == 0 {
		return Unknown
	}
	return TrackedStrings().withStrs(sortedUnique(values), nil)
}

// maxLiteralValue bounds the byte length of a literal string whose value
// StringLiteral keeps: longer literals (SQL, templates) keep their length.
const maxLiteralValue = 64

// literalStr holds the facts of one string literal in a single allocation.
type literalStr struct {
	facts strFacts
	value [1]string
}

// StringLiteral is the string type of literal value v: its byte length and,
// for a short value, the value itself.
func StringLiteral(v string) Type {
	t := StringLength(int64(len(v)))
	if len(v) <= maxLiteralValue {
		l := &literalStr{value: [1]string{v}}
		l.facts.only = l.value[:]
		t.strs = &l.facts
	}
	return t
}

// literalStrs holds string facts and their values.
type literalStrs struct {
	facts  strFacts
	values []string
}

// trackedAny is the facts of a tracked string with no known value.
var trackedAny = &strFacts{}

// StringsTracked reports whether the string member's values are tracked.
func (t Type) StringsTracked() bool { return t.strs != nil && t.Has("string") }

// StringValueList returns the values the string member is limited to, false
// when they are not limited (or the member is absent).
func (t Type) StringValueList() ([]string, bool) {
	if t.strs == nil || t.strs.only == nil || !t.Has("string") {
		return nil, false
	}
	return t.strs.only, true
}

// StringExcludes reports whether the string member's tracked values rule v
// out (true as well when there is no string member).
func (t Type) StringExcludes(v string) bool {
	if !t.Has("string") {
		return true
	}
	if t.strs == nil {
		return false
	}
	if t.strs.only != nil {
		return !slices.Contains(t.strs.only, v)
	}
	_, found := slices.BinarySearch(t.strs.not, v)
	return found
}

// StringsOnly keeps the tracked string member's values among values,
// dropping the member when none remains; an untracked member is kept.
func (t Type) StringsOnly(values ...string) Type {
	if t.strs == nil || !t.Has("string") {
		return t
	}
	keep := slices.DeleteFunc(sortedUnique(values), t.StringExcludes)
	if len(keep) == 0 {
		return t.Without("string")
	}
	return t.withStrs(keep, nil)
}

// StringsWithout rules value v out of the tracked string member, dropping
// it when nothing remains; an untracked member is kept.
func (t Type) StringsWithout(v string) Type {
	if t.strs == nil || !t.Has("string") || t.StringExcludes(v) {
		return t
	}
	if t.strs.only != nil {
		rest := slices.DeleteFunc(slices.Clone(t.strs.only), func(o string) bool { return o == v })
		if len(rest) == 0 {
			return t.Without("string")
		}
		return t.withStrs(rest, nil)
	}
	if len(t.strs.not) >= maxStringValues {
		return t // giving up an exclusion is sound
	}
	return t.withStrs(nil, sortedUnique(append(slices.Clone(t.strs.not), v)))
}

// withStrs records the string member's values.
func (t Type) withStrs(only, not []string) Type {
	if len(only) > maxStringValues {
		only = nil
	}
	t.strs = &strFacts{only: only, not: not}
	return t
}

// unionStrs computes the string values of a union: tracked when every
// member holding a string tracks them; a value is ruled out when every such
// member rules it out.
func unionStrs(ts []Type) *strFacts {
	var lists [][]string
	var nots [][]string
	var first *strFacts // the facts of the first member holding a string
	same := true        // every such member shares them
	for _, t := range ts {
		if !t.Has("string") {
			continue
		}
		if t.strs == nil {
			return nil
		}
		if first == nil {
			first = t.strs
		}
		same = same && t.strs == first
		if t.strs.only != nil {
			lists = append(lists, t.strs.only)
		} else {
			nots = append(nots, t.strs.not)
		}
	}
	if same {
		return first // facts are immutable
	}
	if len(nots) == 0 {
		n := 0
		for _, l := range lists {
			n += len(l)
		}
		// One allocation holds the facts and their values.
		f := &literalStrs{values: make([]string, 0, n)}
		for _, l := range lists {
			f.values = append(f.values, l...)
		}
		slices.Sort(f.values)
		f.values = slices.Compact(f.values)
		if len(f.values) > maxStringValues {
			return trackedAny
		}
		f.facts.only = f.values
		return &f.facts
	}
	var not []string
	for _, v := range nots[0] {
		ruled := true
		for _, n := range nots[1:] {
			if _, found := slices.BinarySearch(n, v); !found {
				ruled = false
				break
			}
		}
		for _, l := range lists {
			ruled = ruled && !slices.Contains(l, v)
		}
		if ruled {
			not = append(not, v)
		}
	}
	return &strFacts{not: not}
}

func sortedUnique(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return slices.Compact(out)
}

// strsString renders the string member with its values: `string<'a'|'b'>`,
// or `string<not '.'|'..'>` for exclusions.
func strsString(f *strFacts) string {
	quote := func(vs []string) string {
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = strconv.Quote(v)
		}
		return strings.Join(parts, "|")
	}
	switch {
	case f.only != nil:
		return "string<" + quote(f.only) + ">"
	case len(f.not) > 0:
		return "string<not " + quote(f.not) + ">"
	}
	return "string"
}
