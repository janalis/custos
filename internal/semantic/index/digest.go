package index

import (
	"hash/maphash"

	"custos/internal/php/syntax"
)

// digestSeed keys Digest for the life of the process: digests are compared
// in memory only, never saved.
var digestSeed = maphash.MakeSeed()

// Digest fingerprints a parsed file: its bytes and the PHP version it was
// parsed for. Two files with the same digest yield the same symbols.
func Digest(f *syntax.File) uint64 {
	var h maphash.Hash
	h.SetSeed(digestSeed)
	_, _ = h.Write(f.Src)
	_ = h.WriteByte(byte(f.Version))
	_ = h.WriteByte(byte(f.Version >> 8))
	return h.Sum64()
}

// Holds reports that this layer indexes path from a file of that digest and
// that no other file of the layer declares a class, function, constant or
// class alias of the same name. An analysis of the file may then read this
// index as it is: a layer of its own above it would resolve every name the
// same way, its own declarations being the only ones of their names here.
func (ix *Index) Holds(path string, digest uint64) bool {
	defer ix.runlock(ix.rlock())
	fs := ix.files[path]
	if fs == nil || fs.Digest == 0 || fs.Digest != digest {
		return false
	}
	for _, c := range fs.Classes {
		if len(ix.classes[key(c.FQN)]) != 1 {
			return false
		}
	}
	for _, f := range fs.Functions {
		if len(ix.functions[key(f.FQN)]) != 1 {
			return false
		}
	}
	for _, c := range fs.Constants {
		if len(ix.constants[c.indexKey()]) != 1 {
			return false
		}
	}
	for _, a := range fs.ClassAliases {
		if len(ix.aliases[key(a[0])]) != 1 {
			return false
		}
	}
	return true
}
