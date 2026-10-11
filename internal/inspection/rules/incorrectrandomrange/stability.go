package incorrectrandomrange

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// stability proves that value discovery saw every value a bound can hold
// (D7): shared discovery only reads plain assignments, so a variable or
// property that is bound any other way would keep a stale literal.
type stability struct {
	ctx  *analysis.Context
	seen map[any]bool
}

type (
	varKey struct {
		scope syntax.Node
		name  string
	}
	propKey struct {
		class *syntax.ClassLike
		name  string
	}
	dynamicScopeKey struct{ body syntax.Node }
)

// expr follows the expressions discovery expands (ternary, `??`, variables,
// properties, class constants); anything else is a value itself.
func (s *stability) expr(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Ternary:
		return s.expr(x.Cond) && s.expr(x.Then) && s.expr(x.Else)
	case *syntax.Binary:
		return x.Op.Kind != syntax.TCoalesce || s.expr(x.Left) && s.expr(x.Right)
	case *syntax.Variable:
		return s.variable(x)
	case *syntax.PropertyFetch:
		id, named := x.Name.(*syntax.Identifier)
		v, ok := x.Var.(*syntax.Variable)
		return named && ok && v.Name == "this" && s.property(x, id.Value)
	case *syntax.StaticPropertyFetch:
		v, named := x.Name.(*syntax.Variable)
		cn, ok := x.Class.(*syntax.Name)
		return named && ok && v.Name != "" && s.property(x, v.Name) &&
			(strings.EqualFold(cn.Value, "self") || strings.EqualFold(cn.Value, "static"))
	case *syntax.ClassConstFetch:
		return lexicallyBound(x)
	}
	return true
}

// variable accepts a local variable whose every occurrence in its scope is
// a read or a plain assignment of a stable value.
func (s *stability) variable(v *syntax.Variable) bool {
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil || v.Name == "" {
		return false
	}
	key := varKey{scope, v.Name}
	if s.seen[key] {
		return true
	}
	s.seen[key] = true
	params, body := astquery.ScopeParts(scope)
	for _, p := range params {
		// The caller supplies the value of a parameter without default.
		if p.Var != nil && p.Var.Name == v.Name && (p.ByRef || p.Default == nil) {
			return false
		}
	}
	if s.dynamicScope(body) {
		return false
	}
	for _, occ := range flowquery.VarOccurrences(s.ctx.File, body)[v.Name] {
		if !s.use(occ) {
			return false
		}
	}
	return true
}

// dynamicScope reports constructs that can bind any variable of the scope:
// variable-variables, extract(), include/require and eval.
func (s *stability) dynamicScope(body syntax.Node) bool {
	return s.ctx.File.Memo(dynamicScopeKey{body}, func() any {
		found := false
		syntax.Inspect(body, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Variable:
				found = x.NameExpr != nil
			case *syntax.Include, *syntax.Eval:
				found = true
			case *syntax.FuncCall:
				found = s.ctx.GlobalFunctionName(x) == "extract"
			}
			return !found
		})
		return found
	}).(bool)
}

// property accepts `$this->p` / `self::$p` for a private, hook-free property
// of the enclosing class that the class only reads or plainly assigns where
// discovery looks (the current scope and the constructor).
func (s *stability) property(fetch syntax.Expr, name string) bool {
	class := syntax.EnclosingClass(fetch)
	if class == nil || !declaresPrivate(class, name) {
		return false
	}
	key := propKey{class, name}
	if s.seen[key] {
		return true
	}
	s.seen[key] = true
	scope := syntax.EnclosingFuncLike(fetch)
	stable := true
	check := func(x syntax.Expr) {
		if a, ok := x.Parent().(*syntax.Assign); ok && a.Var == x {
			w, _ := syntax.EnclosingFuncLike(a).(*syntax.Method)
			ctor := w != nil && w.Name != nil && w.Parent() == syntax.Node(class) &&
				strings.EqualFold(w.Name.Value, "__construct")
			if !astquery.Equivalent(s.ctx.File, x, fetch) || syntax.EnclosingFuncLike(a) != scope && !ctor {
				stable = false
				return
			}
		}
		stable = s.use(x)
	}
	syntax.Inspect(class, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.TraitUse: // trait methods may write private properties
			stable = false
		case *syntax.PropertyFetch:
			if id, ok := x.Name.(*syntax.Identifier); !ok || id.Value == name {
				check(x)
			}
		case *syntax.StaticPropertyFetch:
			if v, ok := x.Name.(*syntax.Variable); !ok || v.Name == name || v.Name == "" {
				check(x)
			}
		}
		return stable
	})
	return stable
}

