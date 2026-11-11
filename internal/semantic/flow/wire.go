package flow

import (
	"errors"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
	"custos/internal/semantic/wire"
)

// The binary form of summaries (see package wire), which a cache kept
// between runs stores. Every field of every type a File or a Basis reaches
// is written, and nil slices and maps stay distinct from empty ones, so
// that a decoded value is reflect.DeepEqual to the one encoded: a decoded
// summary answers Current and Reextract as the one encoded. Options are not
// written: a decoder gives every file and snapshot it reads its own.
// TestWireFields fails when a field is added without being encoded here.
//
// Summaries keep no syntax: a value still holding a node cannot be encoded,
// and fails the encoder instead of losing it.

// errSyntax fails an encoder given a value that holds a syntax node.
var errSyntax = errors.New("flow: summary holds a syntax node")

// errOverlay fails an encoder given a snapshot that is not a Basis.
var errOverlay = errors.New("flow: snapshot is not a basis")

// Encoder writes files and the snapshots they were summarized against.
// They share one table of strings and one of summaries: a summary many
// snapshots hold is written once.
type Encoder struct {
	strings *wire.Strings
	body    *wire.Writer
	table   *wire.Writer
	scratch *wire.Writer
	count   uint64
	// byContent numbers the summaries by their encoding, byPointer by the
	// pointer a snapshot holds them through.
	byContent map[string]uint64
	byPointer map[*Summary]uint64
}

// NewEncoder returns an empty encoder.
func NewEncoder() *Encoder {
	s := wire.NewStrings()
	return &Encoder{strings: s, body: wire.NewWriter(s), table: wire.NewWriter(s), scratch: wire.NewWriter(s), byContent: map[string]uint64{}, byPointer: map[*Summary]uint64{}}
}

// Writer returns the writer of the encoder's body, for the data a caller
// keeps around files and snapshots.
func (e *Encoder) Writer() *wire.Writer { return e.body }

// Mark returns the position Rollback returns to.
func (e *Encoder) Mark() int { return len(e.body.Buf) }

// Rollback drops what the body holds past mark and clears its error: an
// entry that failed to encode leaves nothing. Summaries the table gained
// meanwhile stay, unused.
func (e *Encoder) Rollback(mark int) {
	buf := e.body.Buf[:mark]
	e.body.Reset()
	e.body.Buf = buf
}

// Bytes returns the encoding: the strings, the summaries, then the body.
func (e *Encoder) Bytes() []byte {
	out := e.strings.Append(nil)
	header := wire.NewWriter(e.strings)
	header.Uint(e.count)
	header.Uint(uint64(len(e.table.Buf)))
	out = append(out, header.Buf...)
	out = append(out, e.table.Buf...)
	return append(out, e.body.Buf...)
}

// File writes f; nil is written too.
func (e *Encoder) File(f *File) {
	w := e.body
	w.Bool(f != nil)
	if f == nil {
		return
	}
	w.String(f.Path)
	w.Len(len(f.Summaries), f.Summaries == nil)
	for key, s := range f.Summaries {
		w.String(key)
		w.Uint(e.summary(&s, false))
	}
	encodeFlags(w, f.consulted)
	w.Strings(f.held)
	encodeFlags(w, f.served)
	w.Len(len(f.callables), f.callables == nil)
	for key, reads := range f.callables {
		w.String(key)
		w.Len(len(reads), reads == nil)
		for read, bits := range reads {
			w.String(read)
			w.Uint(uint64(bits))
		}
	}
}

// Snapshot writes s, which must be nil or a Basis.
func (e *Encoder) Snapshot(s *Snapshot) {
	w := e.body
	w.Bool(s != nil)
	if s == nil {
		return
	}
	if s.base != nil || s.own != nil || s.files != nil {
		w.Fail(errOverlay)
		return
	}
	w.Len(len(s.summaries), s.summaries == nil)
	for key, summary := range s.summaries {
		w.String(key)
		if summary == nil {
			w.Uint(0)
		} else {
			w.Uint(e.summary(summary, true) + 1)
		}
	}
	w.Len(len(s.shared), s.shared == nil)
	for key, n := range s.shared {
		w.String(key)
		w.Int(int64(n))
	}
}

// Err returns the error that failed the body since the last Rollback.
func (e *Encoder) Err() error { return e.body.Err() }

