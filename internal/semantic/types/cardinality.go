package types

// arraySize is immutable, like the key and element contracts it complements.
type arraySize struct{ minimum, maximum int }

// WithArrayCardinality attaches finite entry-count bounds to an array-only
// type. Invalid bounds discard the proof. Key-changing operations must derive
// new bounds rather than retaining them through WithoutShape.
func (t Type) WithArrayCardinality(minimum, maximum int) Type {
	return t.withCardinality(minimum, maximum, t.IsArrayLike() && minimum >= 0 && maximum >= minimum)
}

func (t Type) withCardinality(minimum, maximum int, known bool) Type {
	t.card = nil
	if known && t.IsArrayLike() {
		t = t.WithNonEmpty(minimum > 0)
		t.card = &arraySize{minimum, maximum}
	}
	return t
}

func unionCardinality(ts []Type) (minimum, maximum int, known bool) {
	for _, t := range ts {
		if t.IsBottom() {
			continue
		}
		low, high, ok := t.ArrayCardinality()
		if !ok {
			return 0, 0, false
		}
		if !known {
			minimum, maximum, known = low, high, true
		} else {
			minimum, maximum = min(minimum, low), max(maximum, high)
		}
	}
	return
}
