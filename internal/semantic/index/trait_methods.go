package index

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// methodLookup bounds both recursion and total work on incomplete or cyclic
// hierarchies. The fixed path keeps ordinary method lookup allocation-free.
type methodLookup struct {
	ix             *Index
	ver            phpversion.Version
	path           [MaxAncestors]string
	depth, visited int
	magic          *Method
}

func (l *methodLookup) find(class, name string, inherited bool) (*Method, bool) {
	if l.visited == MaxAncestors {
		return nil, false
	}
	k := key(class)
	for _, active := range l.path[:l.depth] {
		if active == k {
			return nil, false
		}
	}
	c := l.ix.Class(k, l.ver)
	if c == nil {
		return nil, false
	}
	l.visited++
	l.path[l.depth] = k
	l.depth++
	defer func() { l.depth-- }()
	if m := c.Methods[name]; m != nil && (l.ver == 0 || m.Avail.In(l.ver)) {
		if !m.Magic {
			return m, false
		}
		if l.magic == nil {
			l.magic = m
			if c.Kind == syntax.KindTrait {
				for i := l.depth - 2; i >= 0; i-- {
					if owner := l.ix.Class(l.path[i], l.ver); owner != nil && owner.Kind != syntax.KindTrait {
						copy := *m.at(l.ver)
						copy.TypeClass = owner.FQN
						l.magic = &copy
						break
					}
				}
			}
		}
	}
	// A trait requirement provides a contract only after inherited concrete
	// implementations have had a chance to fulfill it.
	fallback, conflict := l.compose(c, name)
	if fallback != nil && c.Kind != syntax.KindTrait && fallback.TypeClass == "" {
		if source := l.ix.Class(fallback.Class, l.ver); source != nil && source.Kind == syntax.KindTrait {
			copy := *fallback.at(l.ver)
			copy.TypeClass = c.FQN
			fallback = &copy
		}
	}
	if conflict || (fallback != nil && !fallback.Abstract) {
		return fallback, conflict
	}
	if inherited {
		supers := c.Interfaces
		if c.Parent != "" {
			if m, conflict := l.find(c.Parent, name, true); conflict || (m != nil && !m.Abstract) {
				return m, conflict
			} else if fallback == nil {
				fallback = m
			}
		}
		for _, iface := range supers {
			if m, conflict := l.find(iface, name, true); conflict || (m != nil && !m.Abstract) {
				return m, conflict
			} else if fallback == nil {
				fallback = m
			}
		}
	}
	return fallback, false
}

func (l *methodLookup) compose(c *Class, name string) (*Method, bool) {
	var found *Method
	var source string
	var requirement *Method
	var requirementSource string
	var incompatible bool
	// An alias selects its original method even when insteadof excludes that
	// method from the original name's composition.
	for _, a := range c.TraitAdaptations {
		if a.Alias == "" || !strings.EqualFold(a.Alias, name) {
			continue
		}
		m, conflict := l.selectTrait(c, a.Trait, strings.ToLower(a.Method))
		if conflict {
			return nil, true
		}
		if m != nil {
			if found != nil {
				return nil, true
			}
			// Resolve versioned returns before creating a transient alias, so
			// the version cache retains only the stable declaration.
			found = adaptMethod(m.at(l.ver), a)
		}
	}
	if found != nil {
		return found, false
	}
	for _, trait := range c.Traits {
		if excludedTrait(c, trait, name) {
			continue
		}
		m, conflict := l.find(trait, name, false)
		if conflict {
			return nil, true
		}
		if m == nil {
			continue
		}
		// Requirements constrain implementations without conflicting with them.
		if m.Abstract {
			if requirement == nil {
				requirement, requirementSource = m, trait
			} else if !sameAbstractContract(requirement, m) {
				incompatible = true
			}
			continue
		}
		if found != nil {
			if sameTraitMethod(found, m) {
				continue
			}
			return nil, true
		}
		found, source = m, trait
	}
	if found == nil && !incompatible {
		found, source = requirement, requirementSource
	}
	if found != nil {
		for _, a := range c.TraitAdaptations {
			if a.Alias == "" && strings.EqualFold(a.Method, name) && (a.Trait == "" || key(a.Trait) == key(source)) {
				found = adaptMethod(found.at(l.ver), a)
			}
		}
	}
	return found, false
}

func excludedTrait(c *Class, trait, name string) bool {
	for _, a := range c.TraitAdaptations {
		if !strings.EqualFold(a.Method, name) {
			continue
		}
		for _, excluded := range a.Insteadof {
			if key(excluded) == key(trait) {
				return true
			}
		}
	}
	return false
}

func (l *methodLookup) selectTrait(c *Class, trait, name string) (*Method, bool) {
	var found, requirement *Method
	var incompatible bool
	for _, candidate := range c.Traits {
		if trait != "" && key(candidate) != key(trait) {
			continue
		}
		m, conflict := l.find(candidate, name, false)
		if conflict {
			return nil, true
		}
		if m == nil {
			continue
		}
		if m.Abstract {
			if requirement == nil {
				requirement = m
			} else if !sameAbstractContract(requirement, m) {
				incompatible = true
			}
			continue
		}
		if found != nil {
			return nil, true
		}
		found = m
	}
	if found == nil && !incompatible {
		found = requirement
	}
	return found, false
}

func sameAbstractContract(a, b *Method) bool {
	if a.Return != b.Return || a.DocReturn != b.DocReturn || a.Static != b.Static || a.ByRef != b.ByRef || len(a.Params) != len(b.Params) {
		return false
	}
	for i, p := range a.Params {
		if p != b.Params[i] {
			return false
		}
	}
	return true
}

func adaptMethod(m *Method, a TraitAdaptation) *Method {
	if a.Alias == "" && a.Visibility == nil && !a.Final {
		return m
	}
	copy := *m
	if a.Alias != "" {
		copy.Name = a.Alias
	}
	if a.Visibility != nil {
		copy.Visibility = *a.Visibility
	}
	if a.Final {
		copy.Final = true
	}
	return &copy
}

// Repeated imports of the same declaration through a trait diamond do not
// conflict unless one branch changed its modifiers.
func sameTraitMethod(a, b *Method) bool {
	return a == b || (a.Class != "" && a.Class == b.Class && a.Name == b.Name && a.Visibility == b.Visibility && a.Final == b.Final)
}
