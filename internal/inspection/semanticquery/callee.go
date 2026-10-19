package semanticquery

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// Callee is the agreed parameter contract of a call or construction.
type Callee struct {
	Params []index.Param
	ByRef  bool
	// Builtin tells that PHP itself declares every declaration resolved: its
	// documented parameter types then describe PHP's own checks.
	Builtin bool
}

// ResolveCallee resolves a call or construction to its parameter contract.
// Each inferred receiver class must have the same contract. This query does
// not prove exact runtime dispatch; consumers needing that use flow targets.
func ResolveCallee(ctx *analysis.Context, call syntax.Node) (Callee, bool) {
	return calleeParameters(ctx.Types(), call)
}

func calleeParameters(env *infer.Env, call syntax.Node) (Callee, bool) {
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := env.ResolveFunction(c); f != nil && f.Avail.In(env.PHP) {
			if !f.Builtin {
				declarations := env.Index.FunctionDecls(f.FQN, env.PHP)
				if len(declarations) > 64 {
					return Callee{}, false
				}
				for _, other := range declarations {
					if other.ByRef != f.ByRef || !slices.Equal(other.Params, f.Params) {
						return Callee{}, false
					}
				}
			}
			return Callee{Params: f.Params, ByRef: f.ByRef, Builtin: f.Builtin}, true
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return Callee{}, false
		}
		typ := env.TypeOf(c.Var)
		if c.NullSafe {
			typ = typ.Without("null")
		}
		classes := typ.Classes()
		if len(classes) != len(typ.Atoms()) || len(classes) == 0 {
			return Callee{}, false
		}
		var contract Callee
		for i, cls := range classes {
			m := env.Index.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, env.PHP)
			if m == nil {
				return Callee{}, false
			}
			if i > 0 && (contract.ByRef != m.ByRef || !slices.Equal(contract.Params, m.Params)) {
				return Callee{}, false
			}
			contract = Callee{Params: m.Params, ByRef: m.ByRef, Builtin: m.Builtin && (i == 0 || contract.Builtin)}
		}
		return contract, true
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return Callee{}, false
		}
		if _, named := c.Class.(*syntax.Name); !named {
			return Callee{}, false
		}
		cls := env.ClassRef(c.Class)
		if cls == "" {
			return Callee{}, false
		}
		if m := env.Index.FindMethod(cls, id.Value, env.PHP); m != nil {
			return Callee{Params: m.Params, ByRef: m.ByRef, Builtin: m.Builtin}, true
		}
	case *syntax.New:
		name, ok := c.Class.(*syntax.Name)
		if !ok || strings.EqualFold(name.Value, "static") {
			return Callee{}, false
		}
		cls := env.ClassRef(name)
		if declared := env.Index.Class(cls, env.PHP); declared == nil || !declared.Avail.In(env.PHP) {
			return Callee{}, false
		}
		m, known := constructorMethod(env, cls)
		if !known {
			return Callee{}, false
		}
		if m != nil {
			return Callee{Params: m.Params, ByRef: m.ByRef, Builtin: m.Builtin}, true
		}
		return Callee{}, true
	}
	return Callee{}, false
}

// constructorMethod follows constructor inheritance without treating an ordinary
// inherited method named after the child as a legacy constructor. Missing or
// ambiguous parents and legacy trait composition cannot prove absence.
func constructorMethod(env *infer.Env, class string) (*index.Method, bool) {
	modern := env.Index.FindMethod(class, "__construct", env.PHP)
	for steps := 0; steps < index.MaxAncestors; steps++ {
		declared := env.Index.Class(class, env.PHP)
		if declared == nil || !declared.Avail.In(env.PHP) || env.Index.ClassCount(class) != 1 {
			return nil, false
		}
		if modern != nil && (strings.EqualFold(modern.Class, class) || strings.EqualFold(modern.TypeClass, class)) {
			return modern, env.Index.ClassCount(modern.Class) == 1
		}
		if env.PHP < phpversion.PHP80 {
			short := class[strings.LastIndexByte(class, '\\')+1:]
			if legacy := declared.Methods[strings.ToLower(short)]; legacy != nil && !legacy.Magic && legacy.Avail.In(env.PHP) {
				if !strings.Contains(class, `\`) {
					return legacy, true
				}
				if env.PHP == phpversion.PHP53 {
					return nil, false // namespaced legacy constructors changed during 5.3
				}
			}
		}
		if len(declared.Traits) > 0 {
			return nil, false
		}
		if declared.Parent == "" {
			return nil, true
		}
		class = declared.Parent
	}
	return nil, false
}

// StaticCallClass resolves the class part of a static call or class
// constant fetch (self/static/parent through the enclosing class); "" when
// dynamic or unresolvable.
func StaticCallClass(ctx *analysis.Context, e syntax.Expr) string {
	nm, ok := e.(*syntax.Name)
	if !ok {
		return ""
	}
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		return ctx.Types().ClassFQN(syntax.EnclosingClass(nm))
	case "parent":
		if c := ctx.Index().Class(ctx.Types().ClassFQN(syntax.EnclosingClass(nm)), ctx.PHP); c != nil {
			return strings.TrimPrefix(c.Parent, `\`)
		}
		return ""
	}
	return ctx.Names().Class(nm.Value, nm.Span().Start)
}
