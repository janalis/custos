package types

import "slices"

// enumCases lists the cases a value of one enum class may be.
type enumCases struct {
	class string   // the class atom (`\Suit`)
	names []string // sorted, unique case names
}

// EnumCases builds the enum type class holding only the named cases
// (unknown when there are none). Like int values (see ints.go), the cases
// ride on the class atom: String and Equal ignore them, ShapeString shows
// them.
func EnumCases(class string, names ...string) Type {
	if len(names) == 0 {
		return Unknown
	}
	return Of(class).withCases(class, names)
}

// EnumCaseNames returns the cases known for the enum member class, false
// when they are unconstrained (or the member is absent).
func (t Type) EnumCaseNames(class string) ([]string, bool) {
	if t.cases == nil || t.cases.class != normalizeAtom(class) || !t.Has(class) {
		return nil, false
	}
	return t.cases.names, true
}

// CasesOnly keeps case name of the enum member class, dropping the member
// when its known cases exclude name; other members are kept.
func (t Type) CasesOnly(class, name string) Type {
	if !t.Has(class) {
		return t
	}
	if names, ok := t.EnumCaseNames(class); ok && !slices.Contains(names, name) {
		return t.Without(class)
	}
	return t.withCases(class, []string{name})
}

// CasesWithout removes case name from the enum member class, whose cases are
// all when unconstrained, dropping the member when none remains; other
// members are kept.
func (t Type) CasesWithout(class, name string, all []string) Type {
	if !t.Has(class) {
		return t
	}
	names, ok := t.EnumCaseNames(class)
	if !ok {
		names = all
	}
	rest := slices.DeleteFunc(slices.Clone(names), func(n string) bool { return n == name })
	if len(rest) == 0 {
		return t.Without(class)
	}
	return t.withCases(class, rest)
}

// withCases records names as the cases of the enum member class.
func (t Type) withCases(class string, names []string) Type {
	names = slices.Clone(names)
	slices.Sort(names)
	t.cases = &enumCases{class: normalizeAtom(class), names: slices.Compact(names)}
	return t
}

// unionCases computes the cases of a union: every member holding the class
// atom must constrain the same class, members without it do not count.
func unionCases(ts []Type) *enumCases {
	var class string
	for _, t := range ts {
		if t.cases != nil {
			class = t.cases.class
			break
		}
	}
	var names []string
	for _, t := range ts {
		if !t.Has(class) {
			continue
		}
		if t.cases == nil || t.cases.class != class {
			return nil
		}
		names = append(names, t.cases.names...)
	}
	slices.Sort(names)
	return &enumCases{class: class, names: slices.Compact(names)}
}

// casesString renders the enum member with its cases: `\Suit::Hearts|\Suit::Spades`.
func casesString(c *enumCases) string {
	out := ""
	for i, n := range c.names {
		if i > 0 {
			out += "|"
		}
		out += c.class + "::" + n
	}
	return out
}
