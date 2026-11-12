package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

// Only resolved member metadata is cached. Snapshot reads and live receiver
// proofs stay at their existing call sites, and unresolved inference can refine.
type declarationCache struct {
	index    *index.Index
	revision uint64
	php      phpversion.Version
	members  map[syntax.Node]callDeclaration
}

type callDeclaration struct {
	params []index.Param
	name   string
}

func (e *Env) declaration(at syntax.Node) ([]index.Param, string, bool) {
	switch at.(type) {
	case *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
		if e.types != nil && e.types.Index != nil {
			cache := &e.declarations
			revision := e.types.Index.Revision()
			if cache.index != e.types.Index || cache.revision != revision || cache.php != e.types.PHP {
				*cache = declarationCache{index: e.types.Index, revision: revision, php: e.types.PHP}
			}
			if found, known := cache.members[at]; known {
				return found.params, found.name, true
			}
			params, name, known := e.lookupDeclaration(at)
			if known {
				if cache.members == nil {
					cache.members = map[syntax.Node]callDeclaration{}
				}
				cache.members[at] = callDeclaration{params: params, name: name}
			}
			return params, name, known
		}
	}
	return e.lookupDeclaration(at)
}

func (e *Env) lookupDeclaration(at syntax.Node) ([]index.Param, string, bool) {
	switch n := at.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil {
			return f.Params, names.Fold(f.FQN), true
		}
	case *syntax.MethodCall:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok {
			return nil, "", false
		}
		classes := e.types.TypeOf(n.Var).Classes()
		if len(classes) == 0 {
			if trait := e.traitReceiver(n); trait != "" {
				// The using class may override the method: the declaration
				// gives its parameters and a may-return, never a fixed body.
				if m := e.types.Index.FindMethod(trait, id.Value, e.types.PHP); m != nil && !m.Magic && !m.Abstract {
					return m.Params, strings.ToLower(m.Class + "::" + m.Name), true
				}
				return nil, "", false
			}
		}
		if len(classes) != 1 {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, strings.ToLower(m.Class + "::" + m.Name), true
		}
	case *syntax.StaticCall:
		id, ok := n.Name.(*syntax.Identifier)
		class, named := n.Class.(*syntax.Name)
		if !ok || !named {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, strings.ToLower(m.Class + "::" + m.Name), true
		}
	case *syntax.New:
		if class, ok := n.Class.(*syntax.Name); ok {
			fqn := e.types.Names.Class(class.Value, class.Span().Start)
			if m := e.types.Index.FindMethod(fqn, "__construct", e.types.PHP); m != nil {
				return m.Params, strings.ToLower(m.Class + "::__construct"), true
			}
		}
	}
	return nil, "", false
}

// traitReceiver names the trait whose instance method calls `$this->m()`
// directly. `$this` has no class there, but the trait's own (or used)
// non-abstract declarations are what the call reaches unless overridden.
// Closures are excluded: they may be bound to another object.
func (e *Env) traitReceiver(n *syntax.MethodCall) string {
	if n.NullSafe || plainVariable(n.Var) != "this" {
		return ""
	}
	// A method's `$this` has a class unless the method belongs to a trait.
	method, ok := syntax.EnclosingVariableScope(n).(*syntax.Method)
	if !ok || method.Modifiers.Has(syntax.TStatic) || syntax.EnclosingClass(method).ClassKind != syntax.KindTrait {
		return ""
	}
	return e.types.ClassFQN(syntax.EnclosingClass(method))
}

// dispatched names the summary of the method a call runs, the given one
// when the method is not resolved, and tells whether it is exactly known.
func (e *Env) dispatched(call syntax.Node, followed uint8, name string) (string, bool) {
	if m, sealed := e.dispatch(call, followed); m != nil {
		return strings.ToLower(m.Class + "::" + m.Name), sealed
	}
	return name, false
}

// dispatch resolves the method a method or static call names and tells
// whether that declaration is the one that runs: a final or private method,
// a method of a final class, a static call that names its class, as `self::`
// and `parent::` do too, a method called on an object made here by `new`, or
// a method no class of the project extending the receiver's replaces
// (`static::` and a method called on an object bind late). followed tells
// what is known of the receiver (see follows).
func (e *Env) dispatch(call syntax.Node, followed uint8) (*index.Method, bool) {
	m, receiver, sealed, open := e.resolve(call, followed)
	if m != nil && !sealed && open {
		sealed = !e.replaced(receiver, m)
	}
	return m, sealed
}

