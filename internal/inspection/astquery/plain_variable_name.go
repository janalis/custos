package astquery

import "custos/internal/php/syntax"

// PlainVariableName returns the name of a simple `$name` variable.
func PlainVariableName(e syntax.Node) (string, bool) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" {
		return "", false
	}
	return v.Name, true
}
