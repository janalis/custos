package flow

import (
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/types"
	"github.com/janalis/custos/internal/semantic/wire"
)

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