// resolve is dispatch without the classes of the project: sealed tells that
// the declaration is the one that runs whatever they declare, and open that
// it is unless one of those extending the receiver's class replaces it.
func (e *Env) resolve(call syntax.Node, followed uint8) (m *index.Method, receiver string, sealed, open bool) {
	// A method that is not visible from here is not the one that runs.
	m = e.method(call)
	if m == nil || !e.visible(call, m) {
		return nil, "", false, false
	}
	sealed = m.Final || m.Visibility == index.Private || followed == followExact
	// The class of the object the method is looked up on at run time is the
	// receiver's, the enclosing one for `$this` and `static`, or one that
	// extends it.
	within := syntax.EnclosingClass(call)
	receiver = e.types.ClassFQN(within)
	if static, ok := call.(*syntax.StaticCall); ok {
		class, _ := static.Class.(*syntax.Name)
		sealed = sealed || !strings.EqualFold(class.Value, "static")
	} else if typed := e.types.TypeOf(call.(*syntax.MethodCall).Var).Classes(); len(typed) == 1 {
		receiver, within = typed[0], nil
	}
	// A trait method runs in classes that are not known here.
	return m, receiver, sealed, followed != followLoose && (within == nil || within.ClassKind != syntax.KindTrait)
}

// What is known of the receiver of a method call.
const (
	// followExact: a variable that holds an object made by `new` of the
	// class inferred for it, on every path.
	followExact uint8 = iota + 1
	// followTyped: an object of the class inferred, or of one extending it.
	followTyped
	// followLoose: a variable that may hold something else.
	followLoose
)

