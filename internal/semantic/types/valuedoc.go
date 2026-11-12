package types

import (
	"math"
	"strings"
)

// ValueDocString renders t as DocString does but keeps the values of the
// int member and a non-empty string member, which DocString drops: an
// interval reads `positive-int`, `negative-int`, `non-negative-int`,
// `non-positive-int` or `int<lo, hi>`, a string of at least one byte
// `non-empty-string` (other lengths and string values cannot be kept and
// read `string`). FromDoc parses it back. It is "" when t has neither fact.
func (t Type) ValueDocString() string {
	ints, _ := t.IntRanges()
	nonEmpty := t.Has("string") && len(t.lens) == 1 && t.lens[0] == (IntRange{1, math.MaxInt64})
	if ints == nil && !nonEmpty {
		return ""
	}
	var parts []string
	switch {
	case ints != nil:
		for _, r := range ints {
			parts = append(parts, intDoc(r))
		}
	case t.Has("int"):
		parts = append(parts, "int")
	}
	switch {
	case nonEmpty:
		parts = append(parts, "non-empty-string")
	case t.Has("string"):
		parts = append(parts, "string")
	}
	if rest := t.Without("int", "string"); !rest.IsUnknown() {
		parts = append(parts, rest.DocString())
	}
	return strings.Join(parts, "|")
}

// intDoc is the PHPDoc type of the integers of r.
func intDoc(r IntRange) string {
	const lo, hi = math.MinInt64, math.MaxInt64
	switch r {
	case IntRange{1, hi}:
		return "positive-int"
	case IntRange{lo, -1}:
		return "negative-int"
	case IntRange{0, hi}:
		return "non-negative-int"
	case IntRange{lo, 0}:
		return "non-positive-int"
	}
	return rangesString("int<", []IntRange{r})
}
