package types

import (
	"strings"
)

// Cond is a conditional return type (PHPStan/Psalm):
// `($param is [not] Target ? Then : Else)`, branches possibly nested. The
// index stores it in a canonical form (String) whose names are resolved;
// infer picks a branch when the argument bound to the parameter decides
// the test.
type Cond struct {
	Param   string // the tested parameter, without `$`
	Negated bool   // `is not`
	// Target is what the argument is tested against: a literal (`'a'`,
	// `1`, `-1`, `1.5`), a class constant (`\Foo::BAR`), or a doc type
	// (DocString form).
	Target string
	// Exact reports a type target made of builtin type names and class
	// names only, whose atoms describe it fully (`string`, `?Foo`), so that
	// a value whose atoms all belong to it passes the test; refined types
	// (`non-empty-string`, `int<0, max>`, `array<int>`) only decide the
	// negative outcome.
	Exact      bool
	Then, Else CondBranch
}

// CondBranch is one outcome: a nested conditional, or a type (DocString).
type CondBranch struct {
	Cond *Cond
	Type string
}

// String renders c in the canonical form ParseCond reads back.
func (c *Cond) String() string {
	var b strings.Builder
	b.WriteString("($")
	b.WriteString(c.Param)
	b.WriteString(" is ")
	if c.Negated {
		b.WriteString("not ")
	}
	if !c.Exact && c.TargetKind() == TargetType {
		b.WriteByte('~')
	}
	b.WriteString(c.Target)
	b.WriteString(" ? ")
	b.WriteString(c.Then.String())
	b.WriteString(" : ")
	b.WriteString(c.Else.String())
	b.WriteByte(')')
	return b.String()
}

func (br CondBranch) String() string {
	if br.Cond != nil {
		return br.Cond.String()
	}
	return br.Type
}

// TargetKind classifies Cond.Target.
type TargetKind uint8

// Target kinds.
const (
	TargetType   TargetKind = iota // a doc type
	TargetString                   // a quoted string literal
	TargetNumber                   // an integer or float literal
	TargetConst                    // a class constant `\Foo::BAR`
)

// TargetKind returns the kind of c.Target.
func (c *Cond) TargetKind() TargetKind {
	return targetKind(c.Target)
}

func targetKind(s string) TargetKind {
	switch {
	case s[0] == '\'' || s[0] == '"':
		return TargetString
	case s[0] >= '0' && s[0] <= '9' || s[0] == '-':
		return TargetNumber
	case strings.Contains(s, "::"):
		return TargetConst
	}
	return TargetType
}

// ParseCond parses a conditional type. Names are resolved with resolve
// (nil: as written, for the canonical form); subject maps a non-variable
// subject (a template name) to the parameter it stands for ("" when none).
// ok is false when text is not a conditional on a parameter, or exceeds
// the doc-type caps (MaxDocTypeLen, nesting depth).
func ParseCond(text string, resolve Resolver, subject func(string) string) (*Cond, bool) {
	text = strings.TrimSpace(text)
	if len(text) > MaxDocTypeLen {
		return nil, false
	}
	return parseCond(text, resolve, subject, 0)
}

func parseCond(text string, resolve Resolver, subject func(string) string, depth int) (*Cond, bool) {
	if depth > maxDocDepth || len(text) < 2 || text[0] != '(' || matchingClose(text) != len(text)-1 {
		return nil, false
	}
	inner := strings.TrimSpace(text[1 : len(text)-1])
	test, a, b, ok := splitConditional(inner)
	if !ok {
		return nil, false
	}
	subj, target, found := strings.Cut(test, " is ")
	if !found {
		return nil, false // `is X ? …`: no subject
	}
	c := &Cond{}
	subj = strings.TrimSpace(subj)
	switch {
	case strings.HasPrefix(subj, "$"):
		c.Param = subj[1:]
	case subject != nil:
		c.Param = strings.TrimPrefix(subject(subj), "$")
	}
	if !validIdent(c.Param) {
		return nil, false
	}
	target = strings.TrimSpace(target)
	if rest, ok := strings.CutPrefix(target, "not "); ok {
		c.Negated, target = true, strings.TrimSpace(rest)
	}
	if !c.setTarget(target, resolve) {
		return nil, false
	}
	var okA, okB bool
	c.Then, okA = parseBranch(a, resolve, subject, depth)
	c.Else, okB = parseBranch(b, resolve, subject, depth)
	return c, okA && okB
}