// follows tells whether a receiver is the object inferred for it: a variable
// is only while the flow follows it, created here, received or produced by
// one expression every path agrees on, and bound to nothing. Any other
// receiver is what is declared or documented for it.
func (e *Env) follows(receiver syntax.Expr, f *frame) uint8 {
	name := plainVariable(receiver)
	if name == "" || name == "this" {
		return followTyped
	}
	held := f.vars.value(name)
	if _, shared := f.aliased[name]; shared || held.Identity == 0 && (held.Expr == nil || held.Invalidated != 0) {
		return followLoose
	}
	made, created := syntax.UnwrapParens(held.Expr).(*syntax.New)
	typed := e.types.TypeOf(receiver).Classes()
	if !created || held.Invalidated != 0 || len(typed) != 1 {
		return followTyped
	}
	if class, named := made.Class.(*syntax.Name); named && !strings.EqualFold(class.Value, "static") && strings.EqualFold(strings.TrimPrefix(e.named(class), `\`), strings.TrimPrefix(typed[0], `\`)) {
		return followExact
	}
	return followTyped
}

// Targets returns the declarations a method or static call may run: the one
// it resolves to when no other can (see dispatch), else that one and each
// method that replaces it in a class of the project extending the
// receiver's. Every class from the receiver's up to the one that declares
// the method, and every class extending it, must be declared once: of two
// declarations either may be the one in effect. Nothing is returned when
// the callee is not known that well. A method PHP itself declares only has
// to be the one no class of the project replaces.
func (e *Env) Targets(call syntax.Node) []*index.Method {
	return e.targets(call, e.reached(call))
}

// reached tells what the scope of a method call knows of its receiver (see
// follows): a variable is followed only in a scope analysed to its end.
func (e *Env) reached(call syntax.Node) uint8 {
	r := e.scope(syntax.EnclosingVariableScope(call))
	followed := followTyped
	if method, ok := call.(*syntax.MethodCall); ok && plainVariable(method.Var) != "" && plainVariable(method.Var) != "this" {
		if followed = r.receivers[call]; followed == 0 || !r.complete {
			followed = followLoose
		}
	}
	return followed
}

// targets is Targets for what is known of the receiver.
func (e *Env) targets(call syntax.Node, followed uint8) []*index.Method {
	m, receiver, sealed, open := e.resolve(call, followed)
	if m == nil || m.Magic || !sealed && !open {
		return nil
	}
	targets := []*index.Method{m}
	if !m.Builtin && !e.single(call, m) {
		if targets = e.copies(call, m); targets == nil {
			return nil
		}
	}
	if sealed {
		return targets
	}
	ix := e.types.Index
	if e.below(receiver, func(sub string) bool {
		other := ix.FindMethod(sub, m.Name, e.types.PHP)
		if ix.ClassCount(sub) != 1 || other == nil {
			return true
		}
		if !slices.Contains(targets, other) {
			targets = append(targets, other)
		}
		return false
	}) {
		return nil
	}
	return targets
}

// single reports that each class on the way from the one a call looks its
// method up on to the one that declares it is declared once: none of them
// may be another declaration, with another method.
func (e *Env) single(call syntax.Node, m *index.Method) bool {
	ix := e.types.Index
	from := e.lookedUp(call)
	for steps := 0; steps < MaxDepth*4 && ix.ClassCount(from) == 1; steps++ {
		class := ix.Class(from, e.types.PHP)
		if strings.EqualFold(strings.TrimPrefix(from, `\`), strings.TrimPrefix(m.Class, `\`)) || class.Parent == "" {
			// A method of a trait is declared outside the classes above.
			return ix.ClassCount(m.Class) == 1
		}
		from = class.Parent
	}
	return false
}

// lookedUp names the class a call looks its method up on.
func (e *Env) lookedUp(call syntax.Node) string {
	from := e.types.ClassFQN(syntax.EnclosingClass(call))
	switch n := call.(type) {
	case *syntax.New:
		from = e.named(n.Class.(*syntax.Name))
	case *syntax.StaticCall:
		from = e.named(n.Class.(*syntax.Name))
	case *syntax.MethodCall:
		if typed := e.types.TypeOf(n.Var).Classes(); len(typed) == 1 {
			from = typed[0]
		}
	}
	return from
}

// copies returns the method of each declaration of the class a call looks it
// up on, when the project declares that class more than once and every copy
// declares the method itself: any of them may be the one that runs.
func (e *Env) copies(call syntax.Node, m *index.Method) []*index.Method {
	from := e.lookedUp(call)
	if !strings.EqualFold(strings.TrimPrefix(from, `\`), strings.TrimPrefix(m.Class, `\`)) {
		return nil
	}
	// single found more than one declaration of the class.
	declarations := e.types.Index.ClassDecls(from, e.types.PHP)
	methods := make([]*index.Method, 0, len(declarations))
	for _, class := range declarations {
		own := class.Methods[strings.ToLower(m.Name)]
		if own == nil || own.Magic {
			return nil
		}
		methods = append(methods, own)
	}
	return methods
}

// replaced reports that a class of the project extending the receiver's
// declares, inherits or uses another method of that name than the one
// resolved. A class extends the receiver's through any of its declarations
// and through the names class_alias() gives it; one declared more than once
// may be either declaration. The classes the project does not contain
// cannot be seen: a method meant to be replaced outside it is not known to
// be. The answer is kept per receiver class and method.
func (e *Env) replaced(receiver string, m *index.Method) bool {
	return e.overrides(receiver, m.Name, m.Class)
}

// overrides is replaced for a method name and the class that declares the
// method resolved: none when the receiver's class has no such method, which
// a class extending it then must not have either.
func (e *Env) overrides(receiver, method, declaring string) bool {
	at := [2]string{strings.ToLower(receiver), strings.ToLower(declaring + "::" + method)}
	if known, ok := e.overridden[at]; ok {
		// The answer read the contracts of the subclasses: so does
		// whatever reads it again.
		e.depend(known.reads)
		return known.holds
	}
	if e.overridden == nil {
		e.overridden = map[[2]string]cachedAnswer{}
	}
	reads := e.begin()
	replaced := e.override(receiver, method, declaring)
	e.end(reads)
	e.overridden[at] = cachedAnswer{holds: replaced, reads: reads}
	return replaced
}

// cachedAnswer is a cached yes-or-no answer with the contracts computing it
// read: whatever reads the answer again depends on them too.
type cachedAnswer struct {
	holds bool
	reads map[string]bool
}

func (e *Env) override(receiver, method, declaring string) bool {
	ix := e.types.Index
	return e.below(receiver, func(sub string) bool {
		// The summaries read here depend on what each subclass declares.
		e.contract(strings.ToLower(sub) + "::" + strings.ToLower(method))
		other := ix.FindMethod(sub, method, e.types.PHP)
		return ix.ClassCount(sub) != 1 || (other == nil) != (declaring == "") || other != nil && !strings.EqualFold(other.Class, declaring)
	})
}

// below visits each class of the project that extends the receiver's, through
// any of its declarations and through the names class_alias() gives it, until
// found reports one; it tells whether one was. A final class has none. The
// receiver may be written with a leading backslash; the index names the
// classes extending it without one.
func (e *Env) below(receiver string, found func(sub string) bool) bool {
	ix := e.types.Index
	if declared := ix.Class(receiver, e.types.PHP); declared != nil && declared.Final {
		return false
	}
	for _, sub := range ix.Extending(receiver) {
		if found(sub) {
			return true
		}
	}
	return false
}
