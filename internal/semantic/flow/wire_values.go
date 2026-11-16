package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/wire"
)

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
