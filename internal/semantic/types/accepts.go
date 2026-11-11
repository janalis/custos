package types

import (
	"slices"
	"strings"
)

// Acceptance: whether a value of one type may be passed where another is
// declared (an argument to a parameter, a returned value, a property write).
// The answer has three values: Yes when every value of the actual type is
// accepted, No when none is, and Maybe otherwise, including whenever the
// answer depends on something the types do not show (an unknown or mixed
// value, a template, an unresolved class, a weak-mode coercion). Checks
// report No; Maybe never reports.

// Verdict is a three-valued answer.
type Verdict uint8

const (
	// Maybe is the zero value: nothing is decided.
	Maybe Verdict = iota
	Yes
	No
)

func (v Verdict) String() string {
	switch v {
	case Yes:
		return "yes"
	case No:
		return "no"
	}
	return "maybe"
}

// ClassRelations answers class questions for Accepts; class names are the
// class atoms of types (`\Foo\Bar`, with the leading backslash).
type ClassRelations interface {
	// Resolved reports whether class is known together with its whole
	// hierarchy (parents, interfaces and traits).
	Resolved(class string) bool
	// IsSubtype reports whether child is, extends or implements parent:
	// Maybe when either is not resolved.
	IsSubtype(child, parent string) Verdict
	// IsFinal reports a final class or an enum (false when unknown).
	IsFinal(class string) bool
	// IsInterface reports an interface (false when unknown).
	IsInterface(class string) bool
}

// AcceptContext is where a value is accepted.
type AcceptContext struct {
	// Classes answers class relations; nil makes every class question Maybe.
	Classes ClassRelations
	// Strict is declare(strict_types=1) in effect: scalars are not coerced
	// (int still widens to float).
	Strict bool
	// Internal marks a parameter of a builtin function: outside strict mode,
	// null passed to its scalar parameter is only deprecated.
	Internal bool
}

// Accepts reports whether expected accepts every value of actual (Yes),
// none of them (No) or some only, or cannot tell (Maybe). Each member of
// actual is accepted by the best of the alternatives of expected.
func Accepts(expected, actual Type, ctx AcceptContext) Verdict {
	if expected.IsUnknown() || actual.IsUnknown() {
		return Maybe
	}
	if hasAtom(expected, "mixed") {
		return Yes
	}
	all, none := true, true
	eachMember(expected, actual, ctx, func(v Verdict) bool {
		all, none = all && v == Yes, none && v == No
		return all || none
	})
	switch {
	case all:
		return Yes
	case none:
		return No
	}
	return Maybe
}

// AllRejected reports a definitely wrong value: expected accepts no value
// of actual (Accepts is No).
func AllRejected(expected, actual Type, ctx AcceptContext) bool {
	return Accepts(expected, actual, ctx) == No
}

// SomeRejected reports a possibly wrong value: expected rejects at least one
// member of actual, every value of which is known (a nullable passed where
// null is not allowed). It holds whenever AllRejected does.
func SomeRejected(expected, actual Type, ctx AcceptContext) bool {
	if expected.IsUnknown() || actual.IsUnknown() || hasAtom(expected, "mixed") {
		return false
	}
	rejected := false
	eachMember(expected, actual, ctx, func(v Verdict) bool {
		rejected = v == No
		return !rejected
	})
	return rejected
}

// eachMember calls fn with the verdict on each member of actual, until fn
// returns false. An intersection is one member: an object of all its
// classes.
func eachMember(expected, actual Type, ctx AcceptContext, fn func(Verdict) bool) {
	in := actual.Intersection()
	if in != nil && !fn(acceptIntersection(expected, in, ctx)) {
		return
	}
	for _, a := range actual.Atoms() {
		if in != nil && slices.Contains(in, a) {
			continue
		}
		if !fn(acceptMember(expected, actual, a, ctx)) {
			return
		}
	}
}

