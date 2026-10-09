package infer

import (
	"strings"

	"custos/internal/php/syntax"
)

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
// Deliberately shallower than syntax.Terminates (no if/else, try or switch
// analysis): switching would make more guards count as early exits, which
// changes narrowing and therefore findings.
func terminates(s syntax.Stmt) bool {
	switch n := s.(type) {
	case *syntax.Return, *syntax.Break, *syntax.Continue, *syntax.Goto:
		return true
	case *syntax.ExprStmt:
		if _, ok := syntax.UnwrapParens(n.Expr).(*syntax.Throw); ok {
			return true
		}
		return syntax.ExitInvocation(n.Expr)
	case *syntax.Block:
		if len(n.Stmts) > 0 {
			return terminates(n.Stmts[len(n.Stmts)-1])
		}
	}
	return false
}

func isVar(x syntax.Expr, name string) bool {
	x = syntax.UnwrapParens(x)
	if a, ok := x.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && !a.ByRef {
		x = a.Var // `false === ($x = f())` tests $x
	}
	return narrowKey(x) == name
}

// constLiteral returns "null", "true" or "false" for those constants.
func constLiteral(x syntax.Expr) string {
	c, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch)
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

// maxCondSteps caps the condition nodes one applyCond call examines. The
// negation of `A && B` (and the truth of `A || B`) narrows A both ways, so
// alternating nestings would otherwise cost 3^depth; beyond the budget the
// remaining sub-conditions narrow nothing (sound: the type is kept).
const maxCondSteps = 1024
