package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// A native union of final classes is a closed set even in an open project.
// Limit fanout independently of class hierarchy and summary recursion budgets.
const maxClosedTargets = 8

type closedContract struct {
	method  *index.Method
	summary Summary
}

func (e *Env) closedContracts(at syntax.Node) ([]closedContract, bool) {
	call, ok := at.(*syntax.MethodCall)
	if !ok {
		return nil, false
	}
	name, named := call.Name.(*syntax.Identifier)
	if !named {
		return nil, false
	}
	typeOf := e.types.Native().TypeOf(call.Var)
	classes := typeOf.Classes()
	if len(classes) < 2 || len(classes) > maxClosedTargets || len(classes) != len(typeOf.Atoms()) {
		return nil, false
	}
	if !e.sinkLexicalContext(at) {
		return nil, false
	}
	var out []closedContract
	for _, class := range classes {
		// Both positive and missing metadata are dependency reads. Removing a
		// class or changing its modifiers must invalidate composed summaries.
		e.contract(sinkClassKey(class))
		decls := e.types.Index.ClassDecls(class, e.types.PHP)
		if len(decls) != 1 || !decls[0].Final || decls[0].Kind != syntax.KindClass {
			return nil, false
		}
		method := e.types.Index.FindMethod(class, name.Value, e.types.PHP)
		if method == nil || method.Magic || method.Builtin || !e.visible(at, method) {
			return nil, false
		}
		owner := class
		found := false
		for range MaxDepth * 4 {
			meta, known := e.contract(sinkClassKey(owner))
			decls := e.types.Index.ClassDecls(owner, e.types.PHP)
			if !known || meta.dispatch == nil || len(decls) != 1 || decls[0].Kind == syntax.KindTrait {
				return nil, false
			}
			if strings.EqualFold(strings.TrimPrefix(owner, `\`), strings.TrimPrefix(method.Class, `\`)) {
				if meta.dispatch.Methods[strings.ToLower(method.Name)].Count != 1 {
					return nil, false
				}
				found = true
				break
			}
			if owner = decls[0].Parent; owner == "" {
				return nil, false
			}
		}
		if !found {
			return nil, false
		}
		summary, known := e.contract(strings.ToLower(method.Class + "::" + method.Name))
		if !known || !summary.Fixed || summary.Generator {
			return nil, false
		}
		out = append(out, closedContract{method, summary})
	}
	return out, true
}

func (e *Env) closedReturn(at syntax.Node, values []Value, followed uint8) (Value, bool) {
	if followed == followLoose {
		return Value{}, false
	}
	contracts, known := e.closedContracts(at)
	if !known {
		return Value{}, false
	}
	call := at.(*syntax.MethodCall)
	var result Value
	for i, contract := range contracts {
		args, known := e.bindTo(call.Args, contract.method.Params, values)
		if !known || !contract.summary.Return.Complete {
			return Value{}, false
		}
		value := substitute(contract.summary.Return, args, true)
		if i == 0 {
			result = value
		} else {
			result = merge(result, value)
		}
	}
	return result, true
}

func (e *Env) closedSinks(at syntax.Node) []Sink {
	contracts, known := e.closedContracts(at)
	if !known {
		return nil
	}
	r := e.scope(syntax.EnclosingVariableScope(at))
	if e.reached(at) == followLoose && (!r.complete || !r.closedReceivers[at]) {
		return nil
	}
	call := at.(*syntax.MethodCall)
	for _, called := range r.calls {
		if called.Node != at {
			continue
		}
		var sinks []Sink
		for _, contract := range contracts {
			args, known := e.bindTo(call.Args, contract.method.Params, called.Arguments)
			if !known {
				return nil
			}
			for _, sink := range contract.summary.Sinks {
				sink.Argument = substitute(sink.Argument, args, true)
				sinks = appendSinks(sinks, []Sink{sink})
			}
		}
		return sinks
	}
	return nil
}

// Native type inference may retain a finite union while the value lattice has
// no single identity for it. A complete unescaped binding still carries that
// union; this exception is used only after closedContracts verifies every type.
func closedBinding(receiver syntax.Expr, f *frame) bool {
	name := plainVariable(receiver)
	value, tracked := f.vars[name]
	_, aliased := f.aliased[name]
	return tracked && !aliased && value.Complete && !value.replaced && value.Invalidated == 0
}