// setTarget records the test's target (see Cond.Target).
func (c *Cond) setTarget(s string, resolve Resolver) bool {
	switch targetKind(s) {
	case TargetString:
		if len(s) < 2 || s[len(s)-1] != s[0] || strings.ContainsAny(s[1:len(s)-1], `'"\`) {
			return false
		}
		c.Target = s
		return true
	case TargetNumber:
		if !isNumberLiteral(s) {
			return false
		}
		c.Target = s
		return true
	case TargetConst:
		cls, name, _ := strings.Cut(s, "::")
		if !validClassName(cls) || !validIdent(name) {
			return false
		}
		fqn := strings.TrimPrefix(cls, `\`)
		if resolve != nil {
			if fqn = resolve(cls); fqn == "" || strings.HasPrefix(fqn, "=") {
				return false
			}
		}
		c.Target = `\` + fqn + "::" + name
		return true
	}
	exact := true
	if s[0] == '~' {
		exact, s = false, s[1:]
	}
	t := FromDoc(s, resolve)
	if t.IsUnknown() {
		return false
	}
	c.Target, c.Exact = t.DocString(), exact && exactType(s, resolve)
	return true
}

// parseBranch parses one outcome: a nested conditional or a doc type.
func parseBranch(s string, resolve Resolver, subject func(string) string, depth int) (CondBranch, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "(") && matchingClose(s) == len(s)-1 {
		if _, _, _, ok := splitConditional(strings.TrimSpace(s[1 : len(s)-1])); ok {
			c, ok := parseCond(s, resolve, subject, depth+1)
			return CondBranch{Cond: c}, ok
		}
	}
	t := FromDoc(s, resolve)
	if t.IsUnknown() {
		return CondBranch{}, false
	}
	return CondBranch{Type: t.DocString()}, true
}

// exactType reports a doc type made of builtin type names and plain class
// names (`?`, `|` allowed), which its atoms describe fully.
func exactType(s string, resolve Resolver) bool {
	for _, part := range splitTop(s, '|') {
		part = strings.TrimPrefix(strings.TrimSpace(part), "?")
		low := strings.ToLower(part)
		switch {
		case low == "scalar" || low == "array-key" || low == "number":
			continue
		case low == "static" || low == "self" || low == "parent" || low == "void" || low == "never":
			return false
		case isBuiltinAtom(low):
			continue
		case scalarAliases[low] != "" && low != "$this" && low != "callback":
			continue
		case pseudo[low] != nil, !validClassName(part):
			return false
		}
		if resolve != nil {
			if fqn := resolve(part); fqn == "" || strings.HasPrefix(fqn, "=") {
				return false // a template or a type alias
			}
		}
	}
	return true
}

// splitConditional splits `test ? a : b` at the top level. The `?` and
// `:` separators must stand between blanks (doc tags collapse the spaces
// of a conditional type to one), so `?int` and `callable(): T` parse as
// types.
func splitConditional(s string) (test, a, b string, ok bool) {
	q := topSep(s, '?', 0)
	if q < 0 {
		return "", "", "", false
	}
	test = strings.TrimSpace(s[:q])
	if !strings.Contains(" "+test+" ", " is ") {
		return "", "", "", false
	}
	c := topSep(s, ':', q+1)
	if c < 0 {
		return "", "", "", false
	}
	a, b = strings.TrimSpace(s[q+1:c]), strings.TrimSpace(s[c+1:])
	if a == "" || b == "" {
		return "", "", "", false
	}
	return test, a, b, true
}

// topSep returns the index of the first sep at nesting depth 0 at or after
// from that stands between blanks, or -1.
func topSep(s string, sep byte, from int) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"':
			end := quotedEnd(s, i)
			if end < 0 {
				return -1
			}
			i = end
		case '<', '(', '{', '[':
			depth++
		case '>', ')', '}', ']':
			if depth > 0 {
				depth--
			}
		default:
			if c == sep && depth == 0 && i >= from && i > 0 && isBlank(s[i-1]) && (i+1 == len(s) || isBlank(s[i+1])) {
				return i
			}
		}
	}
	return -1
}

// validIdent reports a PHP identifier.
func validIdent(s string) bool {
	if s == "" || s[0] >= '0' && s[0] <= '9' {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c >= 0x80) {
			return false
		}
	}
	return true
}

// isNumberLiteral reports an integer or decimal float literal (optionally
// negative).
func isNumberLiteral(s string) bool {
	s = strings.TrimPrefix(s, "-")
	digits, dot := 0, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
			digits++
		case c == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return digits > 0
}

func isBlank(c byte) bool { return c == ' ' || c == '\t' || c == '\n' }