func declaresPrivate(class *syntax.ClassLike, name string) bool {
	for _, m := range class.Members {
		prop, ok := m.(*syntax.Property)
		if !ok || !prop.Modifiers.Has(syntax.TPrivate) || len(prop.Hooks) != 0 {
			continue
		}
		for _, it := range prop.Props {
			if it.Var != nil && it.Var.Name == name {
				return true
			}
		}
	}
	return false
}

// use classifies one occurrence of a variable or property: a read, or a
// plain assignment of a stable value; every other binding form (reference,
// compound, ++/--, destructuring, foreach target, by-reference argument,
// global/static/catch/unset, element write) is unstable.
func (s *stability) use(e syntax.Expr) bool {
	switch p := e.Parent().(type) {
	case *syntax.Assign:
		if p.ByRef {
			return false
		}
		return p.Var != e || p.Op.Kind == syntax.TEqual && s.expr(p.Value)
	case *syntax.Arg:
		return s.byValueArg(astquery.ParentFuncCall(e), p)
	case *syntax.ArrayItem:
		return !p.ByRef && !destructures(p)
	case *syntax.ClosureUse:
		return !p.ByRef
	case *syntax.Foreach:
		return p.Expr == e
	case *syntax.ArrayDimFetch:
		return p.Var != e
	case *syntax.IncDec, *syntax.Unset, *syntax.Global, *syntax.StaticVar, *syntax.Catch,
		*syntax.Param, *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		return false
	}
	return true
}

// byValueArg requires a call to an indexed global function none of whose
// arguments binds by reference; methods and unknown callees may take one.
func (s *stability) byValueArg(call *syntax.FuncCall, arg *syntax.Arg) bool {
	if call == nil || arg.ByRef {
		return false
	}
	fn := s.ctx.Index().Function(s.ctx.GlobalFunctionName(call), s.ctx.PHP)
	return fn != nil && !semanticquery.ArgBindsByRef(call.Args, fn.Params, true)
}

// destructures reports whether the array/list holding it is a destructuring
// target (assignment left side or foreach key/value), at any nesting depth.
func destructures(it *syntax.ArrayItem) bool {
	for {
		list := it.Parent()
		switch p := list.Parent().(type) {
		case *syntax.Assign:
			return syntax.Node(p.Var) == list
		case *syntax.Foreach:
			return syntax.Node(p.Expr) != list
		case *syntax.ArrayItem:
			it = p
		default:
			return false
		}
	}
}

// lexicallyBound rejects `static::C` (and `$obj::C`) unless no subclass can
// redeclare the constant: final or anonymous class, or final constant.
func lexicallyBound(c *syntax.ClassConstFetch) bool {
	cn, ok := c.Class.(*syntax.Name)
	if !ok {
		return false
	}
	if !strings.EqualFold(cn.Value, "static") {
		return true
	}
	class := syntax.EnclosingClass(c)
	id, ok := c.Name.(*syntax.Identifier)
	if class == nil || !ok {
		return false
	}
	if class.Name == nil || class.Modifiers.Has(syntax.TFinal) {
		return true
	}
	for _, m := range class.Members {
		if k, ok := m.(*syntax.ClassConst); ok && k.Modifiers.Has(syntax.TFinal) {
			for _, it := range k.Consts {
				if it.Name.Value == id.Value {
					return true
				}
			}
		}
	}
	return false
}
