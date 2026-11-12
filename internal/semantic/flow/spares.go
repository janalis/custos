package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// Spares reports a written call or a construction that stores nothing in a
// request table. A PHP function or method does not, unless it runs code of
// the project: it is given a function to call, loads or evaluates code,
// resumes or invokes some (see runner, loads, invokes and builds), or is a
// method, not final, of a class the project extends. A call given an array
// that shares a variable (`[&$x]`) may store anything; `f(...)` calls
// nothing. A
// callable of the project does not when it is the declaration that runs, or
// each of those that may (see Targets), declared once, takes no argument by
// reference, and its body spares the tables too and calls nothing of the
// project (see Summary.Spares). A
// construction runs the constructor and, once the object is dropped, the
// destructor the class finds: none is none. Anything not known that well may
// store request text anywhere. Code that runs without a call being written
// (a conversion to text, a magic method, an autoloader) is not counted.
func (e *Env) Spares(call syntax.Node) bool {
	return e.spares(call, e.reached(call), false)
}

// spares is Spares for what is known of the receiver. within tells that the
// call is written in a body being summarized: a callable of the project is
// then not looked into, so that a summary depends on no other.
func (e *Env) spares(call syntax.Node, followed uint8, within bool) bool {
	ix := e.types.Index
	if list := arguments(call); list != nil {
		if len(list.Args) == 1 {
			if _, callable := list.Args[0].(*syntax.VariadicPlaceholder); callable {
				// `f(...)` only makes a closure: nothing is called.
				return true
			}
		}
		if shares(list) {
			return false
		}
	}
	switch n := call.(type) {
	case *syntax.FuncCall:
		f := e.types.ResolveFunction(n)
		if f == nil || !f.Avail.In(e.types.PHP) {
			return false
		}
		if f.Builtin {
			name := strings.ToLower(f.FQN)
			return name != "create_function" && !runner(name) && !loads(name, n.Args) && !e.callbackGiven(n.Args, f.Params)
		}
		if within {
			return e.nestedSpares(f)
		}
		summary, ok := e.contract(strings.ToLower(f.FQN))
		return ok && summary.Spares && !referenced(f.Params) && len(ix.FunctionDecls(f.FQN, e.types.PHP)) == 1
	case *syntax.MethodCall:
		return e.sparing(e.targets(call, followed), n.Args, within)
	case *syntax.StaticCall:
		return e.sparing(e.targets(call, followed), n.Args, within)
	case *syntax.New:
		class, named := n.Class.(*syntax.Name)
		if !named || strings.EqualFold(class.Value, "static") {
			return false
		}
		fqn := e.named(class)
		declared := ix.Class(fqn, e.types.PHP)
		if declared == nil || ix.ClassCount(fqn) != 1 {
			return false
		}
		short := fqn[strings.LastIndexByte(fqn, '\\')+1:]
		for _, name := range []string{"__construct", "__destruct", short} {
			m := ix.FindMethod(fqn, name, e.types.PHP)
			switch {
			case m == nil:
			case name == short:
				// Before PHP 8 a method named like its class constructs it.
				if !e.types.PHP.AtLeast(phpversion.PHP80) {
					return false
				}
			case !m.Builtin && !e.single(call, m), !e.sparing([]*index.Method{m}, n.Args, within):
				return false
			}
		}
		return true
	}
	return false
}

