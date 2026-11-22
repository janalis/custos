package flow

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/names"
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
					return m.Params, m.Key(), true
				}
				return nil, "", false
			}
		}
		if len(classes) != 1 {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, m.Key(), true
		}
	case *syntax.StaticCall:
		id, ok := n.Name.(*syntax.Identifier)
		class, named := n.Class.(*syntax.Name)
		if !ok || !named {
			return nil, "", false
		}
		if m := e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP); m != nil && !m.Magic {
			return m.Params, m.Key(), true
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
