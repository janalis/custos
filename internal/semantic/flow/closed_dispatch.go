package flow

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// Native receiver unions have closed targets when every method is immutable
// through inheritance, even in an open project.
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
	var out []closedContract
	for _, class := range classes {
		target, known := e.closedMethod(at, class, name.Value, false)
		if !known {
			return nil, false
		}
		out = append(out, target)
	}
	return out, true
}

// Resolve non-overridable methods using native receiver classes and record the
// entire declaration/visibility dependency chain, without PHPDoc narrowing.
func (e *Env) closedMethod(at syntax.Node, class, name string, static bool) (closedContract, bool) {
	if !e.sinkLexicalContext(at) {
		return closedContract{}, false
	}
	e.contract(e.classKey(class))
	decls := e.types.Index.ClassDecls(class, e.types.PHP)
	if len(decls) != 1 || decls[0].Kind != syntax.KindClass {
		return closedContract{}, false
	}
	method := e.types.Index.FindMethod(class, name, e.types.PHP)
	if method == nil || method.Magic || method.Builtin || !e.visible(at, method) || static && !method.Static || !decls[0].Final && !method.Final && method.Visibility != index.Private {
		return closedContract{}, false
	}
	owner := class
	found := false
	for range MaxDepth * 4 {
		meta, known := e.contract(e.classKey(owner))
		decls := e.types.Index.ClassDecls(owner, e.types.PHP)
		if !known || meta.dispatch == nil || len(decls) != 1 || decls[0].Kind == syntax.KindTrait {
			return closedContract{}, false
		}
		if strings.EqualFold(strings.TrimPrefix(owner, `\`), strings.TrimPrefix(method.Class, `\`)) {
			if meta.dispatch.Methods[strings.ToLower(method.Name)].Count != 1 {
				return closedContract{}, false
			}
			found = true
			break
		}
		if owner = decls[0].Parent; owner == "" {
			return closedContract{}, false
		}
	}
	if !found {
		return closedContract{}, false
	}
	summary, known := e.contract(strings.ToLower(method.Class + "::" + method.Name))
	if !known || !summary.Fixed || summary.Generator {
		return closedContract{}, false
	}
	return closedContract{method, summary}, true
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
