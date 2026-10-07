package types

import (
	"strings"

	"custos/internal/syntax"
)

// Resolver resolves a class name as written to its FQN (no leading backslash).
type Resolver func(written string) string

// pseudo maps PHPDoc pseudo-types to atoms.
var pseudo = map[string][]string{
	"positive-int": {"int"}, "negative-int": {"int"}, "non-positive-int": {"int"}, "non-negative-int": {"int"},
	"non-zero-int": {"int"}, "non-empty-string": {"string"}, "numeric-string": {"string"},
	"literal-string": {"string"}, "lowercase-string": {"string"}, "non-falsy-string": {"string"},
	"truthy-string": {"string"}, "class-string": {"string"}, "interface-string": {"string"},
	"trait-string": {"string"}, "enum-string": {"string"}, "callable-string": {"string"},
	"non-empty-array": {"array"}, "list": {"array"}, "non-empty-list": {"array"},
	"scalar": {"int", "float", "string", "bool"}, "numeric": {"int", "float", "string"},
	"array-key": {"int", "string"}, "callable-array": {"array"}, "closed-resource": {"resource"},
	"open-resource": {"resource"}, "empty": {"mixed"}, "number": {"int", "float"},
	"no-return": {"never"}, "noreturn": {"never"}, "never-return": {"never"}, "never-returns": {"never"},
	"key-of": {"mixed"}, "value-of": {"mixed"}, "int-mask": {"int"}, "int-mask-of": {"int"},
}

// FromDoc parses a PHPDoc type expression.
func FromDoc(text string, resolve Resolver) Type {
	text = strings.TrimSpace(text)
	if text == "" {
		return Unknown
	}
	var atoms []string
	for _, part := range splitTop(text, '|') {
		atoms = append(atoms, docAtoms(part, resolve)...)
	}
	if len(atoms) == 0 {
		return Unknown
	}
	return Of(atoms...)
}

