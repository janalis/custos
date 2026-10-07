package infer

import (
	"strings"

	"custos/internal/syntax"
	"custos/internal/types"
)

// narrow refines the type of variable occurrence v using the conditions that
// guard it: ternary branches, if/elseif/else bodies, the right operand of
// && / ||, and early-exit guards (`if (cond) { return; }`) earlier in the
// same block. Only checks on the same variable name are understood.
func (e *Env) narrow(t types.Type, v *syntax.Variable, scope syntax.Node) types.Type {
	return e.narrowExpr(t, v, v.Name, scope)
}

// narrowKey identifies an expression whose type guards can refine: a
// variable by its name, or `$this->prop` as "this->prop" ("" otherwise).
func narrowKey(x syntax.Expr) string {
	switch n := x.(type) {
	case *syntax.Variable:
		if n.NameExpr == nil {
			return n.Name
		}
	case *syntax.PropertyFetch:
		if v, ok := n.Var.(*syntax.Variable); ok && v.Name == "this" && v.NameExpr == nil {
			if id, ok := n.Name.(*syntax.Identifier); ok {
				return "this->" + id.Value
			}
		}
	}
	return ""
}

// narrowExpr refines the type t of occurrence x (identified by key, see
// narrowKey) by the conditions guarding it.
func (e *Env) narrowExpr(t types.Type, x syntax.Expr, key string, scope syntax.Node) types.Type {
	if t.IsUnknown() || key == "" {
		return t
	}
	var child syntax.Node = x
	for p := x.Parent(); p != nil && p != scope; child, p = p, p.Parent() {
		switch n := p.(type) {
		case *syntax.Ternary:
			if n.Then != nil && child == syntax.Node(n.Then) {
				t = e.applyCond(t, n.Cond, key, true)
			} else if child == syntax.Node(n.Else) && n.Then != nil {
				t = e.applyCond(t, n.Cond, key, false)
			}
		case *syntax.Binary:
			if child == syntax.Node(n.Right) {
				switch n.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					t = e.applyCond(t, n.Left, key, true)
				case syntax.TBooleanOr, syntax.TOr:
					t = e.applyCond(t, n.Left, key, false)
				}
			}
		case *syntax.If:
			if child == syntax.Node(n.Body) {
				t = e.applyCond(t, n.Cond, key, true)
			} else if n.Else != nil && child == syntax.Node(n.Else) && len(n.ElseIfs) == 0 {
				t = e.applyCond(t, n.Cond, key, false)
			}
		case *syntax.While:
			if child == syntax.Node(n.Body) {
				t = e.applyCond(t, n.Cond, key, true)
			}
		case *syntax.ElseIf:
			if child == syntax.Node(n.Body) {
				t = e.applyCond(t, n.Cond, key, true)
			}
		case *syntax.Block:
			t = e.guards(t, n.Stmts, child, key)
		case *syntax.Case:
			t = e.guards(t, n.Stmts, child, key)
		case *syntax.Namespace:
			t = e.guards(t, n.Stmts, child, key)
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
			return t
		}
	}
	if scope == nil {
		t = e.guards(t, e.File.Stmts, child, key)
	}
	return t
}

// guards applies early-exit guards among the statements preceding child.
func (e *Env) guards(t types.Type, stmts []syntax.Stmt, child syntax.Node, name string) types.Type {
	orig := t
	for _, s := range stmts {
		if syntax.Node(s) == child {
			break
		}
		if assigns(s, name) {
			// A later write invalidates earlier guards.
			t = orig
			continue
		}
		if g, ok := s.(*syntax.If); ok && g.Else == nil && len(g.ElseIfs) == 0 {
			if terminates(g.Body) {
				t = e.applyCond(t, g.Cond, name, false)
			} else if val := e.overwrites(g.Body, name); val != nil {
				// `if (false === $x) { $x = $default; }`: past the if, the
				// condition no longer holds (unless the new value matches it).
				vt := e.TypeOf(val)
				if !vt.IsUnknown() && e.applyCond(vt, g.Cond, name, false).String() == vt.String() {
					t = e.applyCond(t, g.Cond, name, false)
				}
			}
		}
	}
	return t
}

// assigns reports whether statement s writes variable name at its top level.
func assigns(s syntax.Stmt, name string) bool {
	es, ok := s.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	a, ok := es.Expr.(*syntax.Assign)
	if !ok {
		return false
	}
	return narrowKey(a.Var) == name
}

// overwrites returns the value of a plain `$name = value;` that is the last
// statement of body, or nil.
func (e *Env) overwrites(body syntax.Stmt, name string) syntax.Expr {
	last := body
	if b, ok := body.(*syntax.Block); ok {
		if len(b.Stmts) == 0 {
			return nil
		}
		last = b.Stmts[len(b.Stmts)-1]
	}
	if !assigns(last, name) {
		return nil
	}
	a := last.(*syntax.ExprStmt).Expr.(*syntax.Assign)
	if a.Op.Kind != syntax.TEqual || a.ByRef {
		return nil
	}
	return a.Value
}

// terminates reports whether a statement always leaves the current block.
func terminates(s syntax.Stmt) bool {
	switch n := s.(type) {
	case *syntax.Return, *syntax.Break, *syntax.Continue, *syntax.Goto:
		return true
	case *syntax.ExprStmt:
		switch n.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	case *syntax.Block:
		if len(n.Stmts) > 0 {
			return terminates(n.Stmts[len(n.Stmts)-1])
		}
	}
	return false
}

