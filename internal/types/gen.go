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

// HasTypeArgs reports whether any atom carries generic arguments.
func (t Type) HasTypeArgs() bool { return len(t.gen) > 0 }

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

// genString renders atom with its generic arguments (doc syntax).
func genString(atom string, args []Type) string {
	var b strings.Builder
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