// sparing reports that none of the declarations a call may run stores
// anything in a request table, and that one of them at least has a body.
func (e *Env) sparing(targets []*index.Method, list *syntax.ArgList, within bool) bool {
	bodies := 0
	for _, m := range targets {
		switch class := strings.ToLower(strings.TrimPrefix(m.Class, `\`)); {
		case referenced(m.Params):
			return false
		case m.Abstract:
			continue
		case !m.Builtin && within:
			return false
		case !m.Builtin:
			if summary, ok := e.contract(strings.ToLower(m.Class + "::" + m.Name)); !ok || !summary.Spares {
				return false
			}
		case class == "closure" || class == "generator" || class == "fiber" || strings.HasPrefix(class, "reflection") || e.callbackGiven(list, m.Params):
			return false
		case strings.EqualFold(m.Name, "__construct"):
			// The object made is of that very class.
			if builds(class, list) {
				return false
			}
		case invokes(class+"::"+strings.ToLower(m.Name), list) || !m.Final && e.extended(m.Class):
			return false
		}
		bodies++
	}
	return bodies > 0
}

// shares reports arguments that hold an array written with an element by
// reference (`[&$_GET]`): the callee then holds what the caller's variable
// holds, and writes it without a parameter by reference.
func shares(list *syntax.ArgList) bool {
	shared := false
	syntax.Inspect(list, func(n syntax.Node) bool {
		item, ok := n.(*syntax.ArrayItem)
		shared = shared || ok && item.ByRef
		return !shared
	})
	return shared
}

// extended reports a class of PHP that a class of the project extends: a
// method PHP declares may then call one the project replaces, as the
// insertion into a heap calls its comparison. The answer is kept per class.
func (e *Env) extended(class string) bool {
	at := strings.ToLower(strings.TrimPrefix(class, `\`))
	if known, ok := e.extendedBy[at]; ok {
		return known
	}
	if e.extendedBy == nil {
		e.extendedBy = map[string]bool{}
	}
	ix := e.types.Index
	e.extendedBy[at] = e.below(class, func(sub string) bool {
		declared := ix.Class(sub, e.types.PHP)
		return declared != nil && !strings.HasPrefix(declared.File, "stubs/")
	})
	return e.extendedBy[at]
}

// referenced reports a parameter taken by reference: the callee writes to
// what the caller hands it.
func referenced(params []index.Param) bool {
	return slices.ContainsFunc(params, func(p index.Param) bool { return p.ByRef })
}

// callbackGiven reports an argument that may be a function for PHP to call:
// one bound to a parameter that takes one, unless it names a PHP function
// that only computes a number, a boolean or text, or one whose parameter is
// not known.
func (e *Env) callbackGiven(list *syntax.ArgList, params []index.Param) bool {
	if list == nil {
		return false
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || invoked(params[at]) && !e.harmless(arg.Value) {
			return true
		}
	}
	return false
}

// sparesTables reports a callable whose body is the one every call runs and
// stores nothing in a request table: it names no request table, no
// `$GLOBALS` and no variable by a name computed at run time, evaluates and
// includes no code, and all it calls or constructs, in the default values
// of its parameters and in the functions it declares inline too, is PHP's own and
// spares them (see Spares), or a class of the project without constructor
// and destructor. A callable of the project it calls is not looked into: one
// summary then depends on no other, and none has to be computed again for
// this. The receivers of the calls in the inline functions are not followed.
func (e *Env) sparesTables(scope, body syntax.Node) bool {
	if body == nil || !unconditional(scope) {
		return false
	}
	spared := true
	look := func(n syntax.Node) bool {
		if !spared {
			// What is written after a refill does not undo it.
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.ClassLike:
			// Declared here, run where it is called or constructed.
			return false
		case *syntax.Variable:
			// A name computed at run time may be the one of a request table.
			spared = x.NameExpr == nil && !external(x.Name) && x.Name != "GLOBALS"
		case *syntax.Eval, *syntax.Include:
			spared = false
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			followed := followLoose
			if syntax.EnclosingVariableScope(n) == scope {
				followed = e.reached(n)
			}
			spared = e.spares(n, followed, true)
		}
		return spared
	}
	// The value a parameter takes when none is given is computed by each
	// call that leaves it out.
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.Default != nil {
			syntax.Inspect(p.Default, look)
		}
	}
	syntax.Inspect(body, look)
	return spared
}

// maxNested bounds how many calls deep nestedSpares looks.
const maxNested = 4

// nestedSpares looks into a function of the file that a body being
// summarized calls, so that a summary still depends on no other: the function
// must be declared once, unconditionally, take nothing by reference, and its
// own body must spare the tables, a few calls deep. A call that comes back
// to a body being looked into is taken to fill them.
func (e *Env) nestedSpares(f *index.Function) bool {
	declaration := e.localBody(strings.ToLower(strings.TrimPrefix(f.FQN, `\`)))
	if declaration == nil || referenced(f.Params) || len(e.types.Index.FunctionDecls(f.FQN, e.types.PHP)) != 1 || len(e.nested) >= maxNested || e.nested[declaration] {
		return false
	}
	if e.nested == nil {
		e.nested = map[syntax.Node]bool{}
	}
	e.nested[declaration] = true
	defer delete(e.nested, declaration)
	return e.sparesTables(declaration, syntax.VariableScopeBody(declaration))
}
