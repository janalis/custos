package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

type boundMethod struct {
	name     string
	params   []index.Param
	receiver Value
}

func (e *Env) captureMethod(at syntax.Node, receiver Value, f *frame) (*boundMethod, bool) {
	var class, name string
	static := false
	switch n := at.(type) {
	case *syntax.MethodCall:
		id, known := n.Name.(*syntax.Identifier)
		if !known || n.NullSafe || !receiver.Complete && plainVariable(n.Var) != "this" || receiver.replaced || receiver.Invalidated != 0 || e.follows(n.Var, f) == followLoose {
			return nil, false
		}
		native := e.types.Native().TypeOf(n.Var)
		if len(native.Classes()) != 1 || len(native.Atoms()) != 1 {
			return nil, false
		}
		class, name = native.Classes()[0], id.Value
	case *syntax.StaticCall:
		cls, named := n.Class.(*syntax.Name)
		id, known := n.Name.(*syntax.Identifier)
		if !named || !known || strings.EqualFold(cls.Value, "static") {
			return nil, false
		}
		class, name, static = e.named(cls), id.Value, true
	default:
		return nil, false
	}
	target, known := e.closedMethod(at, class, name, static)
	if !known || target.method.ByRef {
		return nil, false
	}
	return &boundMethod{name: strings.ToLower(target.method.Class + "::" + target.method.Name), params: target.method.Params, receiver: receiver}, true
}

func (e *Env) arrayMethodCallable(at syntax.Node, array *syntax.Array, f *frame, r *scopeResult) (*boundMethod, bool) {
	if len(array.Items) != 2 {
		return nil, false
	}
	var objectExpr, methodExpr syntax.Expr
	next := 0
	for _, item := range array.Items {
		if item == nil || item.ByRef || item.Unpack {
			return nil, false
		}
		key := strconv.Itoa(next)
		if item.Key != nil {
			var valid bool
			key, valid = syntax.LiteralArrayKey(item.Key)
			if !valid {
				return nil, false
			}
		}
		if key == "0" {
			objectExpr = item.Value
		}
		if key == "1" {
			methodExpr = item.Value
		}
		if n, err := strconv.Atoi(key); err == nil && n >= next {
			next = n + 1
		}
	}
	if objectExpr == nil || methodExpr == nil {
		return nil, false
	}
	object, method := r.values[objectExpr], r.values[methodExpr]
	name, known := literalText(method.Expr)
	if !object.Complete || object.replaced || object.Invalidated != 0 || !method.Complete || !known {
		return nil, false
	}
	class, static := literalText(object.Expr)
	if c, ok := syntax.UnwrapParens(objectExpr).(*syntax.ClassConstFetch); ok && className(c) {
		if name, known := c.Class.(*syntax.Name); known && !strings.EqualFold(name.Value, "static") {
			class, static = e.named(name), true
		}
	}
	if !static {
		typ := e.types.Native().TypeOf(objectExpr)
		if len(typ.Classes()) != 1 || len(typ.Atoms()) != 1 {
			return nil, false
		}
		class = typ.Classes()[0]
	}
	target, resolved := e.closedMethod(at, strings.TrimPrefix(class, `\`), name, static)
	if !resolved || target.method.ByRef {
		return nil, false
	}
	return &boundMethod{name: strings.ToLower(target.method.Class + "::" + target.method.Name), params: target.method.Params, receiver: object}, true
}

func (e *Env) invokeBoundMethod(at syntax.Node, target *boundMethod, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	summary, known := e.contract(target.name)
	if !known || !summary.Fixed || summary.Generator || !summary.Return.Complete {
		return Value{}, nil, false
	}
	args, bound := e.bindTo(list, target.params, values)
	if !bound {
		return Value{}, nil, false
	}
	result := substitute(summary.Return, args, true)
	result.returned = bindReturned(summary.returned, args)
	var sinks []Sink
	for _, sink := range summary.Sinks {
		sink.Argument = substitute(sink.Argument, args, true)
		sinks = appendSinks(sinks, []Sink{sink})
	}
	if r.closureSinks == nil {
		r.closureSinks = map[syntax.Node][]Sink{}
	}
	r.closureSinks[at] = appendSinks(r.closureSinks[at], sinks)
	identity, states, resource := e.instantiateResource(at, summary.resource, args, f, r)
	if resource {
		result.Identity = identity
	}
	if target.receiver.Identity != 0 {
		if states == nil {
			states = map[uint32]State{}
		}
		states[target.receiver.Identity] = State{}
	}
	return result, states, true
}

func (e *Env) invokableObject(expr syntax.Expr, receiver Value, at syntax.Node, f *frame) (*boundMethod, bool) {
	if !receiver.Complete || receiver.replaced || receiver.Invalidated != 0 || e.follows(expr, f) == followLoose {
		return nil, false
	}
	typ := e.types.Native().TypeOf(expr)
	if len(typ.Classes()) != 1 || len(typ.Atoms()) != 1 {
		return nil, false
	}
	target, known := e.closedMethod(at, typ.Classes()[0], "__invoke", false)
	if !known || target.method.ByRef {
		return nil, false
	}
	return &boundMethod{name: strings.ToLower(target.method.Class + "::" + target.method.Name), params: target.method.Params, receiver: receiver}, true
}
