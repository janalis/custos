package flow

import "custos/internal/php/syntax"

// valueExt holds the references and facts a Value rarely carries; most
// values carry none and leave Value.ext nil, which keeps a Value small to
// copy and store. An extension is shared by every copy of the value it was
// made for, so it is never written once installed: a setter or an edit
// installs a modified copy, and an extension whose fields are all zero is nil, so that
// reflect.DeepEqual compares values as it did their fields.
type valueExt struct {
	record   *resourceReturn
	literal  *constantScalar
	returned *callableValue
	targets  *callableSet
	interval *intInterval
	// elements holds the literal-keyed elements of an array variable that
	// are tracked on their own. Only a binding carries them: the value of an
	// expression never does.
	elements *elementSet
	// origins lists the expressions that may have produced a binding the
	// paths disagree on (Expr is then nil), each with what its own paths
	// established. It means nothing while Expr is set.
	origins *originSet
	// tested and unless belong to a binding only, never to the value of an
	// expression. tested is the short-circuit operation whose left operand
	// last read the binding, which was not written since. unless is such an
	// operation whose right operand assigns the binding and tests it in
	// place: the binding is not false (unlessFalse) or not null (unlessNull)
	// except where the last evaluation of the operation skipped that
	// operand. Between the evaluation of the operation and the branch on its
	// outcome, skippedFalse and skippedNull record that the value the
	// skipped operand left in place was not false or not null.
	tested, unless *syntax.Binary
	length         int64
	storedIdentity uint32
	lengthKnown    bool
	headerMask     uint8
	// upload marks the table of uploaded files or one of its entries as PHP
	// built it: only there are the server-chosen fields what they claim.
	upload uint8
	// scalars marks an array whose values are numbers, booleans or null
	// while its keys remain whatever the request sent.
	scalars bool
	// matched marks request data a switch found equal to one of its string
	// labels: the text of a label, not what the request sent.
	matched bool
	// unlessFalse, unlessNull, skippedFalse and skippedNull qualify unless.
	unlessFalse, unlessNull, skippedFalse, skippedNull bool
	// replaced marks a binding that code the analysis does not follow may
	// have replaced since it was last written, on some path.
	replaced bool
	// entry marks the binding of a static variable that still holds what it
	// held when its declaration was reached: nothing wrote it since.
	entry bool
	truth uint8
}

// rare returns an extension holding x, or nil when every field of x is zero.
func rare(x valueExt) *valueExt {
	if x == (valueExt{}) {
		return nil
	}
	p := new(valueExt)
	*p = x
	return p
}

// derived returns an extension holding x, which is base when base holds
// just that.
func derived(base *valueExt, x valueExt) *valueExt {
	if base != nil && *base == x {
		return base
	}
	return rare(x)
}

// edit installs what change makes of a copy of v's extension, or of a zero
// one: v's own extension when change left it as it was, none when it left no
// field set. Setting several fields in one edit allocates once at most. The
// copy goes in and out by value: a pointer to it handed to change would make
// it escape to the heap.
func (v *Value) edit(change func(x valueExt) valueExt) {
	var x valueExt
	if v.ext != nil {
		x = *v.ext
	}
	v.ext = derived(v.ext, change(x))
}

// setUnlessFacts sets unless and the four facts that qualify it at once.
func (v *Value) setUnlessFacts(unless *syntax.Binary, unlessFalse, unlessNull, skippedFalse, skippedNull bool) {
	v.edit(func(x valueExt) valueExt {
		x.unless, x.unlessFalse, x.unlessNull, x.skippedFalse, x.skippedNull = unless, unlessFalse, unlessNull, skippedFalse, skippedNull
		return x
	})
}

// Each field of the extension has an accessor, which reads zero from a
// value without one. Most have a setter, which leaves the extension as it is
// when the field already holds what it sets; the length is set with whether
// it is known, in one edit.

func (v Value) record() *resourceReturn {
	if v.ext == nil {
		return nil
	}
	return v.ext.record
}

func (v *Value) setRecord(x *resourceReturn) {
	if v.record() != x {
		v.edit(func(e valueExt) valueExt { e.record = x; return e })
	}
}

func (v Value) literal() *constantScalar {
	if v.ext == nil {
		return nil
	}
	return v.ext.literal
}