// summary returns the number of s in the table, adding it when no summary
// encoded alike is there; held tells a pointer a snapshot holds, which the
// encoder may number by itself. A summary that cannot be encoded fails the
// body.
func (e *Encoder) summary(s *Summary, held bool) uint64 {
	if held {
		if id, ok := e.byPointer[s]; ok {
			return id
		}
	}
	e.scratch.Reset()
	encodeSummary(e.scratch, s)
	if err := e.scratch.Err(); err != nil {
		e.body.Fail(err)
		return 0
	}
	id, ok := e.byContent[string(e.scratch.Buf)]
	if !ok {
		id = e.count
		e.count++
		e.byContent[string(e.scratch.Buf)] = id
		e.table.Buf = append(e.table.Buf, e.scratch.Buf...)
	}
	if held {
		e.byPointer[s] = id
	}
	return id
}

// Decoder reads what an Encoder wrote, giving every file and snapshot
// options.
type Decoder struct {
	r         *wire.Reader
	summaries []Summary
	options   Options
}

// NewDecoder reads the tables of b; the error is wire.ErrCorrupt when they
// do not decode.
func NewDecoder(b []byte, options Options) (*Decoder, error) {
	table, rest, err := wire.ReadStrings(b)
	if err != nil {
		return nil, err
	}
	header := wire.NewReader(rest, table)
	count, size := header.Uint(), header.Uint()
	start := header.Offset()
	// Every summary takes at least one byte.
	if header.Err() != nil || size > uint64(len(rest)-start) || count > size {
		return nil, wire.ErrCorrupt
	}
	sums := wire.NewReader(rest[start:start+int(size)], table)
	d := &Decoder{summaries: wire.Make[Summary](sums, int(count)), options: options}
	for i := range d.summaries {
		d.summaries[i] = decodeSummary(sums)
	}
	if !sums.Done() {
		return nil, wire.ErrCorrupt
	}
	d.r = wire.NewReader(rest[start+int(size):], table)
	return d, nil
}

// Reader returns the reader of the body, for what the caller wrote around
// files and snapshots.
func (d *Decoder) Reader() *wire.Reader { return d.r }

// ref returns the summary numbered id, failing the reader when there is none.
func (d *Decoder) ref(id uint64) *Summary {
	if id >= uint64(len(d.summaries)) {
		d.r.Fail()
		return &Summary{}
	}
	return &d.summaries[id]
}

// File reads a file File wrote.
func (d *Decoder) File() *File {
	r := d.r
	if !r.Bool() {
		return nil
	}
	f := &File{options: d.options, Path: r.String()}
	if n, isNil := r.Len(); !isNil {
		f.Summaries = wire.MakeMap[string, Summary](r, n)
		for range n {
			key := r.String()
			f.Summaries[key] = *d.ref(r.Uint())
		}
	}
	f.consulted = decodeFlags(r)
	f.held = r.Strings()
	f.served = decodeFlags(r)
	if n, isNil := r.Len(); !isNil {
		f.callables = wire.MakeMap[string, map[string]uint8](r, n)
		for range n {
			key := r.String()
			var reads map[string]uint8
			if m, isNil := r.Len(); !isNil {
				reads = wire.MakeMap[string, uint8](r, m)
				for range m {
					read := r.String()
					reads[read] = uint8(r.Uint())
				}
			}
			f.callables[key] = reads
		}
	}
	return f
}

// Snapshot reads a snapshot Snapshot wrote.
func (d *Decoder) Snapshot() *Snapshot {
	r := d.r
	if !r.Bool() {
		return nil
	}
	s := &Snapshot{options: d.options}
	if n, isNil := r.Len(); !isNil {
		s.summaries = wire.MakeMap[string, *Summary](r, n)
		for range n {
			key := r.String()
			var summary *Summary
			if id := r.Uint(); id > 0 {
				summary = d.ref(id - 1)
			}
			s.summaries[key] = summary
		}
	}
	if n, isNil := r.Len(); !isNil {
		s.shared = wire.MakeMap[string, int](r, n)
		for range n {
			key := r.String()
			s.shared[key] = int(r.Int())
		}
	}
	return s
}

func encodeFlags(w *wire.Writer, m map[string]bool) {
	w.Len(len(m), m == nil)
	for key, v := range m {
		w.String(key)
		w.Bool(v)
	}
}

func decodeFlags(r *wire.Reader) map[string]bool {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	m := wire.MakeMap[string, bool](r, n)
	for range n {
		key := r.String()
		m[key] = r.Bool()
	}
	return m
}

