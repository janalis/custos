package index

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
)

// World says how much of the code a project can load its index holds. The
// zero value knows nothing: a name the index lacks may still be declared by
// a file it never saw, so no check may treat absence as proof.
//
// A complete world (set by the project layer from Composer's metadata, see
// project.ComposerWorld) holds every class-like and function that the
// installed packages and the project's own autoload configuration can
// declare. Code the index never saw (files outside the analyzed paths,
// generated or downloaded at run time, PHP extensions the stubs lack) is
// still possible, which Blind narrows down.
type World struct {
	// Complete is set when the installed packages are the ones the
	// project requires and every one of them was indexed.
	Complete bool
	// Blind lists namespace prefixes (lower case, ending in a backslash)
	// the project autoloads from directories the index does not wholly
	// hold.
	Blind []string
	// BlindExact lists namespaces (lower case, ending in a backslash) of
	// files the index lacks. Such a file may declare anything once loaded,
	// in its namespace or below it, so Unseen hides those as well: silence
	// over a finding the index cannot support.
	BlindExact []string
	// Vendor is the directory Composer installs dependencies into (the
	// project's `config.vendor-dir`, `vendor` by default), absolute or
	// relative to the working directory (see Dependency); empty when no
	// project says.
	Vendor string
	// WorkDir is the working directory relative paths are written against,
	// so that Dependency resolves them without asking the system on every
	// call; empty to ask it.
	WorkDir string
	// Extensions holds the PHP extensions the project or an installed
	// package requires, suggests or requires for development (`ext-…`),
	// named as the stubs name them (see Function.Ext): the project may run
	// a build of one of them that is not PHP's own.
	Extensions map[string]bool
	// AutoloadedClasses lists, lower case, the class name prefixes a
	// package's autoloader may load from where the index does not look: the
	// namespace it registers in (ending in a backslash) and the prefixes it
	// tests the class name for (Rector's `rectorprefix`). An autoloader
	// declares classes only, so these hide no function (see UnseenClass).
	AutoloadedClasses []string
	// Hosts lists namespaces (lower case, ending in a backslash) of
	// packages whose autoloader loads a bundle of other code, such as
	// PHPStan's phar: a file importing from one runs within that package,
	// which may provide any name.
	Hosts []string
}

// Dependency reports whether file lies in the vendor directory. A file
// and a directory written in different forms (one absolute, the other
// relative to the working directory) are compared as absolute paths.
func (w World) Dependency(file string) bool {
	if w.Vendor == "" {
		return false
	}
	file, dir := filepath.Clean(file), filepath.Clean(w.Vendor)
	if filepath.IsAbs(file) != filepath.IsAbs(dir) {
		file, dir = w.absolute(file), w.absolute(dir)
	}
	return len(file) > len(dir) && file[len(dir)] == filepath.Separator && strings.HasPrefix(file, dir)
}

