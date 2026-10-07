package util

import (
	"strings"

	"custos/internal/syntax"
)

// CtorAssignsProperty reports whether class c's own constructor assigns
// `$this->name` a value other than the null literal in a top-level
// statement of its body (so the property no longer holds its implicit null
// once the object is constructed).
func CtorAssignsProperty(c *syntax.ClassLike, name string) bool {
	if c == nil {
		return false
	}
	for _, mem := range c.Members {
		ctor, ok := mem.(*syntax.Method)
		if !ok || ctor.Name == nil || ctor.Body == nil || !strings.EqualFold(ctor.Name.Value, "__construct") {
			continue
		}
		for _, st := range ctor.Body.Stmts {
			es, ok := st.(*syntax.ExprStmt)
			if !ok {
				continue
			}
			a, ok := es.Expr.(*syntax.Assign)
			if !ok || a.Op.Kind != syntax.TEqual {
				continue
			}
			pf, ok := a.Var.(*syntax.PropertyFetch)
			if !ok {
				continue
			}
			v, ok1 := pf.Var.(*syntax.Variable)
			id, ok2 := pf.Name.(*syntax.Identifier)
			if !ok1 || !ok2 || v.NameExpr != nil || v.Name != "this" || id.Value != name {
				continue
			}
			if c, ok := UnwrapParens(a.Value).(*syntax.ConstFetch); ok && c.Name != nil &&
				strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "null") {
				continue
			}
			return true
		}
	}
	return false
}