func encodeSummary(w *wire.Writer, s *Summary) {
	w.Bools(s.pure != nil, s.forward != nil, s.returned != nil, s.resource != nil, s.dispatch != nil, s.argumentSensitive, s.Generator, s.Never, s.Fixed, s.Quiet, s.Spares)
	if s.pure != nil {
		encodePure(w, s.pure)
	}
	encodeValues(w, s.defaults)
	if s.forward != nil {
		encodeForward(w, s.forward)
	}
	if s.returned != nil {
		encodeReturned(w, s.returned)
	}
	if s.resource != nil {
		encodeResource(w, s.resource)
	}
	w.Len(len(s.effects), s.effects == nil)
	for _, x := range s.effects {
		w.Bools(x.onTrue, x.checked)
		w.Int(int64(x.parameter))
		w.String(x.operation)
	}
	if s.dispatch != nil {
		encodeSinkClass(w, s.dispatch)
	}
	encodeValue(w, &s.Return)
	encodeSinks(w, s.Sinks)
	w.Len(len(s.ByRef), s.ByRef == nil)
	for i, v := range s.ByRef {
		w.Int(int64(i))
		encodeValue(w, &v)
	}
}

func decodeSummary(r *wire.Reader) Summary {
	var s Summary
	var pure, forward, returned, resource, dispatch bool
	r.Bools(&pure, &forward, &returned, &resource, &dispatch, &s.argumentSensitive, &s.Generator, &s.Never, &s.Fixed, &s.Quiet, &s.Spares)
	if pure {
		s.pure = decodePure(r)
	}
	s.defaults = decodeValues(r)
	if forward {
		s.forward = decodeForward(r)
	}
	if returned {
		s.returned = decodeReturned(r)
	}
	if resource {
		s.resource = decodeResource(r)
	}
	if n, isNil := r.Len(); !isNil {
		s.effects = wire.Make[resourceEffect](r, n)
		for i := range s.effects {
			x := &s.effects[i]
			r.Bools(&x.onTrue, &x.checked)
			x.parameter = int(r.Int())
			x.operation = r.String()
		}
	}
	if dispatch {
		s.dispatch = decodeSinkClass(r)
	}
	s.Return = decodeValue(r)
	s.Sinks = decodeSinks(r)
	if n, isNil := r.Len(); !isNil {
		s.ByRef = wire.MakeMap[int, Value](r, n)
		for range n {
			i := int(r.Int())
			s.ByRef[i] = decodeValue(r)
		}
	}
	return s
}

func encodePure(w *wire.Writer, c *pureContract) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	encodeParams(w, c.params)
	w.Bools(c.expr != nil, c.weakReturn)
	if c.expr != nil {
		encodeExpr(w, c.expr)
	}
	w.Strings(c.calls)
	c.result.Encode(w)
	w.Int(int64(c.outer))
}

func decodePure(r *wire.Reader) *pureContract {
	c := &pureContract{}
	if !r.Enter() {
		return c
	}
	defer r.Leave()
	c.params = decodeParams(r)
	var expr bool
	r.Bools(&expr, &c.weakReturn)
	if expr {
		c.expr = decodeExpr(r)
	}
	c.calls = r.Strings()
	c.result = types.DecodeType(r)
	c.outer = int(r.Int())
	return c
}

func encodeParams(w *wire.Writer, params []pureParam) {
	w.Len(len(params), params == nil)
	for _, p := range params {
		w.String(p.name)
		p.typ.Encode(w)
		w.Bools(p.initial != nil, p.optional, p.nullable)
		if p.initial != nil {
			encodeExpr(w, p.initial)
		}
	}
}

func decodeParams(r *wire.Reader) []pureParam {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[pureParam](r, n)
	for i := range out {
		p := &out[i]
		p.name = r.String()
		p.typ = types.DecodeType(r)
		var initial bool
		r.Bools(&initial, &p.optional, &p.nullable)
		if initial {
			p.initial = decodeExpr(r)
		}
	}
	return out
}

