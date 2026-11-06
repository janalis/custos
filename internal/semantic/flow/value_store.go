package flow

import "custos/internal/php/syntax"

// valueStore keeps expression values in small, lazily allocated blocks. Hash
// slots contain only indices; hints for unreachable expressions do not reserve
// their payloads. Callers receive value copies, never pointers into the store.
type valueStore struct {
	slots       map[syntax.Expr]uint
	blocks      [][]Value
	width, next int
}

const valueBlockLimit = 16

func newValueStore(capacity int) valueStore {
	width := min(capacity, valueBlockLimit)
	if width == 0 {
		width = valueBlockLimit
	}
	return valueStore{slots: make(map[syntax.Expr]uint, capacity), width: width}
}

func (s *valueStore) get(expr syntax.Expr) (Value, bool) {
	index, known := s.slots[expr]
	if !known {
		return Value{}, false
	}
	return s.blocks[index/valueBlockLimit][index%valueBlockLimit], true
}

func (s *valueStore) at(expr syntax.Expr) Value {
	value, _ := s.get(expr)
	return value
}

func (s *valueStore) set(expr syntax.Expr, value Value) {
	if index, known := s.slots[expr]; known {
		s.blocks[index/valueBlockLimit][index%valueBlockLimit] = value
		return
	}
	if s.slots == nil {
		s.slots = map[syntax.Expr]uint{}
		s.width = valueBlockLimit
	}
	// Indices reserve a full block stride even when a small scope uses a
	// narrower block. Reads and updates can split them with constant shifts.
	if s.next == 0 {
		s.blocks = append(s.blocks, make([]Value, s.width))
	}
	block := len(s.blocks) - 1
	s.slots[expr] = uint(block)*valueBlockLimit + uint(s.next)
	s.blocks[block][s.next] = value
	s.next++
	if s.next == s.width {
		s.next = 0
	}
}
