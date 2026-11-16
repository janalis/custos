package flow

import "custos/internal/php/syntax"

// valueStore keeps expression values in small, lazily allocated blocks. Hash
// slots contain bounded 32-bit indices; hints for unreachable expressions do not reserve
// their payloads. Callers receive value copies, never pointers into the store.
type valueStore struct {
	slots  map[syntax.NodeIdentity]uint32
	blocks [][]Value
	// Parameter-read queries need the evaluated simple variables, not other node types.
	variables            []*syntax.Variable
	width, next, headers int
}

const valueBlockLimit = 4

// Even a narrow block's reserved stride fits within one bounded scope.
const _ uint32 = MaxRetained * valueBlockLimit

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
	return s.blocks[index/valueBlockLimit][index%valueBlockLimit], true
}

func (s *valueStore) at(expr syntax.Expr) Value {
	value, _ := s.get(expr)
	return value
}

func (s *valueStore) set(expr syntax.Expr, value Value) {
	if index, known := s.slots[syntax.Identity(expr)]; known {
		s.blocks[index/valueBlockLimit][index%valueBlockLimit] = value
		return
	}
	if s.slots == nil {
		s.slots = map[syntax.NodeIdentity]uint32{}
		s.width = valueBlockLimit
	}
	// Indices reserve a full block stride even when a small scope uses a
	// narrower block. Reads and updates can split them with constant shifts.
	// Reserve bounded header metadata only when the first payload is reached.
	if s.blocks == nil {
		s.blocks = make([][]Value, 0, max(1, s.headers))
	}
	if s.next == 0 {
		s.blocks = append(s.blocks, make([]Value, s.width))
	}
	if variable, ok := expr.(*syntax.Variable); ok && variable.NameExpr == nil {
		s.variables = append(s.variables, variable)
	}
	block := len(s.blocks) - 1
	s.slots[syntax.Identity(expr)] = uint32(block)*valueBlockLimit + uint32(s.next)
	s.blocks[block][s.next] = value
	s.next++
	if s.next == s.width {
		s.next = 0
	}
}
