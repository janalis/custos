package types

import (
	"slices"
	"sort"
	"strings"
)

// genEntry holds the generic arguments of one atom: a class atom written
// `Collection<int, Foo>` in a doc type, or `iterable<K, V>`. Like array
// facts, generic arguments never change the atom set: `Collection<Foo>` is
// the atom `\Collection`, so Has, Classes, Equal and String behave as
// without them.
type genEntry struct {
	atom string
	args []Type
	sig  *callSig // what a written callable signature says of its parameters
	// props is the shape of an `object{a: int, b?: string}` atom.
	props *objectShape
}

// objectShape is the properties an object shape lists. Unlike an array
// shape it describes a contract on any object, so it only ever adds to the
// `object` atom it sits on.
type objectShape struct {
	keys   []ShapeKey
	sealed bool
}

// sameProps reports two object shapes listing the same properties.
func sameProps(a, b *objectShape) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.sealed != b.sealed || len(a.keys) != len(b.keys) {
		return false
	}
	for i, k := range a.keys {
		o := b.keys[i]
		if k.Name != o.Name || k.Optional != o.Optional || k.Type.DocString() != o.Type.DocString() {
			return false
		}
	}
	return true
}

// sameEntry reports two entries saying the same of their atom.
func sameEntry(a, b *genEntry) bool {
	return sameArgs(a.args, b.args) && sameSig(a.sig, b.sig) && sameProps(a.props, b.props)
}

// genOf returns the entry of atom, nil when t has none.
func (t Type) genOf(atom string) *genEntry {
	for i := range t.gen {
		if t.gen[i].atom == atom {
			return &t.gen[i]
		}
	}
	return nil
}

// callSig is the shape of the parameters a callable signature was written
// with, beside their types in genEntry.args. A callable without one (no
// signature, or one rebuilt from its types alone) has an unknown arity.
//
// A variadic parameter stays out of genEntry.args, whose positions each take
// one argument: its type is rest.
type callSig struct {
	required int    // leading parameters that must be passed
	rest     *Type  // the variadic parameter's type; nil when there is none
	byRef    uint64 // bit i: parameter i is by reference (the first 64)
}

// sameSig reports two signatures saying the same of their parameters.
func sameSig(a, b *callSig) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.required != b.required || a.byRef != b.byRef || (a.rest == nil) != (b.rest == nil) {
		return false
	}
	return a.rest == nil || a.rest.DocString() == b.rest.DocString()
}

// withCallSig returns t, whose atom carries generic arguments, with sig on
// that atom.
func (t Type) withCallSig(atom string, sig *callSig) Type {
	for i := range t.gen {
		if t.gen[i].atom == atom {
			t.gen[i].sig = sig
		}
	}
	return t
}

// TypeArgs returns the generic arguments of atom (nil when none are known).
func (t Type) TypeArgs(atom string) []Type {
	atom = normalizeAtom(atom)
	for _, g := range t.gen {
		if g.atom == atom {
			return g.args
		}
	}
	return nil
}

// callSigOf returns the signature shape written for atom, nil when unknown.
func (t Type) callSigOf(atom string) *callSig {
	if e := t.genOf(atom); e != nil {
		return e.sig
	}
	return nil
}

// objectShapeDoc is the doc syntax of atom's object shape, "" without one.
func (t Type) objectShapeDoc(atom string) string {
	e := t.genOf(atom)
	if e == nil || e.props == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("object{")
	writeShapeBody(&b, e.props.keys, e.props.sealed)
	b.WriteByte('}')
	return b.String()
}

// WithTypeArgs returns t with generic arguments args on atom (which t must
// have; otherwise t is returned unchanged). Nil args remove them.
func (t Type) WithTypeArgs(atom string, args []Type) Type {
	atom = normalizeAtom(atom)
	if !t.Has(atom) {
		return t
	}
	out := make([]genEntry, 0, len(t.gen)+1)
	for _, g := range t.gen {
		if g.atom != atom {
			out = append(out, g)
		}
	}
	if len(args) > 0 {
		out = append(out, genEntry{atom: atom, args: args})
	}
	return t.withGen(out)
}