func encodeExpr(w *wire.Writer, x *pureExpr) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	w.Bools(x.weakArguments, x.projects, x.appendKey, x.stopGuard, x.prepared, x.literal != nil, x.left != nil, x.right != nil, x.other != nil, x.call != nil)
	w.Uint(uint64(x.kind))
	w.Len(len(x.destruct), x.destruct == nil)
	for _, f := range x.destruct {
		w.String(f.name)
		w.String(f.key)
	}
	w.Int(int64(x.slot))
	w.Uint(uint64(x.op))
	if x.literal != nil {
		encodeScalar(w, x.literal)
	}
	for _, sub := range []*pureExpr{x.left, x.right, x.other} {
		if sub != nil {
			encodeExpr(w, sub)
		}
	}
	if x.call != nil {
		encodePure(w, x.call)
	}
	w.Len(len(x.args), x.args == nil)
	for _, a := range x.args {
		w.Bool(a != nil)
		if a != nil {
			encodeExpr(w, a)
		}
	}
	w.Int(int64(x.cost))
	w.Int(int64(x.depth))
}

func decodeExpr(r *wire.Reader) *pureExpr {
	x := &pureExpr{}
	if !r.Enter() {
		return x
	}
	defer r.Leave()
	var literal, left, right, other, call bool
	r.Bools(&x.weakArguments, &x.projects, &x.appendKey, &x.stopGuard, &x.prepared, &literal, &left, &right, &other, &call)
	x.kind = uint8(r.Uint())
	if n, isNil := r.Len(); !isNil {
		x.destruct = wire.Make[pureField](r, n)
		for i := range x.destruct {
			x.destruct[i] = pureField{name: r.String(), key: r.String()}
		}
	}
	x.slot = int(r.Int())
	x.op = syntax.TokenKind(r.Uint())
	if literal {
		x.literal = decodeScalar(r)
	}
	if left {
		x.left = decodeExpr(r)
	}
	if right {
		x.right = decodeExpr(r)
	}
	if other {
		x.other = decodeExpr(r)
	}
	if call {
		x.call = decodePure(r)
	}
	if n, isNil := r.Len(); !isNil {
		x.args = wire.Make[*pureExpr](r, n)
		for i := range x.args {
			if r.Bool() {
				x.args[i] = decodeExpr(r)
			}
		}
	}
	x.cost = int(r.Int())
	x.depth = int(r.Int())
	return x
}

func encodeScalar(w *wire.Writer, c *constantScalar) {
	w.Uint(uint64(c.k))
	w.Int(c.i)
	w.Float(c.f)
	w.String(c.s)
	w.Bool(c.b)
}

func decodeScalar(r *wire.Reader) *constantScalar {
	return &constantScalar{k: byte(r.Uint()), i: r.Int(), f: r.Float(), s: r.String(), b: r.Bool()}
}

func encodeResource(w *wire.Writer, x *resourceReturn) {
	w.Bools(x.fresh, x.checked)
	w.Int(int64(x.token))
	w.Int(int64(x.parameter))
	w.String(x.operation)
}

func decodeResource(r *wire.Reader) *resourceReturn {
	x := &resourceReturn{}
	r.Bools(&x.fresh, &x.checked)
	x.token = int(r.Int())
	x.parameter = int(r.Int())
	x.operation = r.String()
	return x
}

func encodeForward(w *wire.Writer, x *forwardReturn) {
	w.Int(int64(x.parameter))
	x.argument.Encode(w)
	x.result.Encode(w)
	w.Bool(x.weakReturn)
}

func decodeForward(r *wire.Reader) *forwardReturn {
	return &forwardReturn{parameter: int(r.Int()), argument: types.DecodeType(r), result: types.DecodeType(r), weakReturn: r.Bool()}
}

func encodeReturned(w *wire.Writer, x *returnedCallable) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	w.Bools(x.pure != nil, x.forward != nil, x.argumentSensitive)
	if x.pure != nil {
		encodePure(w, x.pure)
	}
	encodeParams(w, x.outerParams)
	if x.forward != nil {
		encodeForward(w, x.forward)
	}
	encodeIndexParams(w, x.params)
	encodeValues(w, x.defaults)
	encodeValue(w, &x.result)
	encodeSinks(w, x.sinks)
}

func decodeReturned(r *wire.Reader) *returnedCallable {
	x := &returnedCallable{}
	if !r.Enter() {
		return x
	}
	defer r.Leave()
	var pure, forward bool
	r.Bools(&pure, &forward, &x.argumentSensitive)
	if pure {
		x.pure = decodePure(r)
	}
	x.outerParams = decodeParams(r)
	if forward {
		x.forward = decodeForward(r)
	}
	x.params = decodeIndexParams(r)
	x.defaults = decodeValues(r)
	x.result = decodeValue(r)
	x.sinks = decodeSinks(r)
	return x
}

