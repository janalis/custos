package semanticquery

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// descendantMethodsLookingUpAll is descendantMethods looking up every
// method of every trait the descendants inherit: the reference the faster
// walk must match.
func descendantMethodsLookingUpAll(ix *index.Index, fqn string, ver phpversion.Version, budget int) MethodOverrides {
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

// Looking up only the methods of the traits a descendant composes, and
// those its traits name by @method, finds what looking up every inherited
// trait method finds, and runs out of budget at the same point. A project given by
// CUSTOS_ORACLE_CORPUS (its src and vendor directories) is checked over
// every class instead of the fixture.
func TestDescendantMethodsMatchesLookingUpAll(t *testing.T) {
	ix := index.New(nil)
	var classes []string
	add := func(fs *index.FileSymbols) {
		ix.Add(fs)
		for _, c := range fs.Classes {
			classes = append(classes, c.FQN)
		}
	}
	if root := os.Getenv("CUSTOS_ORACLE_CORPUS"); root != "" {
		for _, dir := range []string{"src", "vendor"} {
			_ = filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() && strings.HasSuffix(path, ".php") {
					if src, err := os.ReadFile(path); err == nil {
						add(index.Extract(syntax.Parse(path, src, syntax.Options{Version: phpversion.PHP84})))
					}
				}
				return nil
			})
		}
	} else {
		add(index.Extract(syntax.Parse("traits.php", []byte(`<?php
trait Inherited { function run(){} function late(){} function shared(){} }
/** @method void late() */
trait Tagged { use Deep; }
trait Deep { function shared(){} function deep(){} }
trait Own { function own(){} use Tagged { deep as renamed; } }
class Base { use Inherited; function base(){} }
class Child extends Base { use Own; }
class GrandChild extends Child { use Inherited; }
class Diamond extends Base { use Tagged, Deep; }
trait Real { function aliased(){} }
class_alias('Real', 'Alias');
class Aliasing extends Base { use Alias; }
`), syntax.Options{})))
		// Unavailable before 8.0, late() leaves the child's @method tag
		// to answer for it.
		ix.Class("Inherited", 0).Methods["late"].Avail.From = phpversion.PHP80
	}
	ix.Freeze()
	checked := 0
	for _, fqn := range classes {
		for _, ver := range []phpversion.Version{0, phpversion.PHP74, phpversion.PHP84} {
			for _, budget := range []int{65536, 40} {
				got, want := descendantMethods(ix, fqn, ver, budget), descendantMethodsLookingUpAll(ix, fqn, ver, budget)
				// A walk the budget cuts short is incomplete, whatever names
				// it found (MayOverride then holds for every name).
				if got.Complete != want.Complete || got.Complete && !reflect.DeepEqual(got, want) {
					t.Fatalf("%s at %v, budget %d: %+v, want %+v", fqn, ver, budget, got.names, want.names)
				}
				checked++
			}
		}
	}
	t.Logf("%d hierarchies checked", checked)
}
