package index

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
		if constant && !constantMatches(name, member) {
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

// constantMatches reports whether constant name is the one member names, or
// one of those a wildcard member (`BAR_*`, `*`) names.
func constantMatches(name, member string) bool {
	if prefix, wildcard := strings.CutSuffix(member, "*"); wildcard {
		return strings.HasPrefix(name, prefix)
	}
	return name == member
}

// ConstantValues is the type of the values of the constants doc text expr
// names (`\Foo::BAR`, `\Foo::BAR_*`, `\Foo::*`, or `\::NAME` for a global
// constant written unqualified), inherited class constants included: the
// exact value of each, unknown unless every one is declared with an int or
// plain string literal.
func ConstantValues(ix *Index, expr string, ver phpversion.Version) types.Type {
	owner, member, _ := strings.Cut(strings.TrimPrefix(expr, `\`), "::")
	if owner == "" {
		k := ix.Constant(member, ver)
		if k == nil {
			k = ix.Constant(member[strings.LastIndexByte(member, '\\')+1:], ver)
		}
		if k == nil {
			return types.Unknown
		}
		return LiteralType(k.Value)
	}
	if !strings.Contains(member, "*") {
		if k := ix.FindConst(owner, member, ver); k != nil {
			return LiteralType(k.Value)
		}
		return types.Unknown
	}
	if !ix.AncestorsComplete(owner, ver) {
		return types.Unknown
	}
	var values []types.Type
	seen := map[string]bool{}
	for _, c := range ix.Ancestors(owner, ver) {
		for name, k := range c.Consts {
			if seen[name] || !constantMatches(name, member) {
				continue
			}
			seen[name] = true
			v := LiteralType(k.Value)
			if v.IsUnknown() {
				return types.Unknown
			}
			values = append(values, v)
		}
	}
	if len(values) == 0 {
		return types.Unknown
	}
	return types.Union(values...)
}

// LiteralType is the exact type of int or plain string literal source text;
// unknown for anything else.
func LiteralType(text string) types.Type {
	text = strings.TrimSpace(text)
	if v, err := strconv.ParseInt(text, 0, 64); err == nil {
		return types.IntValue(v)
	}
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') && text[len(text)-1] == text[0] && !strings.ContainsAny(text[1:len(text)-1], `\$'"`) {
		return types.StringLiteral(text[1 : len(text)-1])
	}
	return types.Unknown
}