func encodeIndexParams(w *wire.Writer, params []index.Param) {
	w.Len(len(params), params == nil)
	for _, p := range params {
		w.String(p.Name)
		w.String(p.Type)
		w.String(p.DocType)
		w.Bools(p.Optional, p.ByRef, p.Variadic, p.Promoted)
		w.String(p.Default)
		w.String(p.Out)
		w.String(p.Gen)
		w.Uint(uint64(p.Avail.From))
		w.Uint(uint64(p.Avail.To))
	}
}

func decodeIndexParams(r *wire.Reader) []index.Param {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[index.Param](r, n)
	for i := range out {
		p := &out[i]
		p.Name, p.Type, p.DocType = r.String(), r.String(), r.String()
		r.Bools(&p.Optional, &p.ByRef, &p.Variadic, &p.Promoted)
		p.Default, p.Out, p.Gen = r.String(), r.String(), r.String()
		p.Avail.From = phpversion.Version(r.Uint())
		p.Avail.To = phpversion.Version(r.Uint())
	}
	return out
}

func encodeSinkClass(w *wire.Writer, c *sinkClass) {
	w.Bool(c.Final)
	w.Uint(uint64(c.Kind))
	w.String(c.Modifiers)
	w.Strings(c.Extends)
	w.Strings(c.Traits)
	w.Len(len(c.Methods), c.Methods == nil)
	for name, m := range c.Methods {
		w.String(name)
		w.Int(int64(m.Count))
		w.String(m.Modifiers)
		encodeIndexParams(w, m.Parameters)
	}
}

func decodeSinkClass(r *wire.Reader) *sinkClass {
	c := &sinkClass{Final: r.Bool(), Kind: syntax.ClassKind(r.Uint()), Modifiers: r.String(), Extends: r.Strings(), Traits: r.Strings()}
	if n, isNil := r.Len(); !isNil {
		c.Methods = wire.MakeMap[string, sinkMethod](r, n)
		for range n {
			name := r.String()
			c.Methods[name] = sinkMethod{Count: int(r.Int()), Modifiers: r.String(), Parameters: decodeIndexParams(r)}
		}
	}
	return c
}

func encodeSinks(w *wire.Writer, sinks []Sink) {
	w.Len(len(sinks), sinks == nil)
	for i := range sinks {
		s := &sinks[i]
		w.String(s.Name)
		w.Uint(uint64(s.Context))
		encodeValue(w, &s.Argument)
		w.String(s.Path)
		w.Uint(uint64(s.Span.Start))
		w.Uint(uint64(s.Span.End))
	}
}

func decodeSinks(r *wire.Reader) []Sink {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[Sink](r, n)
	for i := range out {
		s := &out[i]
		s.Name = r.String()
		s.Context = Context(r.Uint())
		s.Argument = decodeValue(r)
		s.Path = r.String()
		s.Span = syntax.Span{Start: uint32(r.Uint()), End: uint32(r.Uint())}
	}
	return out
}

func encodeValues(w *wire.Writer, vs []Value) {
	w.Len(len(vs), vs == nil)
	for i := range vs {
		encodeValue(w, &vs[i])
	}
}

func decodeValues(r *wire.Reader) []Value {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := wire.Make[Value](r, n)
	for i := range out {
		out[i] = decodeValue(r)
	}
	return out
}

