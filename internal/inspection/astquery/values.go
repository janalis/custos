package astquery

import (
	"strings"

	"custos/internal/php/syntax"
)

// AssignedValue follows `$a = $b = v` chains (plain `=` only) to v.
func AssignedValue(a *syntax.Assign) syntax.Expr {
	v := a.Value
	for {
		inner, ok := syntax.UnwrapParens(v).(*syntax.Assign)
		if !ok || inner.Op.Kind != syntax.TEqual {
			return v
		}
		v = inner.Value
	}
}

// ScopeParts returns the parameters and the body (block, or expression for
// an arrow function; nil for an abstract method) of a function-like.
func ScopeParts(scope syntax.Node) (params []*syntax.Param, body syntax.Node) {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.Params, s.Body
	case *syntax.Method:
		if s.Body != nil {
			return s.Params, s.Body
		}
		return s.Params, nil
	case *syntax.Closure:
		return s.Params, s.Body
	case *syntax.ArrowFunction:
		return s.Params, s.Expr
	}
	return nil, nil
}

// FileClassNamed returns the first class-like declared in f (preorder)
// whose short name is want (case-insensitive), or nil.
func FileClassNamed(f *syntax.File, want string) *syntax.ClassLike {
	var found *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if found != nil {
			return false
		}
		if cl, ok := n.(*syntax.ClassLike); ok && cl.Name != nil && strings.EqualFold(cl.Name.Value, want) {
			found = cl
		}
		return true
	})
	return found
}

// ClassConstValues returns the value expressions of the constants named
// name declared directly in class, in declaration order.
func ClassConstValues(class *syntax.ClassLike, name string) []syntax.Expr {
	var out []syntax.Expr
	for _, m := range class.Members {
		if cc, ok := m.(*syntax.ClassConst); ok {
			for _, it := range cc.Consts {
				if it.Name != nil && it.Name.Value == name && it.Value != nil {
					out = append(out, it.Value)
				}
			}
		}
	}
	return out
}

// FileConstValue returns the value expression of the global constant name
// declared in f (`const name = v;` or `define('name', v)`), or nil.
func FileConstValue(f *syntax.File, name string) syntax.Expr {
	var found syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if found != nil {
			return false
		}
		switch x := n.(type) {
		case *syntax.ConstStmt:
			for _, it := range x.Consts {
				if it.Name != nil && it.Name.Value == name && it.Value != nil {
					found = it.Value
				}
			}
		case *syntax.FuncCall:
			if CallLastName(x) != "define" {
				return true
			}
			args, ok := CallArgValues(x)
			if !ok || len(args) < 2 {
				return true
			}
			if lit, ok := args[0].(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
				if v, ok := StringLiteralValue(lit.Raw); ok && strings.TrimPrefix(v, `\`) == name {
					found = args[1]
				}
			}
		}
		return true
	})
	return found
}
