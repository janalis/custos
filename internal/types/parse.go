package types

import (
	"strconv"
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

// Limits on PHPDoc type parsing. Doc comments come from analysed (possibly
// hostile) code: parsing is linear in the input (each sub-expression is
// parsed once), and these caps bound the work on pathological input.
const (
	// MaxDocTypeLen is the longest doc type parsed; longer ones are unknown.
	MaxDocTypeLen = 4096
	// maxDocDepth caps the nesting of generic arguments, shapes, `[]`,
	// parentheses and alias expansions; deeper parts read as mixed.
	maxDocDepth = 32
	// maxDocParts caps the number of type parts one FromDoc call parses
	// (alias expansions included); beyond it parts read as mixed.
	maxDocParts = 4096
	// maxDocBytes caps the text one FromDoc call scans, summed over every
	// union it parses (nested members and alias or template expansions are
	// scanned again); beyond it parts read as mixed. An expansion longer
	// than MaxDocTypeLen reads as mixed too.
	maxDocBytes = 64 * MaxDocTypeLen
)

// docParser carries the work budget of one FromDoc call.
type docParser struct {
	budget int // parts left
	bytes  int // text bytes left to scan
}

// FromDoc parses a PHPDoc type expression.
func FromDoc(text string, resolve Resolver) Type {
	text = strings.TrimSpace(text)
	if text == "" || len(text) > MaxDocTypeLen {
		return Unknown
	}
	p := docParser{budget: maxDocParts, bytes: maxDocBytes}
	return p.union(text, resolve, 0)
}

// union parses a `|`-separated type: the atoms of its members, and the
// array facts of the members with an array atom (see unionInfo).
func (p *docParser) union(text string, resolve Resolver, depth int) Type {
	text = strings.TrimSpace(text)
	if text == "" {
		return Unknown
	}
	if p.bytes -= len(text); p.bytes < 0 {
		return Mixed
	}
	parts := splitTop(text, '|')
	if len(parts) == 1 {
		return p.part(parts[0], resolve, depth)
	}
	var atoms []string
	var infos, pts []Type
	var gens [][]genEntry
	for _, part := range parts {
		pt := p.part(part, resolve, depth)
		pts = append(pts, pt)
		atoms = append(atoms, pt.atoms...)
		if pt.hasArrayAtom() {
			infos = append(infos, pt)
		}
		if pt.gen != nil {
			gens = append(gens, pt.gen)
		}
	}
	if len(atoms) == 0 {
		return Unknown
	}
	t := Of(atoms...)
	if len(infos) > 0 {
		t = t.withInfo(unionInfo(infos))
	}
	if len(gens) > 0 {
		t = t.withGen(mergeGen(gens...))
	}
	return t.withInter(unionInter(pts))
}

// part parses one union member: its atoms and array facts (shapes
// `array{k: T, k2?: U}` / `list{T, U}`, `non-empty-*` arrays, the facts of
// element types `array<K, array{...}>` / `array{...}[]`).
func (p *docParser) part(s string, resolve Resolver, depth int) Type {
	s = strings.TrimSpace(s)
	if s == "" {
		return Unknown
	}
	p.budget--
	if depth > maxDocDepth || p.budget < 0 {
		return Mixed
	}
	if s[0] == '?' {
		in := p.part(s[1:], resolve, depth+1)
		return Union(in, Null)
	}
	if s[0] == '(' && s[len(s)-1] == ')' && matchingClose(s) == len(s)-1 {
		if a, b, ok := conditionalBranches(s[1 : len(s)-1]); ok {
			ta, tb := p.union(a, resolve, depth+1), p.union(b, resolve, depth+1)
			return Union(ta, tb)
		}
		return p.union(s[1:len(s)-1], resolve, depth+1)
	}
	if parts := splitTop(s, '&'); len(parts) > 1 {
		var out, classes []string
		var gens [][]genEntry
		for _, x := range parts {
			pt := p.part(x, resolve, depth+1)
			out = append(out, pt.atoms...)
			gens = append(gens, pt.gen)
			if cs := pt.Classes(); len(cs) == 1 && len(pt.atoms) == 1 {
				classes = append(classes, cs[0])
			}
		}
		t := ofAtoms(out).withGen(mergeGen(gens...))
		if len(classes) == len(parts) {
			t = t.withInter(classes)
		}
		return t
	}
	// Literal types: 'foo', 1, 1.5
	if s[0] == '\'' || s[0] == '"' {
		return String
	}
	if s[0] >= '0' && s[0] <= '9' || s[0] == '-' {
		if strings.ContainsAny(s, ".eE") {
			return Float
		}
		return Int
	}
	// T[]
	if strings.HasSuffix(s, "[]") {
		in := p.part(s[:len(s)-2], resolve, depth+1)
		out := make([]string, 0, len(in.atoms))
		for _, a := range in.atoms {
			out = append(out, a+"[]")
		}
		t := ofAtoms(out)
		if in.arr != nil {
			t = t.withInfo(&arrayInfo{elem: in.arr})
		}
		return t
	}
	// Generic forms: name<...> and shapes name{...}, callable(...)
	base, args := s, ""
	if i := strings.IndexAny(s, "<{("); i > 0 {
		base, args = strings.TrimSpace(s[:i]), s[i:]
	}
	low := strings.ToLower(base)
	switch low {
	case "array", "list", "non-empty-array", "non-empty-list", "iterable":
		return p.arrayPart(low, args, resolve, depth)
	case "callable", "closure", "\\closure":
		atom := `\Closure`
		if low == "callable" {
			atom = "callable"
		}
		t := Of(atom)
		if ret, ok := p.callableReturn(args, resolve, depth); ok {
			t = t.WithTypeArgs(atom, []Type{ret})
		}
		return t
	case "int":
		return Int // int<0, max>
	case "class-string", "interface-string":
		// class-string<Foo>: a string carrying the class as generic
		// argument (bound by method templates, see ClassString).
		if strings.HasPrefix(args, "<") && matchingClose(args) == len(args)-1 {
			if in := p.union(args[1:len(args)-1], resolve, depth+1); isClassUnion(in) {
				return ClassString(in)
			}
		}
		return String
	}
	if ps, ok := pseudo[low]; ok {
		// A pseudo-type name that the file imports or declares as a class
		// (`use App\Number;` then `@param Number $n`) is that class.
		if resolve != nil && validClassName(base) {
			if fqn := resolve("!" + base); fqn != "" && !strings.ContainsAny(fqn, "!=~") {
				return Of(`\` + fqn)
			}
		}
		return Of(ps...)
	}
	if isBuiltinAtom(low) || scalarAliases[low] != "" {
		return Of(normalizeAtom(low))
	}
	if strings.HasPrefix(base, "$") || base == "" || !validClassName(base) {
		return Unknown
	}
	// Class name (generic arguments dropped).
	fqn := base
	if resolve != nil {
		fqn = resolve(base)
		if fqn == "" {
			// The resolver reports template parameters (@template T) as "".
			return Mixed
		}
		if strings.HasPrefix(fqn, "=") {
			// Type alias (@phpstan-type / @psalm-type): "=<definition>".
			def := fqn[1:]
			if def == "" || def == base || len(def) > MaxDocTypeLen {
				return Mixed
			}
			// Aliases may use other aliases (not themselves).
			return p.union(def, aliasGuard(resolve, base), depth+1)
		}
	} else {
		fqn = strings.TrimPrefix(base, `\`)
	}
	t := Of(`\` + fqn)
	if strings.HasPrefix(args, "<") && matchingClose(args) == len(args)-1 && strings.HasPrefix(t.atoms[0], `\`) {
		// Generic arguments: Collection<int, Foo>.
		t = t.WithTypeArgs(t.atoms[0], p.typeArgs(args[1:len(args)-1], resolve, depth))
	}
	return t
}

// callableReturn parses the return type of a callable signature
// `(params): R` (args is the text after `callable` / `Closure`). The
// parameters are not kept. A missing, unknown or mixed return type gives
// false: the signature then tells nothing about a call's result.
func (p *docParser) callableReturn(args string, resolve Resolver, depth int) (Type, bool) {
	if !strings.HasPrefix(args, "(") {
		return Unknown, false
	}
	end := matchingClose(args)
	if end < 0 {
		return Unknown, false
	}
	rest := strings.TrimSpace(args[end+1:])
	if !strings.HasPrefix(rest, ":") {
		return Unknown, false
	}
	ret := p.union(rest[1:], resolve, depth+1)
	if ret.IsUnknown() || ret.Has("mixed") {
		return Unknown, false
	}
	return ret, true
}

// typeArgs parses a comma-separated generic argument list.
func (p *docParser) typeArgs(inner string, resolve Resolver, depth int) []Type {
	parts := splitTop(inner, ',')
	out := make([]Type, len(parts))
	for i, a := range parts {
		out[i] = p.union(a, resolve, depth+1)
	}
	return out
}

// ofAtoms is Of for a possibly empty atom list (empty: unknown).
func ofAtoms(atoms []string) Type {
	if len(atoms) == 0 {
		return Unknown
	}
	return Of(atoms...)
}

// arrayPart parses `array`, `list`, `non-empty-array`, `non-empty-list` and
// `iterable`, with optional generic arguments `<K, V>` or a shape `{…}`.
func (p *docParser) arrayPart(low, args string, resolve Resolver, depth int) Type {
	var a arrayInfo
	a.nonEmpty = strings.HasPrefix(low, "non-empty-")
	end := matchingClose(args)
	closed := end == len(args)-1 && end > 0
	if strings.HasPrefix(args, "<") {
		inner := args[1:]
		if end > 0 {
			inner = args[1:end]
		}
		if low == "iterable" {
			// iterable<V> / iterable<K, V>
			return Of("iterable").WithTypeArgs("iterable", p.typeArgs(inner, resolve, depth))
		}
		gen := splitTop(inner, ',')
		// The element type may itself be a union (array<string, int|false>).
		el := p.union(gen[len(gen)-1], resolve, depth+1)
		out := make([]string, 0, len(el.atoms))
		for _, x := range el.atoms {
			out = append(out, x+"[]")
		}
		if len(out) == 0 {
			out = append(out, "array")
		}
		if closed {
			a.elem = el.arr
		}
		return Of(out...).withInfo(&a)
	}
	if low == "iterable" {
		return Of("iterable")
	}
	if closed && args[0] == '{' {
		keys, sealed, ok := p.shape(args[1:end], resolve, depth+1)
		switch {
		case !ok:
		case len(keys) > MaxShapeKeys:
			for _, k := range keys {
				a.nonEmpty = a.nonEmpty || !k.Optional
			}
		default:
			a.shape, a.sealed, a.keys = true, sealed, keys
		}
	}
	return Array.withInfo(&a)
}

// shape parses the body of a doc shape: `k: T, 'k2'?: U, 0: V, ...`
// or positional `T, U` (keys 0, 1, ...). A trailing `...` unseals it.
func (p *docParser) shape(body string, resolve Resolver, depth int) ([]ShapeKey, bool, bool) {
	sealed := true
	var keys []ShapeKey
	next := 0
	for _, entry := range splitTop(body, ',') {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, "...") {
			sealed = false
			continue
		}
		kv := splitTop(entry, ':')
		var k ShapeKey
		if len(kv) >= 2 && isShapeKey(strings.TrimSpace(kv[0])) {
			name := strings.TrimSpace(kv[0])
			if strings.HasSuffix(name, "?") {
				k.Optional = true
				name = strings.TrimSpace(name[:len(name)-1])
			}
			if len(name) >= 2 && (name[0] == '\'' || name[0] == '"') && name[len(name)-1] == name[0] {
				name = name[1 : len(name)-1]
			}
			k.Name = name
			k.Type = p.union(strings.Join(kv[1:], ":"), resolve, depth)
			if IsIntKey(name) {
				if n, _ := strconv.Atoi(name); n >= next {
					next = n + 1
				}
			}
		} else {
			k.Name = strconv.Itoa(next)
			next++
			k.Type = p.union(entry, resolve, depth)
		}
		if _, dup := findKey(keys, k.Name); dup {
			return nil, false, false
		}
		keys = append(keys, k)
	}
	return keys, sealed, true
}

// isShapeKey reports whether s (trimmed, maybe with a trailing `?`) is a
// shape key: identifier-like, integer, or quoted.
func isShapeKey(s string) bool {
	s = strings.TrimSuffix(s, "?")
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0] {
		return true
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || c == '-' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80) {
			return false
		}
	}
	return true
}

// aliasGuard wraps resolve so that expanding alias name cannot recurse into
// itself (a self-reference reads as mixed).
func aliasGuard(resolve Resolver, name string) Resolver {
	return func(w string) string {
		if w == name {
			return ""
		}
		return resolve(w)
	}
}

// conditionalBranches splits a conditional type `T is [not] X ? A : B`
// (PHPStan/Psalm) into its two result branches.
func conditionalBranches(s string) (string, string, bool) {
	if _, a, b, ok := splitConditional(s); ok {
		return a, b, true // blank-separated: `?int` targets, `callable(): T` branches
	}
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
	return Of(nodeAtoms(n, resolve)...).withInter(nodeInter(n, resolve))
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

// validClassName reports whether s is a syntactically valid (possibly
// qualified) class name: `{array}`, `foo-bar` or `a\\b` in a doc tag are not
// types.
func validClassName(s string) bool {
	s = strings.TrimPrefix(s, `\`)
	// The index marks template parameters `~T` (class) and `~~T` (method).
	s = strings.TrimPrefix(strings.TrimPrefix(s, "~"), "~")
	start := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\':
			if start {
				return false
			}
			start = true
			continue
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80:
		case c >= '0' && c <= '9':
			if start {
				return false
			}
		default:
			return false
		}
		start = false
	}
	return !start
}
