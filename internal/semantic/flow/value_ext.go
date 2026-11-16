package flow

import "custos/internal/php/syntax"

// valueExt holds the references and facts a Value rarely carries; most
// values carry none and leave Value.ext nil, which keeps a Value small to
// copy and store. An extension is shared by every copy of the value it was
// made for, so it is never written once installed: a setter installs a
// modified copy, and an extension whose fields are all zero is nil, so that
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

// setUnlessFacts sets unless and the four facts that qualify it at once.
func (v *Value) setUnlessFacts(unless *syntax.Binary, unlessFalse, unlessNull, skippedFalse, skippedNull bool) {
	ext := v.edited()
	ext.unless, ext.unlessFalse, ext.unlessNull, ext.skippedFalse, ext.skippedNull = unless, unlessFalse, unlessNull, skippedFalse, skippedNull
	v.ext = derived(v.ext, ext)
}

// edited returns v's extension, or a zero one, for a setter to modify.
func (v *Value) edited() valueExt {
	if v.ext == nil {
		return valueExt{}
	}
	return *v.ext
}

// Each field of the extension has an accessor, which reads zero from a
// value without one, and a setter, which leaves the extension as it is when
// the field already holds what it sets.

func (v Value) record() *resourceReturn {
	if v.ext == nil {
		return nil
	}
	return v.ext.record
}

func (v *Value) setRecord(x *resourceReturn) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.record == x {
		return
	}
	ext := v.edited()
	ext.record = x
	v.ext = rare(ext)
}

func (v Value) literal() *constantScalar {
	if v.ext == nil {
		return nil
	}
	return v.ext.literal
}

func (v *Value) setLiteral(x *constantScalar) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.literal == x {
		return
	}
	ext := v.edited()
	ext.literal = x
	v.ext = rare(ext)
}

func (v Value) returned() *callableValue {
	if v.ext == nil {
		return nil
	}
	return v.ext.returned
}

func (v *Value) setReturned(x *callableValue) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.returned == x {
		return
	}
	ext := v.edited()
	ext.returned = x
	v.ext = rare(ext)
}

func (v Value) targets() *callableSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.targets
}

func (v *Value) setTargets(x *callableSet) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.targets == x {
		return
	}
	ext := v.edited()
	ext.targets = x
	v.ext = rare(ext)
}

func (v Value) interval() *intInterval {
	if v.ext == nil {
		return nil
	}
	return v.ext.interval
}

func (v *Value) setInterval(x *intInterval) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.interval == x {
		return
	}
	ext := v.edited()
	ext.interval = x
	v.ext = rare(ext)
}

func (v Value) elements() *elementSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.elements
}

func (v *Value) setElements(x *elementSet) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.elements == x {
		return
	}
	ext := v.edited()
	ext.elements = x
	v.ext = rare(ext)
}

func (v Value) origins() *originSet {
	if v.ext == nil {
		return nil
	}
	return v.ext.origins
}

func (v *Value) setOrigins(x *originSet) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.origins == x {
		return
	}
	ext := v.edited()
	ext.origins = x
	v.ext = rare(ext)
}

func (v Value) tested() *syntax.Binary {
	if v.ext == nil {
		return nil
	}
	return v.ext.tested
}

func (v *Value) setTested(x *syntax.Binary) {
	if v.ext == nil && x == nil || v.ext != nil && v.ext.tested == x {
		return
	}
	ext := v.edited()
	ext.tested = x
	v.ext = rare(ext)
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

func (v *Value) setLength(x int64) {
	if v.ext == nil && x == 0 || v.ext != nil && v.ext.length == x {
		return
	}
	ext := v.edited()
	ext.length = x
	v.ext = rare(ext)
}

func (v Value) storedIdentity() uint32 {
	if v.ext == nil {
		return 0
	}
	return v.ext.storedIdentity
}

func (v *Value) setStoredIdentity(x uint32) {
	if v.ext == nil && x == 0 || v.ext != nil && v.ext.storedIdentity == x {
		return
	}
	ext := v.edited()
	ext.storedIdentity = x
	v.ext = rare(ext)
}

func (v Value) lengthKnown() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.lengthKnown
}

func (v *Value) setLengthKnown(x bool) {
	if v.ext == nil && !x || v.ext != nil && v.ext.lengthKnown == x {
		return
	}
	ext := v.edited()
	ext.lengthKnown = x
	v.ext = rare(ext)
}

func (v Value) headerMask() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.headerMask
}

func (v *Value) setHeaderMask(x uint8) {
	if v.ext == nil && x == 0 || v.ext != nil && v.ext.headerMask == x {
		return
	}
	ext := v.edited()
	ext.headerMask = x
	v.ext = rare(ext)
}

func (v Value) upload() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.upload
}

func (v *Value) setUpload(x uint8) {
	if v.ext == nil && x == 0 || v.ext != nil && v.ext.upload == x {
		return
	}
	ext := v.edited()
	ext.upload = x
	v.ext = rare(ext)
}

func (v Value) scalars() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.scalars
}

func (v *Value) setScalars(x bool) {
	if v.ext == nil && !x || v.ext != nil && v.ext.scalars == x {
		return
	}
	ext := v.edited()
	ext.scalars = x
	v.ext = rare(ext)
}

func (v Value) matched() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.matched
}

func (v *Value) setMatched(x bool) {
	if v.ext == nil && !x || v.ext != nil && v.ext.matched == x {
		return
	}
	ext := v.edited()
	ext.matched = x
	v.ext = rare(ext)
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
	if v.ext == nil && !x || v.ext != nil && v.ext.replaced == x {
		return
	}
	ext := v.edited()
	ext.replaced = x
	v.ext = rare(ext)
}

func (v Value) entry() bool {
	if v.ext == nil {
		return false
	}
	return v.ext.entry
}

func (v *Value) setEntry(x bool) {
	if v.ext == nil && !x || v.ext != nil && v.ext.entry == x {
		return
	}
	ext := v.edited()
	ext.entry = x
	v.ext = rare(ext)
}

func (v Value) truth() uint8 {
	if v.ext == nil {
		return 0
	}
	return v.ext.truth
}

func (v *Value) setTruth(x uint8) {
	if v.ext == nil && x == 0 || v.ext != nil && v.ext.truth == x {
		return
	}
	ext := v.edited()
	ext.truth = x
	v.ext = rare(ext)
}