func unparen(x syntax.Expr) syntax.Expr {
	for {
		p, ok := x.(*syntax.Paren)
		if !ok {
			return x
		}
		x = p.Expr
	}
}

func isVar(x syntax.Expr, name string) bool {
	x = unparen(x)
	if a, ok := x.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
		x = a.Var // `false === ($x = f())` tests $x
	}
	return narrowKey(x) == name
}

func isNullConst(x syntax.Expr) bool {
	c, ok := unparen(x).(*syntax.ConstFetch)
	return ok && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "null")
}

// constLiteral returns "null", "true" or "false" for those constants.
func constLiteral(x syntax.Expr) string {
	c, ok := unparen(x).(*syntax.ConstFetch)
	if !ok {
		return ""
	}
	switch v := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)); v {
	case "null", "true", "false":
		return v
	}
	return ""
}

var typeChecks = map[string][]string{
	"is_array": {"array"}, "is_string": {"string"}, "is_int": {"int"}, "is_integer": {"int"},
	"is_long": {"int"}, "is_float": {"float"}, "is_double": {"float"}, "is_bool": {"bool", "true", "false"},
	"is_null": {"null"}, "is_object": {"object"}, "is_callable": {"callable"}, "is_iterable": {"iterable", "array"},
	"is_numeric": {"int", "float", "string"}, "is_scalar": {"int", "float", "string", "bool", "true", "false"},
	"is_resource": {"resource"},
}

// applyCond narrows t assuming cond evaluates to `truthy`.
func (e *Env) applyCond(t types.Type, cond syntax.Expr, name string, truthy bool) types.Type {
	switch c := unparen(cond).(type) {
	case *syntax.Unary:
		if c.Op.Kind == syntax.TExclaim {
			return e.applyCond(t, c.Expr, name, !truthy)
		}
	case *syntax.Binary:
		switch c.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truthy {
				return e.applyCond(e.applyCond(t, c.Left, name, true), c.Right, name, true)
			}
		case syntax.TBooleanOr, syntax.TOr:
			if !truthy {
				return e.applyCond(e.applyCond(t, c.Left, name, false), c.Right, name, false)
			}
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			lit := ""
			switch {
			case isVar(c.Left, name):
				lit = constLiteral(c.Right)
			case isVar(c.Right, name):
				lit = constLiteral(c.Left)
			}
			if lit == "" {
				return t
			}
			if (c.Op.Kind == syntax.TIsIdentical) == truthy {
				if t.Has(lit) || (lit != "null" && t.Has("bool")) {
					return types.Of(lit)
				}
				return t
			}
			out := t.Without(lit)
			if lit != "null" && out.Has("bool") {
				other := map[string]string{"true": "false", "false": "true"}[lit]
				out = types.Union(out.Without("bool"), types.Of(other))
			}
			return out
		case syntax.TIsEqual, syntax.TIsNotEqual:
			isNull := (isVar(c.Left, name) && isNullConst(c.Right)) || (isVar(c.Right, name) && isNullConst(c.Left))
			if !isNull {
				return t
			}
			if (c.Op.Kind == syntax.TIsEqual) != truthy {
				return t.Without("null")
			}
		}
	case *syntax.Instanceof:
		if !isVar(c.Expr, name) {
			return t
		}
		if truthy {
			if cls := e.classRef(c.Class); cls != "" {
				return types.Of(`\` + cls)
			}
			return t
		}
		if cls := e.classRef(c.Class); cls != "" {
			return t.Without(`\` + cls)
		}
	case *syntax.Isset:
		for _, x := range c.Vars {
			if isVar(x, name) && truthy {
				return t.Without("null")
			}
		}
	case *syntax.Variable, *syntax.PropertyFetch:
		if narrowKey(c) == name && truthy {
			return t.Without("null", "false")
		}
	case *syntax.FuncCall:
		nm, ok := c.Name.(*syntax.Name)
		if !ok || c.Args == nil || len(c.Args.Args) == 0 {
			return t
		}
		a, ok := c.Args.Args[0].(*syntax.Arg)
		if !ok || !isVar(a.Value, name) {
			return t
		}
		fn := strings.ToLower(strings.TrimPrefix(nm.Value, `\`))
		atoms, ok := typeChecks[fn]
		if !ok {
			return t
		}
		return narrowAtoms(t, atoms, truthy)
	}
	return t
}

// narrowAtoms keeps (truthy) or removes (falsy) the atoms matching a type check.
func narrowAtoms(t types.Type, atoms []string, truthy bool) types.Type {
	match := func(a string) bool {
		for _, x := range atoms {
			if a == x || (x == "array" && strings.HasSuffix(a, "[]")) || (x == "object" && strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]")) ||
				(x == "bool" && (a == "true" || a == "false")) {
				return true
			}
		}
		return false
	}
	var keep []string
	for _, a := range t.Atoms() {
		if match(a) == truthy {
			keep = append(keep, a)
		}
	}
	if len(keep) == 0 {
		if truthy {
			return types.Of(atoms[0])
		}
		return t
	}
	return types.Of(keep...)
}
