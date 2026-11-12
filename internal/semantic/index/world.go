package index

import (
	"slices"
	"strings"
)

// World says how much of the code a project can load its index holds. The
// zero value knows nothing: a name the index lacks may still be declared by
// a file it never saw, so no check may treat absence as proof.
//
// A complete world (set by the project layer from Composer's metadata, see
// project.ComposerWorld) holds every class-like and function that the
// installed packages and the project's own autoload configuration can
// declare. Code the index never saw (files outside the analysed paths,
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
	// BlindExact lists namespaces (lower case, ending in a backslash)
	// whose own members, not those of their sub-namespaces, a file the
	// index lacks may declare.
	BlindExact []string
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

// Unseen reports whether the class-like or function fqn lies in a namespace
// the project autoloads from a directory the index does not wholly hold:
// below a prefix of Blind, or directly in a namespace of BlindExact.
func (w World) Unseen(fqn string) bool {
	k := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	for _, prefix := range w.Blind {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return slices.Contains(w.BlindExact, k[:strings.LastIndexByte(k, '\\')+1])
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
