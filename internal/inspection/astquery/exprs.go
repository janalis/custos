package astquery

import (
	"strings"

	"custos/internal/php/syntax"
)

// FuncNamePart returns the name node of a call to a named function and its
// last segment as written (`\Foo\bar` -> `bar`); ok is false for dynamic
// calls such as `$f()`.
func FuncNamePart(call *syntax.FuncCall) (name *syntax.Name, part string, ok bool) {
	name, ok = call.Name.(*syntax.Name)
	if !ok {
		return nil, "", false
	}
	return name, LastNamePart(name.Value), true
}

// IsFuncNamed reports whether e is a plain function call whose name part
// equals part exactly (case-sensitive, any namespace qualifier accepted).
func IsFuncNamed(e syntax.Node, part string) (*syntax.FuncCall, bool) {
	return funcNamed(e, part, false)
}

// IsFuncNamedFold is IsFuncNamed with a case-insensitive comparison of the
// name part, as PHP compares function names. part must be lower-case.
func IsFuncNamedFold(e syntax.Node, part string) (*syntax.FuncCall, bool) {
	return funcNamed(e, part, true)
}

func funcNamed(e syntax.Node, part string, fold bool) (*syntax.FuncCall, bool) {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return nil, false
	}
	_, p, ok := FuncNamePart(call)
	if !ok || (fold && !strings.EqualFold(p, part)) || (!fold && p != part) {
		return nil, false
	}
	return call, true
}

// NamePartSpan returns the span of the last segment of a name (the part after
// the last backslash).
func NamePartSpan(name *syntax.Name) syntax.Span {
	s := name.Span()
	i := strings.LastIndexByte(name.Value, '\\')
	return syntax.Span{Start: s.Start + uint32(i+1), End: s.End}
}

// ArgCount returns the number of entries in a call's argument list, spreads
// and placeholders included (0 without a list).
func ArgCount(call *syntax.FuncCall) int {
	if call.Args == nil {
		return 0
	}
	return len(call.Args.Args)
}

// AsExpr returns n as an expression, nil when it is not one.
func AsExpr(n syntax.Node) syntax.Expr {
	e, _ := n.(syntax.Expr)
	return e
}

// IsStaticPropName reports whether v is the member name of a static
// property access (`X::$p`), which the AST models as a variable but is not
// one.
func IsStaticPropName(v *syntax.Variable) bool {
	sp, ok := v.Parent().(*syntax.StaticPropertyFetch)
	return ok && sp.Name == syntax.Expr(v)
}

// MentionsVariable reports whether a simple variable `$name` occurs
// anywhere under n (nested function-likes included); false for nil.
func MentionsVariable(n syntax.Node, name string) bool {
	if n == nil {
		return false
	}
	found := false
	syntax.Inspect(n, func(x syntax.Node) bool {
		if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
			found = true
		}
		return !found
	})
	return found
}

// IsLogicalOperand reports whether n (through parentheses) is used as a
// condition: of if/elseif/while/do-while, of a full ternary, or an operand
// of `!`, `&&`, `||`, `and`, `or`.
func IsLogicalOperand(n syntax.Node) bool {
	parent, child := ParentSkipParens(n)
	switch p := parent.(type) {
	case *syntax.If:
		return p.Cond == child
	case *syntax.ElseIf:
		return p.Cond == child
	case *syntax.While:
		return p.Cond == child
	case *syntax.DoWhile:
		return p.Cond == child
	case *syntax.Unary:
		return p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	case *syntax.Ternary:
		return p.Then != nil && p.Cond == child
	}
	return false
}
