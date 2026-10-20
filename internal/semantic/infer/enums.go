package infer

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// enumCase returns the enum class and the case an `Enum::NAME` expression
// denotes, following constants of the enum that alias one of its cases
// (`const Fallback = self::Left;`); ok is false for anything else or a
// constant that cannot be traced to a case.
func (e *Env) enumCase(x syntax.Expr) (c *index.Class, name string, ok bool) {
	fetch, isFetch := syntax.UnwrapParens(x).(*syntax.ClassConstFetch)
	if !isFetch {
		return nil, "", false
	}
	id, isID := fetch.Name.(*syntax.Identifier)
	if !isID {
		return nil, "", false
	}
	c = e.Index.Class(e.classRef(fetch.Class), e.PHP)
	if c == nil || c.Kind != syntax.KindEnum {
		return nil, "", false
	}
	short := c.FQN[strings.LastIndexByte(c.FQN, '\\')+1:]
	name = id.Value
	// Bounded: a constant chain is short, and a cyclic one never ends.
	for range 8 {
		k := c.Consts[name]
		switch {
		case k == nil:
			return nil, "", false
		case k.Case:
			return c, name, true
		}
		owner, target, found := strings.Cut(k.Value, "::")
		owner = strings.TrimPrefix(strings.TrimSpace(owner), `\`)
		if !found || !strings.EqualFold(owner, "self") && !strings.EqualFold(owner, short) && !strings.EqualFold(owner, c.FQN) {
			return nil, "", false
		}
		name = strings.TrimSpace(target)
	}
	return nil, "", false
}

// enumCaseNames returns the cases of enum c, sorted.
func enumCaseNames(c *index.Class) []string {
	var names []string
	for n, k := range c.Consts {
		if k.Case {
			names = append(names, n)
		}
	}
	slices.Sort(names)
	return names
}
