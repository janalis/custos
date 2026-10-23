package index

import (
	"regexp"
	"strings"

	phpversion "custos/internal/php/version"
)

// maxAliasDepth bounds the imports a type alias is followed through, as a
// cycle of them would be.
const maxAliasDepth = 8

// aliasMarkRef matches the imported aliases left in a recorded type.
var aliasMarkRef = regexp.MustCompile(`\\` + AliasMark + `([A-Za-z_][A-Za-z0-9_]*)`)

// DeferTypeAliases makes Add leave the type aliases the classes of a file
// import unlinked, until LinkTypeAliases links those of every file: a
// whole project is added in any order, so the class an alias comes from
// may be added after the class importing it. Readers must not use the
// index meanwhile.
func (ix *Index) DeferTypeAliases() { ix.deferAliases.Store(true) }

// LinkTypeAliases links the type aliases the classes of every file import
// (see DeferTypeAliases), and makes Add link those of each file it adds.
func (ix *Index) LinkTypeAliases() {
	locked := ix.rlock()
	files := make([]*FileSymbols, 0, len(ix.files))
	for _, fs := range ix.files {
		files = append(files, fs)
	}
	ix.runlock(locked)
	for _, fs := range files {
		ix.linkAliases(fs)
	}
	ix.deferAliases.Store(false)
}

// linkAliases replaces, in the types recorded for the members of the
// classes of fs, each type alias the class imports by its definition
// (mixed when it cannot be found), reading the other classes of fs before
// those of the index.
func (ix *Index) linkAliases(fs *FileSymbols) {
	for _, c := range fs.Classes {
		if len(c.TypeImports) == 0 {
			continue
		}
		link := func(s *string) {
			if strings.Contains(*s, AliasMark) {
				*s = ix.expandAliases(fs, c.FQN, *s, 0)
			}
		}
		for _, m := range c.Methods {
			link(&m.DocReturn)
			link(&m.CondReturn)
			link(&m.GenReturn)
			link(&m.Inferred)
			for i := range m.Params {
				link(&m.Params[i].DocType)
				link(&m.Params[i].Out)
				link(&m.Params[i].Gen)
			}
			if m.Tpl != nil {
				link(&m.Tpl.Return)
				for i := range m.Tpl.Params {
					link(&m.Tpl.Params[i])
				}
			}
		}
		for _, p := range c.Props {
			link(&p.DocType)
			link(&p.Inferred)
		}
	}
}

// TypeAlias returns type alias name of class cls as a doc type: the one it
// declares, with the aliases it imports expanded, or the one it imports;
// ok is false when neither is known.
func (ix *Index) TypeAlias(cls, name string) (string, bool) {
	return ix.typeAlias(nil, cls, name, 0)
}

// typeAlias is TypeAlias reading the classes of fs (when not nil) first.
func (ix *Index) typeAlias(fs *FileSymbols, cls, name string, depth int) (string, bool) {
	c := ix.aliasClass(fs, cls)
	if c == nil || depth > maxAliasDepth {
		return "", false
	}
	if im, ok := c.TypeImports[name]; ok {
		return ix.typeAlias(fs, im.From, im.Name, depth+1)
	}
	def, ok := c.TypeAliases[name]
	if !ok {
		return "", false
	}
	return ix.expandAliases(fs, cls, def, depth), true
}

// expandAliases replaces the imported aliases of class cls left in type
// doc by their definitions, mixed when unknown.
func (ix *Index) expandAliases(fs *FileSymbols, cls, doc string, depth int) string {
	return aliasMarkRef.ReplaceAllStringFunc(doc, func(ref string) string {
		def, ok := ix.typeAlias(fs, cls, aliasMarkRef.FindStringSubmatch(ref)[1], depth+1)
		if !ok {
			return "mixed"
		}
		return "(" + def + ")"
	})
}

// aliasClass returns class cls of fs, else of the index.
func (ix *Index) aliasClass(fs *FileSymbols, cls string) *Class {
	if fs != nil {
		for _, c := range fs.Classes {
			if key(c.FQN) == key(cls) {
				return c
			}
		}
	}
	return ix.Class(cls, phpversion.Max)
}
