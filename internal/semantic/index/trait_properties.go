package index

import (
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// propertyLookup preserves the class into which a nested trait property was
// imported. Work and recursion share the method lookup's hierarchy bound.
type propertyLookup struct {
	ix             *Index
	ver            phpversion.Version
	path           [MaxAncestors]string
	depth, visited int
}

func (l *propertyLookup) find(class, name, owner string) *Property {
	if l.visited == MaxAncestors {
		return nil
	}
	k := key(class)
	for _, active := range l.path[:l.depth] {
		if active == k {
			return nil
		}
	}
	c := l.ix.Class(k, l.ver)
	if c == nil {
		return nil
	}
	l.visited++
	l.path[l.depth] = k
	l.depth++
	defer func() { l.depth-- }()
	if c.Kind != syntax.KindTrait {
		owner = c.FQN
	}
	if p := c.Props[name]; p != nil {
		if c.Kind == syntax.KindTrait && owner != "" {
			copy := *p
			copy.TypeClass = owner
			return &copy
		}
		return p
	}
	for _, trait := range c.Traits {
		if p := l.find(trait, name, owner); p != nil {
			return p
		}
	}
	if c.Parent != "" {
		if p := l.find(c.Parent, name, ""); p != nil {
			return p
		}
	}
	for _, iface := range c.Interfaces {
		if p := l.find(iface, name, ""); p != nil {
			return p
		}
	}
	return nil
}
