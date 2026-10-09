package index

import (
	"strings"
	"sync"
	"sync/atomic"

	"custos/internal/phpver"
	"custos/internal/syntax"
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
	aliases   map[string]string   // lower class_alias() alias -> original FQN
	children  map[string][]string // lower parent/interface FQN -> child FQNs (lazy)

	// gen counts the changes (Add/Remove) of this layer; ancestors caches
	// Ancestors per class and version, valid while the generation of this
	// layer and of every base layer is unchanged (see generation).
	gen       atomic.Uint64
	ancMu     sync.RWMutex
	ancestors map[ancKey]ancEntry
}

type ancKey struct {
	fqn string // lower-case
	ver phpver.Version
}

type ancEntry struct {
	gen      uint64
	cls      []*Class
	complete bool // false when MaxAncestors cut the list
}

// MaxAncestors caps the classes Ancestors returns (the class itself, its
// traits, parents and interfaces, breadth-first). Real hierarchies stay
// far below; a hostile chain of thousands of classes would otherwise make
// every member lookup linear in its depth.
const MaxAncestors = 256

// generation is the sum of the change counters of this layer and its base
// layers: it changes whenever a layer the lookups read changes.
func (ix *Index) generation() uint64 {
	g := ix.gen.Load()
	if ix.base != nil {
		g += ix.base.generation()
	}
	return g
}

// New returns an empty index layered over base (may be nil).
func New(base *Index) *Index {
	return &Index{
		base: base, classes: map[string][]*Class{}, functions: map[string][]*Function{},
		constants: map[string][]*Constant{}, files: map[string]*FileSymbols{},
	}
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
	for _, a := range fs.ClassAliases {
		if ix.aliases == nil {
			ix.aliases = map[string]string{}
		}
		ix.aliases[key(a[0])] = a[1]
	}
	ix.children = nil
	ix.gen.Add(1)
}