// acceptMember judges member a of actual against the alternatives of
// expected: the best answer wins.
func acceptMember(expected, actual Type, a string, ctx AcceptContext) Verdict {
	switch a {
	case "never":
		return Yes // no value
	case "mixed", "void", "static", "self", "parent":
		return Maybe
	}
	if isTemplateAtom(a) {
		return Maybe
	}
	best := No
	in := expected.Intersection()
	switch {
	case in == nil:
	case isClassAtom(a):
		best = classIntersection(in, a, ctx)
	default:
		best = acceptByClass(expected, in[0], actual, a, ctx) // as for one of its classes
	}
	elems := false
	for _, e := range expected.Atoms() {
		if in != nil && slices.Contains(in, e) {
			continue // judged as a whole above
		}
		var v Verdict
		if strings.HasSuffix(e, "[]") {
			if elems {
				continue // the element-typed arrays form one alternative
			}
			elems = true
			v = acceptArrayElems(expected, expected.Elem(), actual, a, ctx)
		} else {
			v = acceptAtom(expected, e, actual, a, ctx)
		}
		switch v {
		case Yes:
			return Yes
		case Maybe:
			best = Maybe
		}
	}
	return best
}

// acceptIntersection judges the intersection member in (class atoms) of
// actual: an alternative accepting one of its classes accepts it; others
// cannot rule it out (a class may implement more than it declares here).
func acceptIntersection(expected Type, in []string, ctx AcceptContext) Verdict {
	for _, c := range in {
		if acceptMember(expected, Of(c), c, ctx) == Yes {
			return Yes
		}
	}
	return Maybe
}

// classIntersection judges class a against the expected intersection in:
// it must be a subtype of all its classes.
func classIntersection(in []string, a string, ctx AcceptContext) Verdict {
	out := Yes
	for _, c := range in {
		switch acceptClass(c, a, ctx) {
		case No:
			return No
		case Maybe:
			out = Maybe
		}
	}
	return out
}

// isScalarAtom reports int, float, string, bool, true or false.
func isScalarAtom(a string) bool {
	switch a {
	case "int", "float", "string", "bool", "true", "false":
		return true
	}
	return false
}

// acceptAtom judges member a of actual against alternative e of expected.
func acceptAtom(expected Type, e string, actual Type, a string, ctx AcceptContext) Verdict {
	switch {
	case e == "void" || e == "never" || e == "static" || e == "self" || e == "parent" || isTemplateAtom(e):
		return Maybe
	case e == "resource" || a == "resource":
		// Stubs still name resources that became objects (CurlHandle).
		if e == a {
			return Yes
		}
		return Maybe
	case isClassAtom(e):
		return acceptByClass(expected, e, actual, a, ctx)
	case isClassAtom(a):
		return acceptClassValue(e, a, ctx)
	}
	switch e {
	case "null":
		return yesIf(a == "null")
	case "array":
		return acceptArray(expected, actual, a)
	case "iterable":
		return acceptIterable(expected, actual, a, ctx)
	case "callable":
		switch a {
		case "callable":
			return Yes
		case "string", "array", "object", "iterable":
			return Maybe // a function name, a [class, method] pair, an invokable object
		}
		return No
	case "object":
		switch a {
		case "object":
			return Yes
		case "callable", "iterable":
			return Maybe
		}
		return No
	}
	return acceptScalar(expected, e, actual, a, ctx)
}

// acceptScalar judges member a against scalar alternative e.
func acceptScalar(expected Type, e string, actual Type, a string, ctx AcceptContext) Verdict {
	if a == "null" {
		if ctx.Internal && !ctx.Strict {
			return Maybe // deprecated, not a TypeError
		}
		return No
	}
	if !isScalarAtom(a) {
		switch a {
		case "callable", "object":
			if e == "string" && (a == "callable" || !ctx.Strict) {
				return Maybe // a function name; an object with __toString
			}
		}
		return No // array, iterable (an array or an object)
	}
	switch {
	case e == a:
		return acceptRefined(expected, e, actual)
	case e == "bool" && (a == "true" || a == "false"):
		return Yes
	case e == "float" && a == "int":
		return Yes
	case (e == "true" || e == "false") && a == "bool":
		return Maybe
	case !ctx.Strict:
		return Maybe // coerced, or not, depending on the value
	}
	return No
}

