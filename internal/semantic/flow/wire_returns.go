package flow

import (
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/types"
	"github.com/janalis/custos/internal/semantic/wire"
)

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
		w.String(p.TypeVer)
		w.String(p.DocType)
		w.String(p.DocValues)
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
		p.Name, p.Type, p.TypeVer, p.DocType, p.DocValues = r.String(), r.String(), r.String(), r.String(), r.String()
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
