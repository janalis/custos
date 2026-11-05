package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func (e *Env) captureCallable(at syntax.Node, receiver Value, f *frame, r *scopeResult) {
	if e.types.PHP < phpversion.PHP81 {
		return
	}
	call, ok := at.(*syntax.FuncCall)
	if !ok {
		target, known := e.captureMethod(at, receiver, f)
		if known {
			if r.boundMethods == nil {
				r.boundMethods = map[syntax.Expr]*boundMethod{}
			}
			r.boundMethods[at.(syntax.Expr)] = target
		}
		return
	}
	function := e.types.ResolveFunction(call)
	if function == nil || function.ByRef || !function.Avail.In(e.types.PHP) {
		return
	}
	if r.callables == nil {
		r.callables = map[syntax.Expr]string{}
	}
	r.callables[call] = strings.ToLower(function.FQN)
}

func (e *Env) invokeCallable(at syntax.Node, values []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	call, ok := at.(*syntax.FuncCall)
	if !ok {
		return Value{}, nil, false
	}
	callee := r.values.at(call.Name)
	list := call.Args
	if name, builtin := e.callName(call); builtin {
		switch name {
		case "call_user_func":
			bound, known := e.bind(call, values)
			if !known || len(bound) == 0 {
				return Value{}, nil, false
			}
			callee = bound[0]
			forwarded := &syntax.ArgList{}
			var forwardedValues []Value
			for i, node := range list.Args {
				arg, ok := node.(*syntax.Arg)
				if !ok || arg.Unpack {
					return Value{}, nil, false
				}
				if arg.Name == nil && i == 0 || arg.Name != nil && arg.Name.Value == "callback" {
					continue
				}
				forwarded.Args = append(forwarded.Args, node)
				forwardedValues = append(forwardedValues, values[i])
			}
			list, values = forwarded, forwardedValues
		case "call_user_func_array":
			args, known := e.types.Native().BuiltinArguments(call)
			if !known || len(args) != 2 {
				return Value{}, nil, false
			}
			callee = r.values.at(args[0])
			packed := r.values.at(args[1])
			packed.elements = e.elements(args[1], f, r)
			list = &syntax.ArgList{Args: []syntax.Expr{&syntax.Arg{Value: args[1], Unpack: true}}}
			values = []Value{packed}
		}
	}
	if !callee.Complete || callee.replaced || callee.Invalidated != 0 {
		return Value{}, nil, false
	}
	if callee.returned != nil {
		return e.invokeReturned(at, callee.returned, list, values, r)
	}
	if callee.targets == nil {
		if target, known := e.invokableObject(call.Name, callee, at, f); known {
			return e.invokeBoundMethod(at, target, list, values, f, r)
		}
	}
	return e.invokeCallableSet(at, callee, list, values, f, r)
}

func (e *Env) invokeCallableTarget(at syntax.Node, expr syntax.Expr, list *syntax.ArgList, values []Value, f *frame, r *scopeResult) (Value, map[uint32]State, bool) {
	expr = syntax.UnwrapParens(expr)
	if captured, known := r.closures[expr]; known {
		return e.invokeClosure(at, expr, captured, list, values, f, r)
	}
	if method := r.boundMethods[expr]; method != nil {
		return e.invokeBoundMethod(at, method, list, values, f, r)
	}
	if array, ok := expr.(*syntax.Array); ok {
		if target, known := e.arrayMethodCallable(at, array, f, r); known {
			return e.invokeBoundMethod(at, target, list, values, f, r)
		}
		return Value{}, nil, false
	}
	name := r.callables[expr]
	if name == "" {
		if literal, known := literalText(expr); known {
			name = strings.TrimPrefix(literal, `\`)
		}
	}
	function := e.types.Index.Function(name, e.types.PHP)
	if function == nil || function.ByRef {
		return Value{}, nil, false
	}
	if function.Builtin {
		return e.invokeNativeCallable(at, list, values, function.FQN, f, r)
	}
	summary, known := e.contract(strings.ToLower(function.FQN))
	if !known || !summary.Fixed || summary.Generator || !summary.Return.Complete {
		return Value{}, nil, false
	}
	list, values, _ = userArguments(list, function.Params, values, summary.Quiet && !summary.argumentSensitive)
	args, bound := e.bindTo(list, function.Params, values)
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
	identity, states, known := e.instantiateResource(at, summary.resource, args, f, r)
	if known {
		result.Identity = identity
	}
	result, records := e.instantiateRecord(at, result, args, f, r)
	for id, state := range records {
		if states == nil {
			states = map[uint32]State{}
		}
		states[id] = state
	}
	effects := e.resourceEffects(at, summary.effects, args, f, r)
	for id, state := range effects {
		if states == nil {
			states = map[uint32]State{}
		}
		states[id] = state
	}
	return result, states, true
}