// acceptRefined judges a member of actual against the same scalar atom e
// of expected, which may limit its values: int ranges, string lengths and
// string values count only against the facts actual carries.
func acceptRefined(expected Type, e string, actual Type) Verdict {
	if expected.ints == nil && expected.lens == nil && expected.strs == nil {
		return Yes // the common case: no value facts
	}
	switch e {
	case "int":
		want, ok := expected.IntRanges()
		if !ok {
			return Yes
		}
		have, ok := actual.IntRanges()
		if !ok {
			return Maybe
		}
		return rangesVerdict(want, have)
	case "string":
		v := Yes
		if want, ok := expected.StringValueList(); ok {
			have, ok := actual.StringValueList()
			if !ok {
				return Maybe
			}
			v = valuesVerdict(want, have)
		}
		if want, ok := expected.StringLengthRanges(); ok && v != No {
			have, ok := actual.StringLengthRanges()
			if !ok {
				if values, listed := actual.StringValueList(); listed {
					have, ok = valueLengths(values), true
				}
			}
			if !ok {
				return Maybe
			}
			v = both(v, rangesVerdict(want, have))
		}
		return v
	}
	return Yes
}

// both combines two conditions a value must meet.
func both(a, b Verdict) Verdict {
	switch {
	case a == No || b == No:
		return No
	case a == Yes && b == Yes:
		return Yes
	}
	return Maybe
}

// rangesVerdict: every value of have in want (Yes), none (No), or some.
func rangesVerdict(want, have []IntRange) Verdict {
	inside, outside := true, true
	for _, h := range have {
		covered := false
		for _, w := range want {
			if w.Lo <= h.Lo && h.Hi <= w.Hi {
				covered = true
			}
			if w.Lo <= h.Hi && h.Lo <= w.Hi {
				outside = false
			}
		}
		inside = inside && covered
	}
	switch {
	case inside:
		return Yes
	case outside:
		return No
	}
	return Maybe
}

// valuesVerdict: every value of have in want (Yes), none (No), or some.
func valuesVerdict(want, have []string) Verdict {
	in, out := true, true
	for _, h := range have {
		found := slices.Contains(want, h)
		in, out = in && found, out && !found
	}
	switch {
	case in:
		return Yes
	case out:
		return No
	}
	return Maybe
}

// valueLengths returns the byte lengths of values as ranges.
func valueLengths(values []string) []IntRange {
	out := make([]IntRange, len(values))
	for i, v := range values {
		out[i] = IntRange{int64(len(v)), int64(len(v))}
	}
	return out
}

// acceptArray judges member a against the plain `array` alternative of
// expected (a non-empty or shaped expected array decides less).
func acceptArray(expected, actual Type, a string) Verdict {
	if a != "array" && !strings.HasSuffix(a, "[]") {
		switch a {
		case "iterable", "callable":
			return Maybe
		}
		return No
	}
	switch {
	case expected.IsNonEmptyArray() && actual.isEmptyArray():
		return No
	case expected.IsNonEmptyArray() && !actual.IsNonEmptyArray(), expected.HasShape():
		return Maybe
	}
	return Yes
}

// acceptIterable judges member a against the `iterable` alternative: arrays
// and Traversable objects, element types compared as for typed arrays.
func acceptIterable(expected, actual Type, a string, ctx AcceptContext) Verdict {
	switch {
	case a == "iterable":
		return Yes
	case a == "array" || strings.HasSuffix(a, "[]"):
		args := expected.TypeArgs("iterable")
		if len(args) == 0 {
			return Yes
		}
		return elemVerdict(args[len(args)-1], actualElem(actual, a), ctx)
	case a == "object" || a == "callable":
		return Maybe
	}
	return No
}

