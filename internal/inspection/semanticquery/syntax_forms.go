package semanticquery

import (
	"strings"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// DirectCondition identifies a direct boolean condition without assuming branch intent.
func DirectCondition(e syntax.Expr) bool {
	p, child := astquery.ParentSkipParens(e)
	if a, ok := p.(*syntax.Assign); ok && a.Value == child && a.Op.Kind == syntax.TEqual && !a.ByRef {
		p, child = astquery.ParentSkipParens(a)
	}
	if u, ok := p.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		p, child = astquery.ParentSkipParens(u)
	}
	switch c := p.(type) {
	case *syntax.If:
		return c.Cond == child
	case *syntax.ElseIf:
		return c.Cond == child
	case *syntax.While:
		return c.Cond == child
	case *syntax.DoWhile:
		return c.Cond == child
	}
	return false
}

// IsFalse reports whether e is the constant false.
func IsFalse(e syntax.Expr) bool {
	v, k := astquery.BoolConst(syntax.UnwrapParens(e))
	return k && !v
}

// LiteralMethodName returns a literal method name, excluding dynamic names.
func LiteralMethodName(c *syntax.MethodCall) string {
	if id, ok := c.Name.(*syntax.Identifier); ok {
		return strings.ToLower(id.Value)
	}
	return ""
}

// IsLiteral reports whether e is a literal or true, false or null.
func IsLiteral(e syntax.Expr) bool {
	switch value := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return value.Name != nil && (strings.EqualFold(value.Name.Value, "true") || strings.EqualFold(value.Name.Value, "false") || strings.EqualFold(value.Name.Value, "null"))
	}
	return false
}
