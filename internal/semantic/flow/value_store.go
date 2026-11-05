package flow

import "custos/internal/php/syntax"

// valueStore keeps expression values in small, lazily allocated blocks. Hash
// slots contain only indices; hints for unreachable expressions do not reserve
// their payloads. Callers receive value copies, never pointers into the store.
type valueStore struct {
	slots        map[syntax.Expr]int
	blocks       [][]Value
	width, count int
}

const valueBlockLimit = 16

func newValueStore(capacity int) valueStore {
	width := min(capacity, valueBlockLimit)
	if width == 0 {
		width = valueBlockLimit
	}
	return valueStore{slots: make(map[syntax.Expr]int, capacity), width: width}
}

func (s *valueStore) get(expr syntax.Expr) (Value, bool) {
	index, known := s.slots[expr]
	if !known {
		return Value{}, false
	}
	return s.blocks[index/s.width][index%s.width], true
}

func (s *valueStore) at(expr syntax.Expr) Value {
	value, _ := s.get(expr)
	return value
}

func (s *valueStore) set(expr syntax.Expr, value Value) {
	if index, known := s.slots[expr]; known {
		s.blocks[index/s.width][index%s.width] = value
		return
	}
	if s.slots == nil {
		s.slots = map[syntax.Expr]int{}
		s.width = valueBlockLimit
	}
	index := s.count
	if index%s.width == 0 {
		s.blocks = append(s.blocks, make([]Value, s.width))
	}
	s.slots[expr] = index
	s.blocks[index/s.width][index%s.width] = value
	s.count++
}
