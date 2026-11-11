package semanticquery

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// MethodOverrides is the bounded set of method names declared by descendants.
// Incomplete enumeration cannot establish the absence of an override.
type MethodOverrides struct {
	names    map[string]bool
	declared map[string][]*index.Method
	Complete bool
}

// MayOverride reports a known override or an incomplete hierarchy.
func (m MethodOverrides) MayOverride(name string) bool {
	return !m.Complete || m.names[strings.ToLower(name)]
}

// Declared returns the methods named name that descendants declare
// themselves (composed trait methods aside).
func (m MethodOverrides) Declared(name string) []*index.Method {
	return m.declared[strings.ToLower(name)]
}

// DescendantMethods follows classes and interfaces, including methods composed
// from traits and their aliases. A frozen index computes it once per class
// and version; callers must not modify the result.
func DescendantMethods(ix *index.Index, fqn string, ver phpversion.Version) MethodOverrides {
	key := descendantMethodsKey{fqn: strings.ToLower(strings.TrimPrefix(fqn, `\`)), ver: ver}
	return ix.Memo(key, func() any { return descendantMethods(ix, fqn, ver, 65536) }).(MethodOverrides)
}

// descendantMethodsKey keys DescendantMethods in the index memo: the
// enumeration only depends on the declarations of the index.
type descendantMethodsKey struct {
	fqn string
	ver phpversion.Version
}

func descendantMethods(ix *index.Index, fqn string, ver phpversion.Version, budget int) MethodOverrides {
	out := MethodOverrides{names: map[string]bool{}, declared: map[string][]*index.Method{}, Complete: true}
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
					out.declared[name] = append(out.declared[name], method)
				}
			}
			if !ix.AncestorsComplete(child, ver) {
				out.Complete = false
			}
			composed := composedTraits(ix, c, ver)
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
					own := composed[strings.ToLower(strings.TrimPrefix(ancestor.FQN, `\`))]
					for name := range ancestor.Methods {
						if !charge() {
							return out
						}
						// A method of a trait the child does not compose
						// itself is owned by the class that does, unless
						// one the child composes declares it too (even by
						// an @method tag): looked up there.
						if !own {
							continue
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

// composedTraits returns the traits c composes, directly or through other
// traits, lower-case and without a leading backslash.
func composedTraits(ix *index.Index, c *index.Class, ver phpversion.Version) map[string]bool {
	composed := map[string]bool{}
	queue := slices.Clone(c.Traits)
	for len(queue) > 0 {
		name := strings.ToLower(strings.TrimPrefix(queue[0], `\`))
		queue = queue[1:]
		if composed[name] {
			continue
		}
		composed[name] = true
		if trait := ix.Class(name, ver); trait != nil {
			queue = append(queue, trait.Traits...)
		}
	}
	return composed
}
