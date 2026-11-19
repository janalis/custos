package flow

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
)

// ReferenceResult returns what a fixed project callable leaves in an argument.
// Sources include inputs retained on any path; Expr is present only for a
// single reaching expression from the local reference channel.
// Unpacked, variadic and unresolved bindings cannot establish this proof.
// Multiple reference outputs require distinct, unaliased local argument cells.
// A callee that may modify request globals can also write the actual argument
// through another route; its parameter-only summary cannot prove this result.
func (e *Env) ReferenceResult(call syntax.Node, target *syntax.Arg) (Value, bool) {
	list := arguments(call)
	if list == nil || target == nil {
		return Value{}, false
	}
	if method, ok := call.(*syntax.MethodCall); ok {
		native := e.types.Native().TypeOf(method.Var)
		if method.NullSafe || len(native.Classes()) != 1 || len(native.Atoms()) != 1 {
			return Value{}, false
		}
	}
	if static, ok := call.(*syntax.StaticCall); ok {
		name, named := static.Class.(*syntax.Name)
		if !named || strings.EqualFold(name.Value, "static") {
			return Value{}, false
		}
	}
	params, name, known := e.declaration(call)
	if _, function := call.(*syntax.FuncCall); !function && (e.sinkSummaryName(call) != name || e.reached(call) == followLoose) {
		return Value{}, false
	}
	summary, described := e.contract(name)
	if !known || !described || !summary.Fixed || !summary.Spares {
		return Value{}, false
	}
	r := e.scope(syntax.EnclosingVariableScope(call))
	if !r.complete {
		return Value{}, false
	}
	if len(list.Args) > 256 || len(params) > 64 {
		r.gap("argument-budget", call.Span())
		return Value{}, false
	}
	var few [8]int
	positions := parameters(list, params, few[:0])
	at, references := -1, 0
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		p := positions[i]
		if !ok || arg.Unpack || arg.ByRef || p < 0 || params[p].Variadic {
			return Value{}, false
		}
		if _, evaluated := r.values.get(arg.Value); !evaluated {
			return Value{}, false
		}
		if params[p].ByRef {
			references++
		}
		if arg == target {
			at = i
		}
	}
	if at < 0 || references != 1 && !r.distinctReferences[call] || !params[positions[at]].ByRef {
		return Value{}, false
	}
	if _, described := summary.ByRef[positions[at]]; !described {
		return Value{}, false
	}
	v := e.left(call, at, list, positions, r.values.at(target.Value), nil, r)
	if body := e.localBody(name); body != nil {
		p := syntax.VariableScopeParams(body)[positions[at]]
		v.Expr = e.scope(body).exit.value(received(p.Var.Name)).Expr
	}
	return v, true
}
