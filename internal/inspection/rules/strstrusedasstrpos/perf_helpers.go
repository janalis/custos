package strstrusedasstrpos

import (
	"custos/internal/php/syntax"
)

// perfIsFalse reports whether e is the unqualified constant false (any case).
func perfIsFalse(e syntax.Node) bool {
	c, ok := e.(*syntax.ConstFetch)
	if !ok || c.Name == nil || c.Name.NameKind != syntax.NameUnqualified {
		return false
	}
	v := c.Name.Value
	return len(v) == 5 && (v[0]|0x20) == 'f' && (v[1]|0x20) == 'a' && (v[2]|0x20) == 'l' && (v[3]|0x20) == 's' && (v[4]|0x20) == 'e'
}