func (v *Value) setLiteral(x *constantScalar) {
	if v.literal() != x {
		v.edit(func(e valueExt) valueExt { e.literal = x; return e })
	}
}

func (v Value) returned() *callableValue {
	if v.ext == nil {
		return nil
	}
	return v.ext.returned
}

func (v *Value) setReturned(x *callableValue) {
	if v.returned() != x {
		v.edit(func(e valueExt) valueExt { e.returned = x; return e })
	}
}

func (v Value) targets() *callableSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.targets
}

func (v *Value) setTargets(x *callableSet) {
	if v.targets() != x {
		v.edit(func(e valueExt) valueExt { e.targets = x; return e })
	}
}

func (v Value) interval() *intInterval {
	if v.ext == nil {
		return nil
	}
	return v.ext.interval
}

func (v *Value) setInterval(x *intInterval) {
	if v.interval() != x {
		v.edit(func(e valueExt) valueExt { e.interval = x; return e })
	}
}

func (v Value) elements() *elementSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.elements
}

func (v *Value) setElements(x *elementSet) {
	if v.elements() != x {
		v.edit(func(e valueExt) valueExt { e.elements = x; return e })
	}
}

func (v Value) origins() *originSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.origins
}

func (v *Value) setOrigins(x *originSet) {
	if v.origins() != x {
		v.edit(func(e valueExt) valueExt { e.origins = x; return e })
	}
}

func (v Value) tested() *syntax.Binary {
	if v.ext == nil {
		return nil
	}
	return v.ext.tested
}

func (v *Value) setTested(x *syntax.Binary) {
	if v.tested() != x {
		v.edit(func(e valueExt) valueExt { e.tested = x; return e })
	}
}

func (v Value) unless() *syntax.Binary {
	if v.ext == nil {
		return nil
	}
	return v.ext.unless
}

func (v Value) length() int64 {
	if v.ext == nil {
		return 0
	}
	return v.ext.length
}

func (v Value) storedIdentity() uint32 {
	if v.ext == nil {
		return 0
	}
	return v.ext.storedIdentity
}

func (v *Value) setStoredIdentity(x uint32) {
	if v.storedIdentity() != x {
		v.edit(func(e valueExt) valueExt { e.storedIdentity = x; return e })
	}
}

func (v Value) lengthKnown() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.lengthKnown
}

func (v *Value) setLengthKnown(x bool) {
	if v.lengthKnown() != x {
		v.edit(func(e valueExt) valueExt { e.lengthKnown = x; return e })
	}
}

func (v Value) headerMask() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.headerMask
}

func (v *Value) setHeaderMask(x uint8) {
	if v.headerMask() != x {
		v.edit(func(e valueExt) valueExt { e.headerMask = x; return e })
	}
}

func (v Value) upload() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.upload
}

func (v *Value) setUpload(x uint8) {
	if v.upload() != x {
		v.edit(func(e valueExt) valueExt { e.upload = x; return e })
	}
}

func (v Value) scalars() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.scalars
}

func (v *Value) setScalars(x bool) {
	if v.scalars() != x {
		v.edit(func(e valueExt) valueExt { e.scalars = x; return e })
	}
}

func (v Value) matched() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.matched
}

func (v *Value) setMatched(x bool) {
	if v.matched() != x {
		v.edit(func(e valueExt) valueExt { e.matched = x; return e })
	}
}

func (v Value) unlessFalse() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.unlessFalse
}

func (v Value) unlessNull() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.unlessNull
}

func (v Value) skippedFalse() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.skippedFalse
}

func (v Value) skippedNull() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.skippedNull
}

func (v Value) replaced() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.replaced
}

func (v *Value) setReplaced(x bool) {
	if v.replaced() != x {
		v.edit(func(e valueExt) valueExt { e.replaced = x; return e })
	}
}

func (v Value) entry() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.entry
}

func (v *Value) setEntry(x bool) {
	if v.entry() != x {
		v.edit(func(e valueExt) valueExt { e.entry = x; return e })
	}
}

func (v Value) truth() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.truth
}

func (v *Value) setTruth(x uint8) {
	if v.truth() != x {
		v.edit(func(e valueExt) valueExt { e.truth = x; return e })
	}
}
