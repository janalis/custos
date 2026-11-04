package flowquery

import (
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// SingleStringLiteral returns e when it is a quoted string literal without
// interpolation; otherwise it runs PossibleValues on e and returns the only
// quoted, non-interpolated string literal among the results (other values
// are ignored). nil when there is none or more than one, and when the
// discovery result is unknown (an unstable variable on the path).
func SingleStringLiteral(f *syntax.File, e syntax.Expr) syntax.Expr {
	return StringLiteralWithReferences(f, e, nil)
}

// StringLiteralWithReferences is SingleStringLiteral with resolved argument
// effects. Unknown bindings must be treated as possible writes.
func StringLiteralWithReferences(f *syntax.File, e syntax.Expr, mayWrite func(*syntax.Arg) bool) syntax.Expr {
	if _, _, ok := astquery.QuotedStringRaw(e); ok {
		return e
	}
	if _, ok := e.(*syntax.Literal); ok {
		return nil
	}
	var found syntax.Expr
	values, _ := PossibleValuesWithReferences(f, e, mayWrite)
	for _, v := range values {
		if _, _, ok := astquery.QuotedStringRaw(v); ok {
			if found != nil {
				return nil
			}
			found = v
		}
	}
	return found
}

// ConstantStringWithReferences finds one noninterpolated constant string,
// including heredoc and nowdoc, among lexical candidates. Other values are
// ignored, as in SingleStringLiteral; unknown discovery proves nothing.
func ConstantStringWithReferences(f *syntax.File, e syntax.Expr, mayWrite func(*syntax.Arg) bool) syntax.Expr {
	var found syntax.Expr
	values, known := PossibleValuesWithReferences(f, e, mayWrite)
	if !known {
		return nil
	}
	for _, value := range values {
		literal, ok := value.(*syntax.Literal)
		if !ok {
			continue
		}
		if _, ok := syntax.LiteralStringValue(literal); ok {
			if found != nil {
				return nil
			}
			found = value
		}
	}
	return found
}
