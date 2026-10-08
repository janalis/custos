package types

import (
	"slices"
	"strings"

	"custos/internal/syntax"
)

// Intersections: `A&B` (PHPDoc or native, also `(A&B)|null`) keeps its
// class atoms in the atom set, as before, and records them in inter: the
// value is an instance of all of them at once, not of one of them. Has,
// Classes and the other set operations are unchanged; member lookups use
// Intersection to accept a member found on either side. The fact is kept
// only while the type has no other class alternative (a union with another
// class, or narrowing away one side, drops it: plain union semantics).

// Intersection returns the class atoms forming the type's intersection
// member (sorted; nil when it has none).
func (t Type) Intersection() []string {
	if t.inter == nil {
		return nil
	}
	return *t.inter
}

// Intersect is the intersection type of class atoms (one class: that
// class).
func Intersect(classes ...string) Type {
	return Of(classes...).withInter(classes)
}

// withInter records classes (atoms of t) as an intersection; fewer than two
// distinct class atoms, or another class alternative in t, record nothing.
func (t Type) withInter(classes []string) Type {
	t.inter = nil
	if len(classes) < 2 {
		return t
	}
	var in []string
	for _, c := range classes {
		if c = normalizeAtom(c); strings.HasPrefix(c, `\`) && !strings.HasSuffix(c, "[]") && t.Has(c) && !slices.Contains(in, c) {
			in = append(in, c)
		}
	}
	if len(in) < 2 || len(in) != len(t.Classes()) {
		return t
	}
	slices.Sort(in)
	t.inter = &in
	return t
}

// unionInter is the intersection a union of ts keeps: the one intersection
// they share, when no member brings another class alternative.
func unionInter(ts []Type) []string {
	var inter []string
	for _, t := range ts {
		if in := t.Intersection(); in != nil {
			if inter != nil && !slices.Equal(inter, in) {
				return nil
			}
			inter = in
		}
	}
	if inter == nil {
		return nil // the common case: no intersection, no class scan
	}
	for _, t := range ts {
		if t.inter == nil && len(t.Classes()) > 0 {
			return nil
		}
	}
	return inter
}

// joinParts renders the per-atom texts parts of t: `|` between
// alternatives, `&` inside the intersection (`(\A&\B)|null`).
func (t Type) joinParts(parts []string) string {
	inter := t.Intersection()
	if inter == nil {
		return strings.Join(parts, "|")
	}
	var in, out []string
	for i, a := range t.atoms {
		if slices.Contains(inter, a) {
			in = append(in, parts[i])
		} else {
			out = append(out, parts[i])
		}
	}
	s := strings.Join(in, "&")
	if len(out) == 0 {
		return s
	}
	return "(" + s + ")|" + strings.Join(out, "|")
}

// nodeInter returns the class names of the intersection in a native type
// declaration (`A&B`, `(A&B)|null`); nil when there is none.
func nodeInter(n syntax.Expr, resolve Resolver) []string {
	switch t := n.(type) {
	case *syntax.IntersectionType:
		return nodeAtoms(t, resolve)
	case *syntax.UnionType:
		var found []string
		for _, x := range t.Types {
			if in, ok := x.(*syntax.IntersectionType); ok {
				if found != nil {
					return nil // (A&B)|(C&D): alternatives of intersections
				}
				found = nodeAtoms(in, resolve)
			}
		}
		return found
	}
	return nil
}