func docAtoms(s string, resolve Resolver) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "?") {
		return append(docAtoms(s[1:], resolve), "null")
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") && matchingClose(s) == len(s)-1 {
		if a, b, ok := conditionalBranches(s[1 : len(s)-1]); ok {
			return append(FromDoc(a, resolve).Atoms(), FromDoc(b, resolve).Atoms()...)
		}
		var out []string
		for _, p := range splitTop(s[1:len(s)-1], '|') {
			out = append(out, docAtoms(p, resolve)...)
		}
		return out
	}
	if parts := splitTop(s, '&'); len(parts) > 1 {
		var out []string
		for _, p := range parts {
			out = append(out, docAtoms(p, resolve)...)
		}
		return out
	}
	// Literal types: 'foo', 1, 1.5
	if s[0] == '\'' || s[0] == '"' {
		return []string{"string"}
	}
	if s[0] >= '0' && s[0] <= '9' || s[0] == '-' {
		if strings.ContainsAny(s, ".eE") {
			return []string{"float"}
		}
		return []string{"int"}
	}
	// T[]
	if strings.HasSuffix(s, "[]") {
		inner := docAtoms(s[:len(s)-2], resolve)
		out := make([]string, 0, len(inner))
		for _, a := range inner {
			out = append(out, a+"[]")
		}
		return out
	}
	// Generic forms: name<...> and shapes name{...}, callable(...)
	base, args := s, ""
	if i := strings.IndexAny(s, "<{("); i > 0 {
		base, args = s[:i], s[i:]
	}
	low := strings.ToLower(base)
	switch low {
	case "array", "list", "non-empty-array", "non-empty-list", "iterable":
		if strings.HasPrefix(args, "<") {
			inner := args[1:]
			if end := matchingClose(args); end > 0 {
				inner = args[1:end]
			}
			gen := splitTop(inner, ',')
			// The element type may itself be a union (array<string, int|false>).
			elem := FromDoc(gen[len(gen)-1], resolve).Atoms()
			if low == "iterable" {
				return []string{"iterable"}
			}
			out := make([]string, 0, len(elem))
			for _, a := range elem {
				out = append(out, a+"[]")
			}
			if len(out) == 0 {
				return []string{"array"}
			}
			return out
		}
		if low == "iterable" {
			return []string{"iterable"}
		}
		return []string{"array"}
	case "callable", "closure", "\\closure":
		if low == "callable" {
			return []string{"callable"}
		}
		return []string{`\Closure`}
	case "int":
		return []string{"int"} // int<0, max>
	}
	if p, ok := pseudo[low]; ok {
		return p
	}
	if isBuiltinAtom(low) || scalarAliases[low] != "" {
		return []string{normalizeAtom(low)}
	}
	if low == "$this" {
		return []string{"static"}
	}
	if strings.HasPrefix(base, "$") || base == "" {
		return nil
	}
	// Class name (generic arguments dropped).
	fqn := base
	if resolve != nil {
		fqn = resolve(base)
		if fqn == "" {
			// The resolver reports template parameters (@template T) as "".
			return []string{"mixed"}
		}
		if strings.HasPrefix(fqn, "=") {
			// Type alias (@phpstan-type / @psalm-type): "=<definition>".
			def := fqn[1:]
			if def == "" || def == base {
				return []string{"mixed"}
			}
			return FromDoc(def, nil).Atoms()
		}
	} else {
		fqn = strings.TrimPrefix(base, `\`)
	}
	return []string{`\` + fqn}
}

// conditionalBranches splits a conditional type `T is [not] X ? A : B`
// (PHPStan/Psalm) into its two result branches.
func conditionalBranches(s string) (string, string, bool) {
	parts := splitTop(s, '?')
	if len(parts) < 2 {
		return "", "", false
	}
	cond := " " + strings.TrimSpace(parts[0]) + " "
	if !strings.Contains(cond, " is ") {
		return "", "", false
	}
	rest := strings.Join(parts[1:], "?")
	branches := splitTop(rest, ':')
	if len(branches) < 2 {
		return "", "", false
	}
	a, b := strings.TrimSpace(branches[0]), strings.TrimSpace(strings.Join(branches[1:], ":"))
	if a == "" || b == "" {
		return "", "", false
	}
	return a, b, true
}

// splitTop splits s on sep at nesting depth 0.
func splitTop(s string, sep byte) []string {
	var out []string
	depth, start := 0, 0
	inStr := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr != 0 {
			if c == inStr {
				inStr = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			inStr = c
		case '<', '(', '{', '[':
			depth++
		case '>', ')', '}', ']':
			if depth > 0 {
				depth--
			}
		default:
			if c == sep && depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	return append(out, s[start:])
}

// FromNode converts a declared type node (param/return/property type).
func FromNode(n syntax.Expr, resolve Resolver) Type {
	if n == nil {
		return Unknown
	}
	return Of(nodeAtoms(n, resolve)...)
}

func nodeAtoms(n syntax.Expr, resolve Resolver) []string {
	switch t := n.(type) {
	case *syntax.NullableType:
		return append(nodeAtoms(t.Type, resolve), "null")
	case *syntax.UnionType:
		var out []string
		for _, x := range t.Types {
			out = append(out, nodeAtoms(x, resolve)...)
		}
		return out
	case *syntax.IntersectionType:
		var out []string
		for _, x := range t.Types {
			out = append(out, nodeAtoms(x, resolve)...)
		}
		return out
	case *syntax.Name:
		// Native types know no aliases: `integer`, `double`, `boolean` in a
		// declaration are class names.
		low := strings.ToLower(t.Value)
		if isBuiltinAtom(low) {
			return []string{normalizeAtom(low)}
		}
		fqn := strings.TrimPrefix(t.Value, `\`)
		if resolve != nil {
			fqn = resolve(t.Value)
		}
		return []string{`\` + fqn}
	}
	return nil
}

// matchingClose returns the index of the bracket closing s[0] ('<', '{', '(').
func matchingClose(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<', '{', '(', '[':
			depth++
		case '>', '}', ')', ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
