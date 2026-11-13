package types

import "custos/internal/semantic/wire"

// The binary form of a type (see package wire). Every field is written, and
// nil slices stay distinct from empty ones, so that a decoded type is
// reflect.DeepEqual to the one encoded; TestWireFields fails when a field is
// added without being encoded here.

// Which pointer fields a type or its array facts hold.
const (
	wireCard = 1 << iota
	wireArr
	wireElement
	wireInter
	wireCases
	wireStrs
)

// Encode writes t to w.
func (t Type) Encode(w *wire.Writer) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	var held uint64
	for bit, set := range []bool{t.card != nil, t.arr != nil, t.element != nil, t.inter != nil, t.cases != nil, t.strs != nil} {
		if set {
			held |= 1 << bit
		}
	}
	w.Uint(held)
	if t.card != nil {
		w.Int(int64(t.card.minimum))
		w.Int(int64(t.card.maximum))
	}
	w.Strings(t.atoms)
	if t.arr != nil {
		t.arr.encode(w)
	}
	if t.element != nil {
		t.element.Encode(w)
	}
	w.Len(len(t.gen), t.gen == nil)
	for _, g := range t.gen {
		w.String(g.atom)
		encodeTypes(w, g.args)
	}
	if t.inter != nil {
		w.Strings(*t.inter)
	}
	encodeRanges(w, t.ints)
	encodeRanges(w, t.lens)
	if t.cases != nil {
		w.String(t.cases.class)
		w.Strings(t.cases.names)
	}
	if t.strs != nil {
		w.Strings(t.strs.only)
		w.Strings(t.strs.not)
	}
}

// DecodeType reads a type Encode wrote.
func DecodeType(r *wire.Reader) Type {
	var t Type
	if !r.Enter() {
		return t
	}
	defer r.Leave()
	held := r.Uint()
	if held&wireCard != 0 {
		t.card = &arraySize{minimum: int(r.Int()), maximum: int(r.Int())}
	}
	t.atoms = r.Strings()
	if held&wireArr != 0 {
		t.arr = decodeArrayInfo(r)
	}
	if held&wireElement != 0 {
		element := DecodeType(r)
		t.element = &element
	}
	if n, isNil := r.Len(); !isNil {
		t.gen = wire.Make[genEntry](r, n)
		for i := range t.gen {
			t.gen[i] = genEntry{atom: r.String(), args: decodeTypes(r)}
		}
	}
	if held&wireInter != 0 {
		inter := r.Strings()
		t.inter = &inter
	}
	t.ints = decodeRanges(r)
	t.lens = decodeRanges(r)
	if held&wireCases != 0 {
		t.cases = &enumCases{class: r.String(), names: r.Strings()}
	}
	if held&wireStrs != 0 {
		t.strs = &strFacts{only: r.Strings(), not: r.Strings()}
	}
	return t
}

func (a *arrayInfo) encode(w *wire.Writer) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	w.Bools(a.shape, a.sealed, a.nonEmpty, a.keyPattern != nil, a.elem != nil)
	w.Uint(uint64(a.key))
	w.Len(len(a.keys), a.keys == nil)
	for _, k := range a.keys {
		w.String(k.Name)
		k.Type.Encode(w)
		w.Bool(k.Optional)
	}
	if a.keyPattern != nil {
		a.keyPattern.Encode(w)
	}
	if a.elem != nil {
		a.elem.encode(w)
	}
}

func decodeArrayInfo(r *wire.Reader) *arrayInfo {
	a := &arrayInfo{}
	if !r.Enter() {
		return a
	}
	defer r.Leave()
	var pattern, elem bool
	r.Bools(&a.shape, &a.sealed, &a.nonEmpty, &pattern, &elem)
	a.key = uint8(r.Uint())
	if n, isNil := r.Len(); !isNil {
		a.keys = wire.Make[ShapeKey](r, n)
		for i := range a.keys {
			a.keys[i] = ShapeKey{Name: r.String(), Type: DecodeType(r), Optional: r.Bool()}
		}
	}
	if pattern {
		p := DecodeType(r)
		a.keyPattern = &p
	}
	if elem {
		a.elem = decodeArrayInfo(r)
	}
	return a
}

func encodeTypes(w *wire.Writer, ts []Type) {
	w.Len(len(ts), ts == nil)
	for _, t := range ts {
		t.Encode(w)
	}
}

func decodeTypes(r *wire.Reader) []Type {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[Type](r, n)
	for i := range out {
		out[i] = DecodeType(r)
	}
	return out
}

func encodeRanges(w *wire.Writer, rs []IntRange) {
	w.Len(len(rs), rs == nil)
	for _, x := range rs {
		w.Int(x.Lo)
		w.Int(x.Hi)
	}
}

func decodeRanges(r *wire.Reader) []IntRange {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[IntRange](r, n)
	for i := range out {
		out[i] = IntRange{Lo: r.Int(), Hi: r.Int()}
	}
	return out
}
