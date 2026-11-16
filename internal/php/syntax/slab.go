package syntax

// slabChunk caps the number of nodes allocated at once per node type.
const slabChunk = 64

// slab hands out pointers into chunked backing arrays, turning one heap
// allocation per node into one per chunk, growing to slabChunk. Nodes stay
// valid for the lifetime of the File; File.Release hands the chunks, cleared,
// to the next parse.
type slab[T any] struct {
	buf   []T
	chunk int
	used  [][]T // chunks handed out to this parse, the last one being buf
	spare [][]T // cleared chunks of a released tree, taken before allocating
}

func (s *slab[T]) next() *T {
	if len(s.buf) == 0 {
		if n := len(s.spare); n > 0 {
			s.buf, s.spare = s.spare[n-1], s.spare[:n-1]
		} else {
			s.chunk = min(slabChunk, max(8, s.chunk*2))
			s.buf = make([]T, s.chunk)
		}
		s.used = append(s.used, s.buf)
	}
	x := &s.buf[0]
	s.buf = s.buf[1:]
	return x
}

// recycle clears the nodes handed out (so spare chunks pin nothing and come
// back zeroed) and makes every chunk spare.
func (s *slab[T]) recycle() {
	if n := len(s.used); n > 0 {
		last := s.used[n-1]
		for _, c := range s.used[:n-1] {
			clear(c)
		}
		clear(last[:len(last)-len(s.buf)])
	}
	s.spare = append(s.spare, s.used...)
	s.used = s.used[:0]
	s.buf = nil
}

// put copies v into the next slab slot and returns its address.
func put[T any](s *slab[T], v T) *T {
	x := s.next()
	*x = v
	return x
}

// slabs holds per-type node slabs for one parse.
type slabs struct {
	sVariable        slab[Variable]
	sLiteral         slab[Literal]
	sIdentifier      slab[Identifier]
	sArrayDimFetch   slab[ArrayDimFetch]
	sArg             slab[Arg]
	sStaticCall      slab[StaticCall]
	sPropertyFetch   slab[PropertyFetch]
	sClassConstFetch slab[ClassConstFetch]
	sParen           slab[Paren]
	sName            slab[Name]
	sFuncCall        slab[FuncCall]
	sExprStmt        slab[ExprStmt]
	sAssign          slab[Assign]
	sStringPart      slab[StringPart]
	sReturn          slab[Return]
	sParam           slab[Param]
	sMethodCall      slab[MethodCall]
	sConstFetch      slab[ConstFetch]
	sBinary          slab[Binary]
	sArrayItem       slab[ArrayItem]
	sArgList         slab[ArgList]
	sBlock           slab[Block]
}

func (s *slabs) recycle() {
	s.sVariable.recycle()
	s.sLiteral.recycle()
	s.sIdentifier.recycle()
	s.sArrayDimFetch.recycle()
	s.sArg.recycle()
	s.sStaticCall.recycle()
	s.sPropertyFetch.recycle()
	s.sClassConstFetch.recycle()
	s.sParen.recycle()
	s.sName.recycle()
	s.sFuncCall.recycle()
	s.sExprStmt.recycle()
	s.sAssign.recycle()
	s.sStringPart.recycle()
	s.sReturn.recycle()
	s.sParam.recycle()
	s.sMethodCall.recycle()
	s.sConstFetch.recycle()
	s.sBinary.recycle()
	s.sArrayItem.recycle()
	s.sArgList.recycle()
	s.sBlock.recycle()
}
