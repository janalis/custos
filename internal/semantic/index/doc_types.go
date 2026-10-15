package index

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// ResolveDocType resolves constant PHPDoc types without coupling type parsing
// to symbol ownership. lookup must return the class visible to the caller.
func ResolveDocType(expr string, resolve types.Resolver, lookup func(string) *Class) types.Type {
	values := false
	keys := false
	if strings.HasPrefix(expr, "key-of<") && strings.HasSuffix(expr, ">") {
		keys = true
		expr = expr[len("key-of<") : len(expr)-1]
	}
	if strings.HasPrefix(expr, "value-of<") && strings.HasSuffix(expr, ">") {
		values = true
		expr = expr[len("value-of<") : len(expr)-1]
	}
	owner, member, constant := strings.Cut(expr, "::")
	if !constant && !values && !keys {
		return types.Unknown
	}
	if !constant {
		owner = expr
	}
	if resolve != nil {
		owner = resolve(owner)
	}
	c := lookup(strings.TrimPrefix(owner, `\`))
	if c == nil {
		return types.Unknown
	}
	if keys {
		if !constant && c.Kind == syntax.KindEnum {
			return types.String
		}
		return types.Unknown
	}
	result := types.Unknown
	for name, cc := range c.Consts {
		if values && !constant && !cc.Case {
			continue
		}
		if constant && name != member && !(strings.HasSuffix(member, "*") && strings.HasPrefix(name, strings.TrimSuffix(member, "*"))) {
			continue
		}
		t := types.Mixed
		if cc.Case {
			if values {
				if backing := c.Props["value"]; backing != nil && (backing.Type == "int" || backing.Type == "string") {
					t = types.FromDoc(backing.Type, nil)
				}
			} else {
				t = types.Of(`\` + c.FQN)
			}
		} else if !values && cc.Type != "" {
			t = types.FromDoc(cc.Type, nil)
		}
		if result.IsUnknown() {
			result = t
		} else {
			result = types.Union(result, t)
		}
	}
	return result
}
