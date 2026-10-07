package syntax

// slabChunk is the number of nodes allocated at once per node type.
const slabChunk = 64

// slab hands out pointers into chunked backing arrays, turning one heap
// allocation per node into one per slabChunk nodes. Nodes stay valid for the
// lifetime of the File (chunks are never reused).
type slab[T any] struct{ buf []T }

func (s *slab[T]) next() *T {
	if len(s.buf) == 0 {
		s.buf = make([]T, slabChunk)
	}
	x := &s.buf[0]
	s.buf = s.buf[1:]
	return x
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