func encodeValue(w *wire.Writer, v *Value) {
	if !w.Enter() {
		return
	}
	defer w.Leave()
	if v.Expr != nil || v.tested != nil || v.unless != nil {
		w.Fail(errSyntax)
		return
	}
	w.Bools(v.record != nil, v.literal != nil, v.returned != nil, v.targets != nil, v.interval != nil, v.elements != nil, v.origins != nil,
		v.Complete, v.NonFalse, v.NonNull, v.LengthKnown, v.scalars, v.matched, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull, v.replaced, v.entry, v.plain, v.deep)
	if v.record != nil {
		encodeResource(w, v.record)
	}
	if v.literal != nil {
		encodeScalar(w, v.literal)
	}
	if v.returned != nil {
		w.Bool(v.returned.contract != nil)
		if v.returned.contract != nil {
			encodeReturned(w, v.returned.contract)
		}
		encodeValues(w, v.returned.outer)
	}
	if v.targets != nil {
		for _, item := range v.targets.items {
			if item != nil {
				w.Fail(errSyntax)
				return
			}
		}
		w.Len(len(v.targets.items), v.targets.items == nil)
	}
	if v.interval != nil {
		w.Int(v.interval.lo)
		w.Int(v.interval.hi)
	}
	w.Len(len(v.Sources), v.Sources == nil)
	for _, s := range v.Sources {
		w.String(s.Kind)
		w.String(s.Path)
		w.Uint(uint64(s.Span.Start))
		w.Uint(uint64(s.Span.End))
		w.Int(int64(s.Parameter))
		w.Bools(s.ValuesOnly, s.Keys, s.htmlDecoded)
	}
	if v.elements != nil {
		e := v.elements
		w.Bools(e.hasResources, e.whole, e.optional)
		w.Len(len(e.items), e.items == nil)
		for key, item := range e.items {
			w.String(key)
			encodeValue(w, &item)
		}
		w.Strings(e.order)
		w.Int(int64(e.depth))
		w.Int(int64(e.width))
	}
	if v.origins != nil {
		o := v.origins
		w.Bools(o.full, o.compressed, o.other)
		w.Len(len(o.items), o.items == nil)
		for _, item := range o.items {
			if item.Expr != nil {
				w.Fail(errSyntax)
				return
			}
			w.Bools(item.NonFalse, item.NonNull)
		}
	}
	w.Int(v.Length)
	w.Uint(uint64(v.Identity))
	w.Uint(uint64(v.storedIdentity))
	w.Uint(uint64(v.Invalidated))
	w.Uint(uint64(v.Safe))
	w.Uint(uint64(v.headerMask))
	w.Uint(uint64(v.upload))
	w.Uint(uint64(v.truth))
}

func decodeValue(r *wire.Reader) Value {
	var v Value
	if !r.Enter() {
		return v
	}
	defer r.Leave()
	var record, literal, returned, targets, interval, elements, origins bool
	r.Bools(&record, &literal, &returned, &targets, &interval, &elements, &origins,
		&v.Complete, &v.NonFalse, &v.NonNull, &v.LengthKnown, &v.scalars, &v.matched, &v.unlessFalse, &v.unlessNull, &v.skippedFalse, &v.skippedNull, &v.replaced, &v.entry, &v.plain, &v.deep)
	if record {
		v.record = decodeResource(r)
	}
	if literal {
		v.literal = decodeScalar(r)
	}
	if returned {
		v.returned = &callableValue{}
		if r.Bool() {
			v.returned.contract = decodeReturned(r)
		}
		v.returned.outer = decodeValues(r)
	}
	if targets {
		v.targets = &callableSet{}
		if n, isNil := r.Len(); !isNil {
			v.targets.items = wire.Make[syntax.Expr](r, n)
		}
	}
	if interval {
		v.interval = &intInterval{lo: r.Int(), hi: r.Int()}
	}
	if n, isNil := r.Len(); !isNil {
		v.Sources = wire.Make[Source](r, n)
		for i := range v.Sources {
			s := &v.Sources[i]
			s.Kind, s.Path = r.String(), r.String()
			s.Span = syntax.Span{Start: uint32(r.Uint()), End: uint32(r.Uint())}
			s.Parameter = int(r.Int())
			r.Bools(&s.ValuesOnly, &s.Keys, &s.htmlDecoded)
		}
	}
	if elements {
		e := &elementSet{}
		r.Bools(&e.hasResources, &e.whole, &e.optional)
		if n, isNil := r.Len(); !isNil {
			e.items = wire.MakeMap[string, Value](r, n)
			for range n {
				key := r.String()
				e.items[key] = decodeValue(r)
			}
		}
		e.order = r.Strings()
		e.depth = int(r.Int())
		e.width = int(r.Int())
		v.elements = e
	}
	if origins {
		o := &originSet{}
		r.Bools(&o.full, &o.compressed, &o.other)
		if n, isNil := r.Len(); !isNil {
			o.items = wire.Make[Origin](r, n)
			for i := range o.items {
				r.Bools(&o.items[i].NonFalse, &o.items[i].NonNull)
			}
		}
		v.origins = o
	}
	v.Length = r.Int()
	v.Identity = uint32(r.Uint())
	v.storedIdentity = uint32(r.Uint())
	v.Invalidated = uint32(r.Uint())
	v.Safe = Context(r.Uint())
	v.headerMask = uint8(r.Uint())
	v.upload = uint8(r.Uint())
	v.truth = uint8(r.Uint())
	return v
}
