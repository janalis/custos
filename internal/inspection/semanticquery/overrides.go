package semanticquery

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// MethodOverrides is the bounded set of method names declared by descendants.
// Incomplete enumeration cannot establish the absence of an override.
type MethodOverrides struct {
	names    map[string]bool
	Complete bool
}

// MayOverride reports a known override or an incomplete hierarchy.
func (m MethodOverrides) MayOverride(name string) bool {
	return !m.Complete || m.names[strings.ToLower(name)]
}

// DescendantMethods follows classes and interfaces, including methods composed
// from traits and their aliases. Callers should memoize this per base class.
func DescendantMethods(ix *index.Index, fqn string, ver phpversion.Version) MethodOverrides {
	return descendantMethods(ix, fqn, ver, 65536)
}

func descendantMethods(ix *index.Index, fqn string, ver phpversion.Version, budget int) MethodOverrides {
	out := MethodOverrides{names: map[string]bool{}, Complete: true}
	seen := map[string]bool{}
	queue := []string{fqn}
	charge := func() bool {
		budget--
		if budget < 0 {
			out.Complete = false
			return false
		}
		return true
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range ix.ChildrenAll(cur) {
			if !charge() {
				return out
			}
			key := strings.ToLower(strings.TrimPrefix(child, `\`))
			if seen[key] {
				continue
			}
			seen[key] = true
			queue = append(queue, child)
			c := ix.Class(child, ver)
			if c == nil || ver != 0 && !c.Avail.In(ver) {
				out.Complete = false
				continue
			}
			if ix.ClassCount(child) != 1 {
				out.Complete = false
			}
			for name, method := range c.Methods {
				if !charge() {
					return out
				}
				if ver == 0 || method.Avail.In(ver) {
					out.names[name] = true
				}
			}
			if !ix.AncestorsComplete(child, ver) {
				out.Complete = false
			}
			for _, ancestor := range ix.Ancestors(child, ver) {
				if !charge() {
					return out
				}
				dependencies := append([]string{ancestor.Parent}, ancestor.Interfaces...)
				dependencies = append(dependencies, ancestor.Traits...)
				for _, dependency := range dependencies {
					if !charge() {
						return out
					}
					if dependency != "" {
						declared := ix.Class(dependency, ver)
						if declared == nil || ver != 0 && !declared.Avail.In(ver) {
							out.Complete = false
						}
					}
				}
				if ancestor.Kind == syntax.KindTrait {
					for name := range ancestor.Methods {
						if !charge() {
							return out
						}
						if method := ix.FindMethod(child, name, ver); method != nil && strings.EqualFold(method.TypeClass, c.FQN) {
							out.names[name] = true
						}
					}
				}
				for _, adaptation := range ancestor.TraitAdaptations {
					if !charge() {
						return out
					}
					if adaptation.Alias != "" {
						name := strings.ToLower(adaptation.Alias)
						if method := ix.FindMethod(child, name, ver); method != nil && strings.EqualFold(method.TypeClass, c.FQN) {
							out.names[name] = true
						}
					}
				}
			}
		}
	}
	return out
}