// WithoutClassTypeArgs returns t without the generic arguments of its class
// atoms, in its element type and shape values too. A closure keeps its signature.
func (t Type) WithoutClassTypeArgs() Type {
	if t.element != nil {
		element := t.element.WithoutClassTypeArgs()
		t.element = &element
	}
	if t.arr != nil && len(t.arr.keys) > 0 {
		arr := *t.arr
		arr.keys = slices.Clone(arr.keys)
		for i := range arr.keys {
			arr.keys[i].Type = arr.keys[i].Type.WithoutClassTypeArgs()
		}
		t.arr = &arr
	}
	kept := make([]genEntry, 0, len(t.gen))
	for _, g := range t.gen {
		if !IsClassAtom(g.atom) || g.atom == `\Closure` {
			kept = append(kept, g)
		}
	}
	return t.withGen(kept)
}

// WithTypeArgsFrom copies the generic arguments src has on atoms of t.
func (t Type) WithTypeArgsFrom(src Type) Type {
	if len(src.gen) == 0 {
		return t
	}
	return t.withGen(append(append([]genEntry(nil), t.gen...), src.gen...))
}

// withGen attaches the entries of gen whose atom t has (the first entry of
// an atom wins), sorted by atom.
func (t Type) withGen(gen []genEntry) Type {
	if len(gen) == 0 {
		t.gen = nil
		return t
	}
	var out []genEntry
	for _, g := range gen {
		if !t.Has(g.atom) || len(g.args) == 0 && g.props == nil {
			continue
		}
		dup := false
		for _, o := range out {
			if o.atom == g.atom {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].atom < out[j].atom })
	t.gen = out
	return t
}

// unionGen keeps the generic arguments of an atom of u when every member of
// ts having that atom carries the same arguments.
func unionGen(u Type, ts []Type) []genEntry {
	any := false
	for _, t := range ts {
		if len(t.gen) > 0 {
			any = true
			break
		}
	}
	if !any {
		return nil
	}
	var out []genEntry
	for _, a := range u.atoms {
		var first *genEntry
		ok := true
		for _, t := range ts {
			if !t.Has(a) {
				continue
			}
			e := t.genOf(a)
			switch {
			case e == nil:
				ok = false
			case first == nil:
				first = e
			case !sameEntry(first, e):
				ok = false
			}
			if !ok {
				break
			}
		}
		if ok && first != nil {
			out = append(out, *first)
		}
	}
	return out
}

// mergeGen combines the entries of the members of one doc union: an atom
// listed with different arguments loses them.
func mergeGen(lists ...[]genEntry) []genEntry {
	var out []genEntry
	drop := map[string]bool{}
	for _, l := range lists {
		for _, g := range l {
			found := false
			for _, o := range out {
				if o.atom == g.atom {
					found = true
					if !sameEntry(&o, &g) {
						drop[g.atom] = true
					}
				}
			}
			if !found {
				out = append(out, g)
			}
		}
	}
	if len(drop) == 0 {
		return out
	}
	kept := out[:0]
	for _, g := range out {
		if !drop[g.atom] {
			kept = append(kept, g)
		}
	}
	return kept
}

func sameArgs(a, b []Type) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].DocString() != b[i].DocString() {
			return false
		}
	}
	return true
}

// ClassString is `class-string<C>`: the string atom carrying the class
// type c (class atoms only) as generic argument. `Foo::class` has this type.
func ClassString(c Type) Type {
	if !isClassUnion(c) {
		return String
	}
	return String.WithTypeArgs("string", []Type{c})
}

// ClassStringOf returns the class type C of a `class-string<C>` type t (a
// string, possibly nullable or false, carrying it); unknown otherwise.
func ClassStringOf(t Type) Type {
	if !t.Has("string") || !t.Without("null", "false").OnlyOf("string") {
		return Unknown
	}
	if args := t.TypeArgs("string"); len(args) == 1 && isClassUnion(args[0]) {
		return args[0]
	}
	return Unknown
}

