package index

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// constantLookup keeps the effective importer through nested traits, but
// starts a new owner when entering an inherited class. Ambiguous trait
// declarations and incomplete traversal produce no constant.
type constantLookup struct {
	ix             *Index
	ver            phpversion.Version
	path           [MaxAncestors]string
	depth, visited int
}

func (l *constantLookup) find(class, name, owner string) (*ClassConst, bool) {
	if l.visited == MaxAncestors {
		return nil, true
	}
	k := key(class)
	for _, active := range l.path[:l.depth] {
		if active == k {
			return nil, true
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
	if c.Kind != syntax.KindTrait {
		owner = c.FQN
	}
	if declaration := c.Consts[name]; declaration != nil {
		if c.Kind == syntax.KindTrait && owner != "" {
			copy := *declaration
			copy.TypeClass = owner
			return &copy, false
		}
		return declaration, false
	}
	var found *ClassConst
	for _, trait := range c.Traits {
		candidate, incomplete := l.find(trait, name, owner)
		if incomplete {
			return nil, true
		}
		if candidate == nil {
			continue
		}
		if found != nil && !compatibleConstants(found, candidate) {
			return nil, true
		}
		found = candidate
	}
	if found != nil {
		return found, false
	}
	if c.Parent != "" {
		if inherited, incomplete := l.find(c.Parent, name, ""); inherited != nil || incomplete {
			return inherited, incomplete
		}
	}
	for _, iface := range c.Interfaces {
		if inherited, incomplete := l.find(iface, name, ""); inherited != nil || incomplete {
			return inherited, incomplete
		}
	}
	return nil, false
}

// Different expressions may evaluate to the same value; without evaluating
// them, only matching declarations can safely compose.
func compatibleConstants(a, b *ClassConst) bool {
	return a.Type == b.Type && a.Value == b.Value && a.Visibility == b.Visibility && a.Final == b.Final && a.Case == b.Case
}
