package syntax

import "sync"

// arena is the memory of one parsed tree that File.Release can hand to a
// later parse: the node slabs and the token stream. Dependency indexing
// parses every file only to read its declarations, so recycling it removes
// most of the parser's allocations there.
type arena struct {
	slabs slabs
	toks  []Token // a released token buffer, until the next parse takes it
}

// arenaPool holds the arenas of released trees.
var arenaPool sync.Pool // *arena

// sigPool holds the parser's significant-token index, which never outlives
// a parse.
var sigPool sync.Pool // *[]int32

func getArena() *arena {
	if ar, ok := arenaPool.Get().(*arena); ok {
		return ar
	}
	return new(arena)
}

// sigBuffer returns an empty index buffer with room for n entries.
func sigBuffer(n int) []int32 {
	if p, ok := sigPool.Get().(*[]int32); ok && cap(*p) >= n {
		return (*p)[:0]
	}
	return make([]int32, 0, n)
}

func putSig(sig []int32) {
	sigPool.Put(&sig)
}

// Release hands the memory of a tree nobody uses any more to later parses,
// which overwrite it: afterwards f has no statements and no tokens, and no
// node, token slice or parent link taken from f may be read again. Strings
// and the source stay valid. Call it only on a tree being discarded, at most
// once.
func (f *File) Release() {
	ar := f.arena
	if ar == nil {
		return
	}
	f.arena = nil
	ar.slabs.recycle()
	ar.toks = f.Tokens[:0]
	f.Tokens, f.Stmts = nil, nil
	arenaPool.Put(ar)
}
