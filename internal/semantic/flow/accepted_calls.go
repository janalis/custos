package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// accepted records what a call proves by returning. Since PHP 8 a PHP
// function throws a TypeError for false or null given to a parameter that
// only takes arrays, objects or callables (booleans and null still convert
// to scalar parameters). A function of the project, or a static method
// named by its class, `self` or `parent`, throws one for null given to a
// required parameter whose declared type does not admit it, as does a method
// called on an object when it is the one resolved (see dispatch), since PHP 7:
// before, the refusal is an error a handler can recover from, and a scalar
// type names a class. The proof holds where that error cannot be caught:
// outside any `try` with a `catch`.
func (e *Env) accepted(call syntax.Node, list *syntax.ArgList, builtin bool, followed uint8, f *frame) {
	switch n := call.(type) {
	case *syntax.FuncCall:
		// One of two declarations of a name may be the one in effect.
		if f := e.types.ResolveFunction(n); !builtin && (f == nil || len(e.types.Index.FunctionDecls(f.FQN, e.types.PHP)) != 1) {
			return
		}
	case *syntax.StaticCall:
		// A method bound late, or called on an object, may be one that
		// declares a wider type; one that is not visible from here leaves
		// the call to `__callStatic`.
		class, named := n.Class.(*syntax.Name)
		m := e.method(call)
		if builtin || !named || m == nil || strings.EqualFold(class.Value, "static") {
			return
		}
		if !e.visible(call, m) || e.types.Index.ClassCount(m.Class) != 1 {
			return
		}
	case *syntax.MethodCall:
		// A method called on an object is the one resolved when no class
		// can replace it; a null-safe call may not be made.
		if builtin || conditional(n) || !e.provable(call, list, f) {
			return
		}
		if m, exact := e.dispatch(call, followed); m == nil || !exact || !e.single(call, m) {
			return
		}
	}
	if builtin && !e.types.PHP.AtLeast(phpversion.PHP80) || !e.types.PHP.AtLeast(phpversion.PHP70) {
		return
	}
	params, _, _ := e.declaration(call)
	for p := call.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if guarded, ok := p.(*syntax.Try); ok && len(guarded.Catches) > 0 {
			return
		}
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || params[at].ByRef || params[at].Variadic {
			continue
		}
		// A later argument that assigns the variable replaces what this
		// parameter received.
		written := false
		for j, other := range list.Args {
			if rest, ok := other.(*syntax.Arg); ok && j != i {
				written = written || e.rebinds(rest.Value, arg.Value, f)
			}
		}
		if written {
			continue
		}
		if compound, nullable := compoundOnly(params[at].Type); builtin && compound {
			exclude(arg.Value, f, true, !nullable)
		} else if resource, nullable := resourceOnly(params[at]); builtin && resource {
			excludeValue(arg.Value, f, true, !nullable)
		} else if !builtin && !params[at].Optional && rejectsNull(params[at].Type) {
			exclude(arg.Value, f, false, true)
		}
	}
}

// provable reports a call with something to prove: an argument that is a
// variable not yet known to differ from null, bound to a required parameter
// that does not admit null. Only then is what runs worth resolving.
func (e *Env) provable(call syntax.Node, list *syntax.ArgList, f *frame) bool {
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || params[at].Optional || !rejectsNull(params[at].Type) {
			continue
		}
		if variable := guarded(arg.Value); variable != nil {
			if held, tracked := f.vars.lookup(variable.Name); tracked && !held.NonNull {
				return true
			}
		}
	}
	return false
}

// rejectsNull reports a declared parameter type that does not admit null.
func rejectsNull(declared string) bool {
	if declared == "" || strings.HasPrefix(declared, "?") {
		return false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(strings.TrimPrefix(part, `\`)) {
		case "null", "mixed":
			return false
		}
	}
	return true
}

// resourceOnly reports a builtin parameter PHP checks as a resource (no
// declared type, documented `resource` in the stubs: `fread($stream)`),
// and whether it also takes null.
func resourceOnly(p index.Param) (bool, bool) {
	if p.Type != "" || p.DocType == "" {
		return false, false
	}
	resource, nullable := false, false
	for _, part := range strings.Split(p.DocType, "|") {
		switch strings.ToLower(strings.TrimPrefix(part, "?")) {
		case "resource":
			resource = true
		case "null":
			nullable = true
		default:
			return false, false
		}
		nullable = nullable || strings.HasPrefix(part, "?")
	}
	return resource, nullable
}

// compoundOnly reports a declared parameter type made only of arrays,
// iterables, callables, objects and classes, and whether it also takes null.
func compoundOnly(declared string) (bool, bool) {
	nullable := strings.HasPrefix(declared, "?")
	declared = strings.TrimPrefix(declared, "?")
	if declared == "" {
		return false, false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch name := strings.ToLower(strings.TrimPrefix(part, `\`)); name {
		case "null":
			nullable = true
		case "array", "iterable", "callable", "object":
		case "int", "float", "string", "bool", "false", "true", "mixed", "resource", "void", "never", "static", "self", "":
			return false, false
		default:
			// A class (GdImage, CurlHandle, Countable): PHP's own functions
			// throw for false given to it since PHP 8. Doc-only spellings
			// (`positive-int`, `list<int>`) are no class names.
			if strings.ContainsFunc(name, func(r rune) bool {
				return !(r == '_' || r == '\\' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= 0x80)
			}) {
				return false, false
			}
		}
	}
	return true, nullable
}