// isClassUnion reports a known type made of class atoms only.
func isClassUnion(t Type) bool {
	return !t.IsUnknown() && len(t.Classes()) == len(t.atoms)
}

// CallableReturn returns the return type R that a `callable(…): R` or
// `Closure(…): R` signature gives the callable members of t, or that a
// closure's inferred body gives it (see WithCallableReturn); unknown when
// t has no such member or its members disagree.
func CallableReturn(t Type) Type {
	var ret Type
	for _, a := range [...]string{"callable", `\Closure`} {
		if !t.Has(a) {
			continue
		}
		args := t.TypeArgs(a)
		if len(args) == 0 || args[0].IsUnknown() {
			return Unknown
		}
		if !ret.IsUnknown() && ret.DocString() != args[0].DocString() {
			return Unknown
		}
		ret = args[0]
	}
	return ret
}

// CallableParam returns the type that a `callable(…)` or `Closure(…)`
// signature documents for parameter i of the callable members of t;
// unknown when it documents none or the members disagree.
func CallableParam(t Type, i int) Type {
	var p Type
	for _, a := range [...]string{"callable", `\Closure`} {
		if !t.Has(a) {
			continue
		}
		args := t.TypeArgs(a)
		if i+1 >= len(args) || args[i+1].IsUnknown() {
			return Unknown
		}
		if !p.IsUnknown() && p.DocString() != args[i+1].DocString() {
			return Unknown
		}
		p = args[i+1]
	}
	return p
}

// WithCallableReturn returns t (which has the callable or \Closure atom)
// with return type ret on that atom; an unknown or mixed ret leaves t
// unchanged.
func WithCallableReturn(t Type, atom string, ret Type) Type {
	if ret.IsUnknown() || ret.Has("mixed") {
		return t
	}
	return t.WithTypeArgs(atom, []Type{ret})
}

// isCallableAtom reports the atoms whose generic arguments are a callable
// signature: its return type, then its parameter types.
func isCallableAtom(atom string) bool { return atom == "callable" || atom == `\Closure` }

// refMark is `&` when parameter i is by reference.
func (s *callSig) refMark(i int) string {
	if i < 64 && s.byRef&(1<<i) != 0 {
		return "&"
	}
	return ""
}

// paramSuffix is the doc syntax after the type of parameter i: `&` for by
// reference, then `=` for an optional one.
func (s *callSig) paramSuffix(i int) string {
	if s == nil {
		return ""
	}
	if i >= s.required {
		return s.refMark(i) + "="
	}
	return s.refMark(i)
}

// docOrMixed is the doc syntax of t, mixed when unknown.
func docOrMixed(t Type) string {
	if t.IsUnknown() {
		return "mixed"
	}
	return t.DocString()
}

// genString renders atom with its generic arguments (doc syntax); the
// signature of a callable as `callable(P1, P2): (R)`.
func genString(atom string, args []Type, sig *callSig) string {
	var b strings.Builder
	if isCallableAtom(atom) {
		b.WriteString(atom)
		b.WriteByte('(')
		for i, p := range args[1:] {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(docOrMixed(p))
			b.WriteString(sig.paramSuffix(i))
		}
		if sig != nil && sig.rest != nil {
			if len(args) > 1 {
				b.WriteString(", ")
			}
			b.WriteString(docOrMixed(*sig.rest))
			b.WriteString(sig.refMark(len(args) - 1))
			b.WriteString("...")
		}
		b.WriteString("): (")
		b.WriteString(docOrMixed(args[0]))
		b.WriteByte(')')
		return b.String()
	}
	if atom == "string" {
		atom = "class-string"
	}
	b.WriteString(atom)
	b.WriteByte('<')
	for i, a := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(docOrMixed(a))
	}
	b.WriteByte('>')
	return b.String()
}
