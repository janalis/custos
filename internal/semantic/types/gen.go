package types

import (
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
		if !t.Has(g.atom) || len(g.args) == 0 {
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
		var args []Type
		ok := true
		for _, t := range ts {
			if !t.Has(a) {
				continue
			}
			ta := t.TypeArgs(a)
			switch {
			case ta == nil:
				ok = false
			case args == nil:
				args = ta
			case !sameArgs(args, ta):
				ok = false
			}
			if !ok {
				break
			}
		}
		if ok && args != nil {
			out = append(out, genEntry{atom: a, args: args})
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
					if !sameArgs(o.args, g.args) {
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
		if len(args) != 1 {
			return Unknown
		}
		if !ret.IsUnknown() && ret.DocString() != args[0].DocString() {
			return Unknown
		}
		ret = args[0]
	}
	return ret
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

// isCallableAtom reports the atoms whose generic argument is a callable
// signature's return type.
func isCallableAtom(atom string) bool { return atom == "callable" || atom == `\Closure` }

// genString renders atom with its generic arguments (doc syntax); the
// return type of a callable signature as `callable(): (R)`.
func genString(atom string, args []Type) string {
	var b strings.Builder
	if isCallableAtom(atom) && len(args) == 1 {
		b.WriteString(atom)
		b.WriteString("(): (")
		b.WriteString(args[0].DocString())
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
		if a.IsUnknown() {
			b.WriteString("mixed")
		} else {
			b.WriteString(a.DocString())
		}
	}
	b.WriteByte('>')
	return b.String()
}
