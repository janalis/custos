package flow

import "custos/internal/php/syntax"

// valueStore keeps expression values in small, lazily allocated blocks. Hash
// slots contain bounded 32-bit indices; hints for unreachable expressions do not reserve
// their payloads. Callers receive value copies, never pointers into the store.
//
// Most expression values hold none of the references a Value rarely carries
// (see leanValue): those live in lean blocks, about half the size, and only
// the others in full blocks, whose widths grow from one.
type valueStore struct {
	slots map[syntax.NodeIdentity]uint32
	leans [][]leanValue
	fulls [][]Value
	// Parameter-read queries need the evaluated simple variables, not other node types.
	variables []*syntax.Variable
	// width is that of the lean blocks; leanNext and fullNext are the next
	// free entry of the last block of each kind.
	width, leanNext, fullNext, headers int
}

const valueBlockLimit = 4

// leanSlot marks the index of a value kept in a lean block.
const leanSlot uint32 = 1 << 31

// Even a narrow block's reserved stride fits within one bounded scope, below
// the lean mark.
const _ uint32 = leanSlot - 1 - MaxRetained*valueBlockLimit

func newValueStore(capacity int) valueStore {
	width := min(capacity, valueBlockLimit)
	if width == 0 {
		width = valueBlockLimit
	}
	return valueStore{slots: make(map[syntax.NodeIdentity]uint32, capacity), width: width, headers: min(64, max(1, (capacity+width-1)/width))}
}

func (s *valueStore) get(expr syntax.Expr) (Value, bool) {
	index, known := s.slots[syntax.Identity(expr)]
	if !known {
		return Value{}, false
	}
	if index&leanSlot != 0 {
		index &^= leanSlot
		return s.leans[index/valueBlockLimit][index%valueBlockLimit].value(), true
	}
	return s.fulls[index/valueBlockLimit][index%valueBlockLimit], true
}

func (s *valueStore) at(expr syntax.Expr) Value {
	value, _ := s.get(expr)
	return value
}

func (s *valueStore) set(expr syntax.Expr, value Value) {
	lean, isLean := leanOf(&value)
	index, known := s.slots[syntax.Identity(expr)]
	switch {
	case known && index&leanSlot == 0:
		s.fulls[index/valueBlockLimit][index%valueBlockLimit] = value
		return
	case known && isLean:
		index &^= leanSlot
		s.leans[index/valueBlockLimit][index%valueBlockLimit] = lean
		return
	}
	if s.slots == nil {
		s.slots = map[syntax.NodeIdentity]uint32{}
		s.width = valueBlockLimit
	}
	if !known {
		if variable, ok := expr.(*syntax.Variable); ok && variable.NameExpr == nil {
			s.variables = append(s.variables, variable)
		}
	}
	// Indices reserve a full block stride even when a small scope uses a
	// narrower block. Reads and updates can split them with constant shifts.
	// A lean value that becomes full moves: its lean entry stays unused.
	if isLean {
		s.slots[syntax.Identity(expr)] = s.putLean(lean) | leanSlot
	} else {
		s.slots[syntax.Identity(expr)] = s.putFull(value)
	}
}

// putLean stores a lean value in the next free entry and returns its index.
// Bounded header metadata is reserved only when the first payload is reached.
func (s *valueStore) putLean(lean leanValue) uint32 {
	if s.leans == nil {
		s.leans = make([][]leanValue, 0, max(1, s.headers))
	}
	if s.leanNext == 0 {
		s.leans = append(s.leans, make([]leanValue, s.width))
	}
	block := len(s.leans) - 1
	s.leans[block][s.leanNext] = lean
	index := uint32(block)*valueBlockLimit + uint32(s.leanNext)
	if s.leanNext++; s.leanNext == s.width {
		s.leanNext = 0
	}
	return index
}

// putFull stores a full value in the next free entry and returns its index.
// Few values are full: the first blocks hold one, then two, then the limit.
func (s *valueStore) putFull(value Value) uint32 {
	if s.fullNext == 0 {
		s.fulls = append(s.fulls, make([]Value, fullWidth(len(s.fulls))))
	}
	block := len(s.fulls) - 1
	s.fulls[block][s.fullNext] = value
	index := uint32(block)*valueBlockLimit + uint32(s.fullNext)
	if s.fullNext++; s.fullNext == len(s.fulls[block]) {
		s.fullNext = 0
	}
	return index
}

// fullWidth is the width of the full block after the first blocks ones.
func fullWidth(blocks int) int {
	if blocks >= 2 {
		return valueBlockLimit
	}
	return 1 << blocks
}

// leanValue is a Value without the references it rarely carries: the
// record, literal, callable, target set, interval, elements, origins and
// short-circuit operations stay nil. Every other field is kept as is.
type leanValue struct {
	Sources        []Source
	Expr           syntax.Expr
	Length         int64
	Identity       uint32
	storedIdentity uint32
	Invalidated    uint32
	Safe           Context
	Complete       bool
	NonFalse       bool
	NonNull        bool
	LengthKnown    bool
	headerMask     uint8
	upload         uint8
	scalars        bool
	matched        bool
	unlessFalse    bool
	unlessNull     bool
	skippedFalse   bool
	skippedNull    bool
	replaced       bool
	entry          bool
	plain          bool
	deep           bool
	truth          uint8
}

// leanOf returns v as a lean value, when it holds none of the references
// a lean value leaves out.
func leanOf(v *Value) (leanValue, bool) {
	if v.record != nil || v.literal != nil || v.returned != nil || v.targets != nil || v.interval != nil || v.elements != nil || v.origins != nil || v.tested != nil || v.unless != nil {
		return leanValue{}, false
	}
	return leanValue{
		Sources: v.Sources, Expr: v.Expr, Length: v.Length,
		Identity: v.Identity, storedIdentity: v.storedIdentity, Invalidated: v.Invalidated,
		Safe: v.Safe, Complete: v.Complete, NonFalse: v.NonFalse, NonNull: v.NonNull, LengthKnown: v.LengthKnown,
		headerMask: v.headerMask, upload: v.upload, scalars: v.scalars, matched: v.matched,
		unlessFalse: v.unlessFalse, unlessNull: v.unlessNull, skippedFalse: v.skippedFalse, skippedNull: v.skippedNull,
		replaced: v.replaced, entry: v.entry, plain: v.plain, deep: v.deep, truth: v.truth,
	}, true
}

// value returns the Value l was made from.
func (l *leanValue) value() Value {
	return Value{
		Sources: l.Sources, Expr: l.Expr, Length: l.Length,
		Identity: l.Identity, storedIdentity: l.storedIdentity, Invalidated: l.Invalidated,
		Safe: l.Safe, Complete: l.Complete, NonFalse: l.NonFalse, NonNull: l.NonNull, LengthKnown: l.LengthKnown,
		headerMask: l.headerMask, upload: l.upload, scalars: l.scalars, matched: l.matched,
		unlessFalse: l.unlessFalse, unlessNull: l.unlessNull, skippedFalse: l.skippedFalse, skippedNull: l.skippedNull,
		replaced: l.replaced, entry: l.entry, plain: l.plain, deep: l.deep, truth: l.truth,
	}
}