// acceptArrayElems judges member a against the element-typed arrays of
// expected (`int[]`, `list<Foo>`), whose element type is elem.
func acceptArrayElems(expected, elem, actual Type, a string, ctx AcceptContext) Verdict {
	if a != "array" && !strings.HasSuffix(a, "[]") {
		return acceptArray(expected, actual, a)
	}
	if actual.isEmptyArray() {
		return Yes
	}
	v := elemVerdict(elem, actualElem(actual, a), ctx)
	if v == Yes && (!expected.ArrayKey().IsUnknown() || expected.IsNonEmptyArray() || expected.HasShape()) {
		return Maybe // the keys or the size may not fit
	}
	return v
}

// actualElem returns the element type of array member a of actual.
func actualElem(actual Type, a string) Type {
	if a == "array" {
		return Unknown
	}
	return Of(strings.TrimSuffix(a, "[]"))
}

// elemVerdict compares element types: arrays are never coerced, so the
// comparison is strict, and it answers No only when both are known and no
// element can fit.
func elemVerdict(want, have Type, ctx AcceptContext) Verdict {
	return Accepts(want, have, AcceptContext{Classes: ctx.Classes, Strict: true})
}

// acceptByClass judges member a against class alternative e.
func acceptByClass(expected Type, e string, actual Type, a string, ctx AcceptContext) Verdict {
	if isClassAtom(a) {
		v := acceptClass(e, a, ctx)
		if v == Yes {
			if want, ok := expected.EnumCaseNames(e); ok {
				have, ok := actual.EnumCaseNames(a)
				if !ok || a != e {
					return Maybe
				}
				return valuesVerdict(want, have)
			}
		}
		return v
	}
	if ctx.Classes == nil || !ctx.Classes.Resolved(e) {
		return Maybe
	}
	switch a {
	case "object", "iterable":
		return Maybe
	case "callable":
		return Maybe // a Closure, an invokable object
	}
	return No
}

// acceptClass judges class a against class e.
func acceptClass(e, a string, ctx AcceptContext) Verdict {
	if e == a {
		return Yes
	}
	if isTemplateAtom(e) || isTemplateAtom(a) || ctx.Classes == nil {
		return Maybe
	}
	switch ctx.Classes.IsSubtype(a, e) {
	case Yes:
		return Yes
	case Maybe:
		return Maybe
	}
	// A subclass of a may still extend or implement e, unless a is final
	// or both are classes (single inheritance).
	if ctx.Classes.IsFinal(a) || !ctx.Classes.IsInterface(e) && !ctx.Classes.IsInterface(a) {
		return No
	}
	return Maybe
}

// acceptClassValue judges class member a against non-class alternative e.
func acceptClassValue(e, a string, ctx AcceptContext) Verdict {
	if ctx.Classes == nil || !ctx.Classes.Resolved(a) {
		return Maybe
	}
	switch e {
	case "object":
		return Yes
	case "callable":
		if a == `\Closure` {
			return Yes
		}
		return Maybe // __invoke
	case "iterable":
		return acceptClass(`\Traversable`, a, ctx)
	case "string":
		if !ctx.Strict {
			return Maybe // __toString
		}
	}
	return No
}

// isClassAtom reports a class atom (not an element-typed array).
func isClassAtom(a string) bool { return strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]") }

// isTemplateAtom reports a template parameter atom (`\~T`, `\~~T`).
func isTemplateAtom(a string) bool { return strings.HasPrefix(a, `\~`) }

// hasAtom is Has for an atom already normalized.
func hasAtom(t Type, atom string) bool {
	_, found := slices.BinarySearch(t.atoms, atom)
	return found
}

func yesIf(b bool) Verdict {
	if b {
		return Yes
	}
	return No
}