// Remove drops the symbols of a file.
func (ix *Index) Remove(path string) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.removeLocked(path)
	ix.children = nil
	ix.gen.Add(1)
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
	for _, a := range old.ClassAliases {
		delete(ix.aliases, key(a[0]))
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

// maxAliasHops bounds the class_alias() chain Class follows (an alias of an
// alias…; a cycle ends there too).
const maxAliasHops = 8

// Class returns the class-like with the given FQN available at ver (0 = any).
// A name declared by class_alias() resolves to the original class.
func (ix *Index) Class(fqn string, ver phpver.Version) *Class {
	for hop := 0; hop <= maxAliasHops; hop++ {
		if c := ix.declared(fqn, ver); c != nil {
			return c
		}
		if fqn = ix.aliasOf(fqn); fqn == "" {
			return nil
		}
	}
	return nil
}

// aliasOf returns the original class of a class_alias() alias, in this
// layer or a base ("" when fqn is no alias).
func (ix *Index) aliasOf(fqn string) string {
	for l := ix; l != nil; l = l.base {
		l.mu.RLock()
		t := l.aliases[key(fqn)]
		l.mu.RUnlock()
		if t != "" {
			return t
		}
	}
	return ""
}

// declared is Class without aliases.
func (ix *Index) declared(fqn string, ver phpver.Version) *Class {
	ix.mu.RLock()
	c, ok := pick(ix.classes[key(fqn)], func(c *Class) Avail { return c.Avail }, ver)
	ix.mu.RUnlock()
	if ok {
		return c
	}
	if ix.base != nil {
		return ix.base.declared(fqn, ver)
	}
	return nil
}

// Function returns the function with the given FQN available at ver.
func (ix *Index) Function(fqn string, ver phpver.Version) *Function {
	ix.mu.RLock()
	f, ok := pick(ix.functions[key(fqn)], func(f *Function) Avail { return f.Avail }, ver)
	ix.mu.RUnlock()
	if ok {
		// PHP cannot redeclare a built-in function: a source declaration of
		// a name the stubs know at this version is a polyfill (guarded by
		// function_exists), so the built-in is what runs.
		if ix.base != nil {
			if b := ix.builtinFunction(fqn, ver); b != nil {
				return b
			}
		}
		return f.at(ver)
	}
	if ix.base != nil {
		return ix.base.Function(fqn, ver)
	}
	return nil
}

// FunctionDecls returns every declaration of function fqn available at ver
// in the first layer declaring it (several when a project declares it
// twice, e.g. a real and a no-op version loaded conditionally); nil when
// a builtin of that name exists (see Function) or none is declared.
func (ix *Index) FunctionDecls(fqn string, ver phpver.Version) []*Function {
	ix.mu.RLock()
	cands := ix.functions[key(fqn)]
	var out []*Function
	for _, f := range cands {
		if ver == 0 || f.Avail.In(ver) {
			out = append(out, f.at(ver))
		}
	}
	ix.mu.RUnlock()
	if len(cands) > 0 {
		if ix.base != nil && ix.builtinFunction(fqn, ver) != nil {
			return nil
		}
		return out
	}
	if ix.base != nil {
		return ix.base.FunctionDecls(fqn, ver)
	}
	return nil
}

// verKey identifies a version-resolved copy of a function or method.
type verKey struct {
	decl any // *Function or *Method
	ver  phpver.Version
}

// verCopies caches the copies of declarations whose return type differs at
// a PHP version (see VerType): the same copy is returned for one
// declaration and version, so identity comparisons keep working. Only
// stub declarations carry version maps, so the cache stays small.
var verCopies sync.Map

// at returns f with Return resolved for ver (f itself when unchanged).
func (f *Function) at(ver phpver.Version) *Function {
	if len(f.RetVer) == 0 {
		return f
	}
	r := returnAt(f.Return, f.RetVer, ver)
	if r == f.Return {
		return f
	}
	k := verKey{f, ver}
	if c, ok := verCopies.Load(k); ok {
		return c.(*Function)
	}
	cp := *f
	cp.Return, cp.RetVer = r, nil
	cp.DocReturn = docAt(f.DocReturn, f.Return, r)
	c, _ := verCopies.LoadOrStore(k, &cp)
	return c.(*Function)
}

// at returns m with Return resolved for ver (m itself when unchanged).
func (m *Method) at(ver phpver.Version) *Method {
	if len(m.RetVer) == 0 {
		return m
	}
	r := returnAt(m.Return, m.RetVer, ver)
	if r == m.Return {
		return m
	}
	k := verKey{m, ver}
	if c, ok := verCopies.Load(k); ok {
		return c.(*Method)
	}
	cp := *m
	cp.Return, cp.RetVer = r, nil
	cp.DocReturn = docAt(m.DocReturn, m.Return, r)
	c, _ := verCopies.LoadOrStore(k, &cp)
	return c.(*Method)
}

// docAt adds to the documented return type doc the members an older
// version's declared type old has beyond the newest one (newest): a doc
// `array` would otherwise hide `array_chunk()`'s pre-8.0 null.
func docAt(doc, newest, old string) string {
	if doc == "" {
		return doc
	}
	have := map[string]bool{}
	for _, a := range strings.Split(doc, "|") {
		have[strings.ToLower(a)] = true
	}
	for _, a := range strings.Split(newest, "|") {
		have[strings.ToLower(a)] = true
	}
	for _, a := range strings.Split(old, "|") {
		if !have[strings.ToLower(a)] {
			doc += "|" + a
			have[strings.ToLower(a)] = true
		}
	}
	return doc
}

// builtinFunction looks fqn up in the bottom layer (the embedded stubs),
// returning it only when available at ver.
func (ix *Index) builtinFunction(fqn string, ver phpver.Version) *Function {
	b := ix.base
	for b.base != nil {
		b = b.base
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, f := range b.functions[key(fqn)] {
		if ver == 0 || f.Avail.In(ver) { // strictly available: else the polyfill runs
			return f.at(ver)
		}
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
// At most MaxAncestors are returned. The result is cached (callers must not
// modify it).
func (ix *Index) Ancestors(fqn string, ver phpver.Version) []*Class {
	return ix.ancestorEntry(fqn, ver).cls
}

// AncestorsComplete reports whether Ancestors(fqn, ver) lists every
// ancestor: false when the MaxAncestors cap cut it (a class implementing
// hundreds of interfaces), so callers that need the whole hierarchy can
// stay conservative.
func (ix *Index) AncestorsComplete(fqn string, ver phpver.Version) bool {
	return ix.ancestorEntry(fqn, ver).complete
}

func (ix *Index) ancestorEntry(fqn string, ver phpver.Version) ancEntry {
	k := ancKey{key(fqn), ver}
	g := ix.generation()
	ix.ancMu.RLock()
	e, ok := ix.ancestors[k]
	ix.ancMu.RUnlock()
	if ok && e.gen == g {
		return e
	}
	out, complete := ix.computeAncestors(fqn, ver)
	e = ancEntry{gen: g, cls: out, complete: complete}
	ix.ancMu.Lock()
	if ix.ancestors == nil {
		ix.ancestors = map[ancKey]ancEntry{}
	}
	ix.ancestors[k] = e
	ix.ancMu.Unlock()
	return e
}

func (ix *Index) computeAncestors(fqn string, ver phpver.Version) ([]*Class, bool) {
	var out []*Class
	seen := map[string]bool{}
	queue := []string{fqn}
	for len(queue) > 0 && len(out) < MaxAncestors {
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
	// Cut when a class not yet listed remains to be visited.
	for _, q := range queue {
		if !seen[key(q)] && ix.Class(q, ver) != nil {
			return out, false
		}
	}
	return out, true
}

// ParentChain returns the parent classes (nearest first), cycle-safe.
func (ix *Index) ParentChain(fqn string, ver phpver.Version) []*Class {
	var out []*Class
	seen := map[string]bool{key(fqn): true}
	c := ix.Class(fqn, ver)
	for c != nil && c.Parent != "" && !seen[key(c.Parent)] && len(out) < MaxAncestors {
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
	if c := ix.Class(class, ver); c != nil {
		if m := c.Methods[lname]; m != nil && !m.Magic && (ver == 0 || m.Avail.In(ver)) {
			return m.at(ver)
		}
	}
	return ix.findComposedMethod(class, lname, ver)
}

func (ix *Index) findComposedMethod(class, name string, ver phpver.Version) *Method {
	lookup := methodLookup{ix: ix, ver: ver}
	m, conflict := lookup.find(class, name, true)
	if conflict {
		return nil
	}
	if m == nil {
		m = lookup.magic
	}
	if m != nil {
		return m.at(ver)
	}
	return nil
}

// FindProperty looks a property up through the class hierarchy.
func (ix *Index) FindProperty(class, name string, ver phpver.Version) *Property {
	for _, c := range ix.Ancestors(class, ver) {
		if c.Kind == syntax.KindTrait && key(c.FQN) != key(class) {
			l := propertyLookup{ix: ix, ver: ver}
			return l.find(class, name, "")
		}
		if p := c.Props[name]; p != nil {
			return p
		}
	}
	return nil
}

// FindConst looks a class constant (or enum case) up through declarations,
// traits, parents and interfaces, in that order. Imported trait constants
// carry TypeClass on transient copies. Ambiguous trait composition returns nil.
func (ix *Index) FindConst(class, name string, ver phpver.Version) *ClassConst {
	if c := ix.Class(class, ver); c != nil {
		if k := c.Consts[name]; k != nil {
			return k
		}
	}
	l := constantLookup{ix: ix, ver: ver}
	k, _ := l.find(class, name, "")
	return k
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
		for _, p := range c.Props {
			p.Inferred = ""
		}
	}
}
