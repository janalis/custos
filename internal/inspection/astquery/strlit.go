package astquery

import (
	"custos/internal/php/syntax"
)

// StringLiteralValue decodes the raw source of a single- or double-quoted
// string literal without interpolation (quotes included) according to PHP's
// escape rules for its quote style. ok is false for other literals
// (heredoc/nowdoc, numbers).
func StringLiteralValue(raw string) (val string, ok bool) {
	return syntax.StringLiteralValue(raw)
}

// QuotedStringValue decodes a single- or double-quoted string literal
// without interpolation (see StringLiteralValue); ok is false for any other
// node.
func QuotedStringValue(e syntax.Expr) (string, bool) {
	l, ok := e.(*syntax.Literal)
	if !ok || l.LitKind != syntax.LitString {
		return "", false
	}
	return syntax.StringLiteralValueForVersion(l.Raw, syntax.SourceVersion(l))
}

// IsStringLiteral reports whether e is a string literal: quoted, heredoc or
// nowdoc, interpolated or not (backtick shell commands excluded).
func IsStringLiteral(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Literal:
		return x.LitKind == syntax.LitString
	case *syntax.InterpolatedString:
		return !x.Backtick
	}
	return false
}

// IsNumberLiteral reports whether e is an int or float literal, or `-`
// applied directly to one (no parentheses looked through).
func IsNumberLiteral(e syntax.Expr) bool {
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		e = u.Expr
	}
	l, ok := e.(*syntax.Literal)
	return ok && (l.LitKind == syntax.LitInt || l.LitKind == syntax.LitFloat)
}