// absolute returns clean path made absolute against WorkDir, or else the
// current working directory; as is when that is unknown.
func (w World) absolute(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if w.WorkDir != "" {
		return filepath.Join(w.WorkDir, path)
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// SetWorld records what the index knows about the project's world. Call it
// before Freeze; layers over this index report the same world.
func (ix *Index) SetWorld(w World) {
	ix.mutable()
	ix.world = &w
}

// World returns the world recorded by this index or its nearest base layer
// that has one; the zero World when none does.
func (ix *Index) World() World {
	for l := ix; l != nil; l = l.base {
		if l.world != nil {
			return *l.world
		}
	}
	return World{}
}

// UnseenClass is Unseen for a class-like name, which a package's
// autoloader may also declare (see AutoloadedClasses).
func (w World) UnseenClass(fqn string) bool {
	k := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	for _, prefix := range w.AutoloadedClasses {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return w.Unseen(fqn)
}

// Unseen reports whether the class-like or function fqn lies in a namespace
// the project autoloads from a directory the index does not wholly hold:
// below a prefix of Blind or of BlindExact.
func (w World) Unseen(fqn string) bool {
	k := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	for _, prefix := range w.Blind {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	for _, namespace := range w.BlindExact {
		if strings.HasPrefix(k, namespace) {
			return true
		}
	}
	return false
}

// Symbols returns what the index holds of the file at path, in this layer
// or a base, nil when none indexes it. The result is shared: callers must
// not modify it.
func (ix *Index) Symbols(path string) *FileSymbols {
	for l := ix; l != nil; l = l.base {
		locked := l.rlock()
		fs := l.files[path]
		l.runlock(locked)
		if fs != nil {
			return fs
		}
	}
	return nil
}

type dynamicAliasesKey struct{}

// DynamicAliases reports whether an indexed file calls class_alias() with
// names the index cannot read, which may declare any class name.
func (ix *Index) DynamicAliases() bool {
	for l := ix; l != nil; l = l.base {
		found := l.Memo(dynamicAliasesKey{}, func() any {
			locked := l.rlock()
			defer l.runlock(locked)
			for _, fs := range l.files {
				if fs.DynamicAliases {
					return true
				}
			}
			return false
		}).(bool)
		if found {
			return true
		}
	}
	return false
}

// BuiltinPrefixes describes the names PHP extensions use, from the bottom
// layer (the embedded stubs).
type BuiltinPrefixes struct {
	// Namespaces holds the first namespace segment (lower case) of every
	// namespaced builtin class-like and function.
	Namespaces map[string]bool
	// Words holds the leading word, up to the first underscore (lower
	// case), of every global builtin function whose name has one.
	Words map[string]bool
	// ClassWords holds the leading word (see ClassWord) of every global
	// builtin class-like.
	ClassWords map[string]bool
}

type builtinPrefixesKey struct{}

// BuiltinPrefixes returns the name prefixes of the bottom layer.
func (ix *Index) BuiltinPrefixes() BuiltinPrefixes {
	b := ix
	for b.base != nil {
		b = b.base
	}
	return b.Memo(builtinPrefixesKey{}, func() any {
		p := BuiltinPrefixes{Namespaces: map[string]bool{}, Words: map[string]bool{}, ClassWords: map[string]bool{}}
		locked := b.rlock()
		defer b.runlock(locked)
		for k, c := range b.classes {
			if ns, _, ok := strings.Cut(k, `\`); ok {
				p.Namespaces[ns] = true
			} else {
				for _, decl := range c {
					p.ClassWords[ClassWord(decl.FQN)] = true
				}
			}
		}
		for k := range b.functions {
			if ns, _, ok := strings.Cut(k, `\`); ok {
				p.Namespaces[ns] = true
			} else if word, _, ok := strings.Cut(k, "_"); ok {
				p.Words[word] = true
			}
		}
		return p
	}).(BuiltinPrefixes)
}

// ClassWord returns, lower case, the leading word of a global class name:
// up to its first underscore or the start of its second word in camel case,
// an acronym counting as one word (`Spl` of SplObjectStorage, `pdo` of
// PDOStatement, `date` of DateTime).
func ClassWord(name string) string {
	upper := func(c byte) bool { return 'A' <= c && c <= 'Z' }
	for i := 1; i < len(name); i++ {
		switch c := name[i]; {
		case c == '_':
			return strings.ToLower(name[:i])
		case !upper(c):
		case !upper(name[i-1]):
			return strings.ToLower(name[:i])
		case i+1 < len(name) && !upper(name[i+1]) && name[i+1] != '_':
			// The last capital of an acronym starts the next word.
			return strings.ToLower(name[:i])
		}
	}
	return strings.ToLower(name)
}

// LoaderFacts returns what the callable k, an AutoloadCallbacks entry,
// would load as an autoloader (its Loader facts): those of the function,
// or of the method with a body the class resolves the name to (its own,
// a trait's or a parent's; not a @method tag). A class the code names only
// when it runs (`$this`, `static`, `self` or `parent` in a trait) stands
// for each class it may be, whose facts add up. Nothing when the index
// does not declare the callable.
func (ix *Index) LoaderFacts(k string) *Loader {
	class, method, ok := strings.Cut(k, "::")
	if !ok {
		for alt := range strings.SplitSeq(k, "|") {
			if f := ix.Function(alt, 0); f != nil {
				return ix.loader(f.File, alt)
			}
		}
		return nil
	}
	var out *Loader
	shared := true // out is still a file's own Loader
	for _, c := range ix.loaderClasses(class) {
		l := ix.methodLoader(c, method)
		switch {
		case l == nil:
		case out == nil:
			out = l
		default:
			if shared {
				out = &Loader{Prefixes: slices.Clone(out.Prefixes), Phar: out.Phar}
				shared = false
			}
			for _, p := range l.Prefixes {
				if len(out.Prefixes) < maxAutoloadPrefixes && !slices.Contains(out.Prefixes, p) {
					out.Prefixes = append(out.Prefixes, p)
				}
			}
			out.Phar = out.Phar || l.Phar
		}
	}
	return out
}

// maxLoaderClasses caps the classes a callback's class may stand for.
const maxLoaderClasses = 64

// loaderClasses returns the classes the class part of a callback key
// stands for (see lateMark).
func (ix *Index) loaderClasses(class string) []string {
	name := strings.TrimLeft(class, lateMark+usersMark+parentMark)
	var roots []string
	switch class[:len(class)-len(name)] {
	case "":
		return []string{name}
	case usersMark:
		return ix.traitUsers(name)
	case parentMark:
		for _, u := range ix.traitUsers(name) {
			if c := ix.Class(u, 0); c != nil && c.Parent != "" && !slices.Contains(roots, c.Parent) {
				roots = append(roots, c.Parent)
			}
		}
		return roots
	}
	if c := ix.Class(name, 0); c != nil && c.Kind == syntax.KindTrait {
		roots = ix.traitUsers(name)
	} else {
		roots = []string{name}
	}
	out := slices.Clone(roots)
	for _, r := range roots {
		for _, sub := range ix.Subclasses(r) {
			if len(out) >= maxLoaderClasses {
				return out
			}
			if !slices.Contains(out, sub) {
				out = append(out, sub)
			}
		}
	}
	return out
}

// methodLoader returns the Loader facts of the method with a body class
// resolves name to (nil for none, or a @method tag).
func (ix *Index) methodLoader(class, name string) *Loader {
	m := ix.FindMethod(class, name, 0)
	var c *Class // the declaring class
	if m != nil {
		c = ix.Class(m.Class, 0)
	}
	if c == nil || m.Magic || m.Abstract {
		return nil
	}
	if m.Declared != "" { // imported under a trait alias
		name = key(m.Declared)
	}
	return ix.loader(c.File, key(c.FQN)+"::"+name)
}

// traitUsers returns the classes and enums using trait, directly or
// through other traits, at most maxLoaderClasses.
func (ix *Index) traitUsers(trait string) []string {
	var out []string
	seen := map[string]bool{key(trait): true}
	for queue := []string{trait}; len(queue) > 0; queue = queue[1:] {
		for l := ix; l != nil; l = l.base {
			for _, u := range l.table(&l.users, l.buildTraitUsers)[key(queue[0])] {
				// The declaration the index serves decides: another one, or
				// one a layer above replaced, may use other traits.
				c := ix.Class(u, 0)
				if seen[key(u)] || len(out) >= maxLoaderClasses || c == nil || !slices.ContainsFunc(c.Traits, func(t string) bool { return key(t) == key(queue[0]) }) {
					continue
				}
				seen[key(u)] = true
				if c.Kind == syntax.KindTrait {
					queue = append(queue, u)
				} else {
					out = append(out, u)
				}
			}
		}
	}
	return out
}

// buildTraitUsers maps each trait, folded, to the class-likes of this
// layer using it, sorted.
func (ix *Index) buildTraitUsers() map[string][]string {
	users := map[string][]string{}
	for _, cs := range ix.classes {
		for _, c := range cs {
			for _, t := range c.Traits {
				users[key(t)] = append(users[key(t)], c.FQN)
			}
		}
	}
	for _, list := range users {
		slices.Sort(list)
	}
	return users
}

// loader returns the Loader facts of callable k declared in the file at
// path (none when the file has no loader, or was removed since the
// declaration was found).
func (ix *Index) loader(path, k string) *Loader {
	fs := ix.Symbols(path)
	if fs == nil || fs.Loaders == nil {
		return nil
	}
	return fs.Loaders[k]
}
