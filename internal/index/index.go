package index

import (
	"strings"
	"sync"

	"custos/internal/phpver"
)

// Index holds symbols keyed by lower-case FQN (constants: exact case, with
// a lower-case fallback). It is safe for concurrent reads; Add/Remove take a
// write lock. A base index (the embedded stubs) is consulted after the
// project symbols.
type Index struct {
	mu        sync.RWMutex
	base      *Index
	classes   map[string][]*Class
	functions map[string][]*Function
	constants map[string][]*Constant
	files     map[string]*FileSymbols
	children  map[string][]string // lower parent/interface FQN -> child FQNs (lazy)
}

// New returns an empty index layered over base (may be nil).
func New(base *Index) *Index {
	return &Index{base: base, classes: map[string][]*Class{}, functions: map[string][]*Function{},
		constants: map[string][]*Constant{}, files: map[string]*FileSymbols{}}
}

func key(fqn string) string { return strings.ToLower(strings.TrimPrefix(fqn, `\`)) }

// Add inserts (or replaces) the symbols of one file.
func (ix *Index) Add(fs *FileSymbols) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.removeLocked(fs.Path)
	ix.files[fs.Path] = fs
	for _, c := range fs.Classes {
		k := key(c.FQN)
		ix.classes[k] = append(ix.classes[k], c)
	}
	for _, f := range fs.Functions {
		k := key(f.FQN)
		ix.functions[k] = append(ix.functions[k], f)
	}
	for _, c := range fs.Constants {
		k := strings.TrimPrefix(c.FQN, `\`)
		ix.constants[k] = append(ix.constants[k], c)
	}
	ix.children = nil
}

// Remove drops the symbols of a file.
func (ix *Index) Remove(path string) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.removeLocked(path)
	ix.children = nil
}

func (ix *Index) removeLocked(path string) {
	old, ok := ix.files[path]
	if !ok {
		return
	}
	delete(ix.files, path)
	for _, c := range old.Classes {
		ix.classes[key(c.FQN)] = without(ix.classes[key(c.FQN)], c)
	}
	for _, f := range old.Functions {
		ix.functions[key(f.FQN)] = without(ix.functions[key(f.FQN)], f)
	}
	for _, c := range old.Constants {
		k := strings.TrimPrefix(c.FQN, `\`)
		ix.constants[k] = without(ix.constants[k], c)
	}
}

func without[T comparable](list []T, x T) []T {
	out := list[:0]
	for _, v := range list {
		if v != x {
			out = append(out, v)
		}
	}
	return out
}

// pick returns the first candidate available at ver (or the first one).
func pick[T any](cands []T, avail func(T) Avail, ver phpver.Version) (T, bool) {
	var zero T
	if len(cands) == 0 {
		return zero, false
	}
	if ver != 0 {
		for _, c := range cands {
			if avail(c).In(ver) {
				return c, true
			}
		}
	}
	return cands[0], true
}

// Class returns the class-like with the given FQN available at ver (0 = any).
func (ix *Index) Class(fqn string, ver phpver.Version) *Class {
	ix.mu.RLock()
	c, ok := pick(ix.classes[key(fqn)], func(c *Class) Avail { return c.Avail }, ver)
	ix.mu.RUnlock()
	if ok {
		return c
	}
	if ix.base != nil {
		return ix.base.Class(fqn, ver)
	}
	return nil
}

// Function returns the function with the given FQN available at ver.
func (ix *Index) Function(fqn string, ver phpver.Version) *Function {
	ix.mu.RLock()
	f, ok := pick(ix.functions[key(fqn)], func(f *Function) Avail { return f.Avail }, ver)
	ix.mu.RUnlock()
	if ok {
		return f
	}
	if ix.base != nil {
		return ix.base.Function(fqn, ver)
	}
	return nil
}

// Constant returns a global constant (exact case first, then case-insensitive
// for the namespace part).
func (ix *Index) Constant(fqn string, ver phpver.Version) *Constant {
	fqn = strings.TrimPrefix(fqn, `\`)
	ix.mu.RLock()
	c, ok := pick(ix.constants[fqn], func(c *Constant) Avail { return c.Avail }, ver)
	ix.mu.RUnlock()
	if ok {
		return c
	}
	if ix.base != nil {
		return ix.base.Constant(fqn, ver)
	}
	return nil
}

// ResolveFunction resolves a call target: the namespaced candidate when it
// exists, else the global fallback (PHP runtime semantics).
func (ix *Index) ResolveFunction(fqn, fallback string, ver phpver.Version) *Function {
	if f := ix.Function(fqn, ver); f != nil {
		return f
	}
	if fallback != "" {
		return ix.Function(fallback, ver)
	}
	return nil
}

// Ancestors returns the class, its parents, traits and interfaces
// (breadth-first, each once, cycle-safe). The class itself comes first.
func (ix *Index) Ancestors(fqn string, ver phpver.Version) []*Class {
	var out []*Class
	seen := map[string]bool{}
	queue := []string{fqn}
	for len(queue) > 0 {
		k := key(queue[0])
		queue = queue[1:]
		if seen[k] {
			continue
		}
		seen[k] = true
		c := ix.Class(k, ver)
		if c == nil {
			continue
		}
		out = append(out, c)
		queue = append(queue, c.Traits...)
		if c.Parent != "" {
			queue = append(queue, c.Parent)
		}
		queue = append(queue, c.Interfaces...)
	}
	return out
}

// ParentChain returns the parent classes (nearest first), cycle-safe.
func (ix *Index) ParentChain(fqn string, ver phpver.Version) []*Class {
	var out []*Class
	seen := map[string]bool{key(fqn): true}
	c := ix.Class(fqn, ver)
	for c != nil && c.Parent != "" && !seen[key(c.Parent)] {
		seen[key(c.Parent)] = true
		c = ix.Class(c.Parent, ver)
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}

// IsSubtype reports whether class `child` is, extends or implements `parent`.
func (ix *Index) IsSubtype(child, parent string, ver phpver.Version) bool {
	pk := key(parent)
	for _, c := range ix.Ancestors(child, ver) {
		if key(c.FQN) == pk {
			return true
		}
	}
	return false
}

// FindMethod looks a method up through the class, its traits, parents and
// interfaces. Returns nil when not found.
func (ix *Index) FindMethod(class, name string, ver phpver.Version) *Method {
	lname := strings.ToLower(name)
	for _, c := range ix.Ancestors(class, ver) {
		if m, ok := c.Methods[lname]; ok && (ver == 0 || m.Avail.In(ver)) {
			return m
		}
	}
	return nil
}

// FindProperty looks a property up through the class hierarchy.
func (ix *Index) FindProperty(class, name string, ver phpver.Version) *Property {
	for _, c := range ix.Ancestors(class, ver) {
		if p, ok := c.Props[name]; ok {
			return p
		}
	}
	return nil
}

// FindConst looks a class constant (or enum case) up through the hierarchy.
func (ix *Index) FindConst(class, name string, ver phpver.Version) *ClassConst {
	for _, c := range ix.Ancestors(class, ver) {
		if k, ok := c.Consts[name]; ok {
			return k
		}
	}
	return nil
}

// Children returns the FQNs of classes directly extending/implementing fqn
// (project symbols only).
func (ix *Index) Children(fqn string) []string {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	if ix.children == nil {
		ix.children = map[string][]string{}
		for _, cs := range ix.classes {
			for _, c := range cs {
				if c.Parent != "" {
					ix.children[key(c.Parent)] = append(ix.children[key(c.Parent)], c.FQN)
				}
				for _, i := range c.Interfaces {
					ix.children[key(i)] = append(ix.children[key(i)], c.FQN)
				}
			}
		}
	}
	return ix.children[key(fqn)]
}

// Stats returns symbol counts (project layer only).
func (ix *Index) Stats() (files, classes, functions, constants int) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	for _, cs := range ix.classes {
		classes += len(cs)
	}
	for _, fs := range ix.functions {
		functions += len(fs)
	}
	for _, cs := range ix.constants {
		constants += len(cs)
	}
	return len(ix.files), classes, functions, constants
}

// DropStaleInferred clears the body-inferred return types of fs when the
// index declares one of its ReturnDeps (a namespaced function the inference
// took for the global one). Call it before fs is shared with readers.
func (ix *Index) DropStaleInferred(fs *FileSymbols) {
	if len(fs.ReturnDeps) == 0 {
		return
	}
	stale := false
	ix.mu.RLock()
	for _, d := range fs.ReturnDeps {
		if len(ix.functions[key(d)]) > 0 {
			stale = true
			break
		}
	}
	ix.mu.RUnlock()
	if !stale {
		return
	}
	for _, f := range fs.Functions {
		f.Inferred = ""
	}
	for _, c := range fs.Classes {
		for _, m := range c.Methods {
			m.Inferred = ""
		}
	}
}
