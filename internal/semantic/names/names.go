// Package names resolves names to fully qualified names using the
// namespace and `use` imports in effect at each position of a file.
package names

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

// Scope is one namespace region with its imports.
type Scope struct {
	Span      syntax.Span
	Namespace string
	Classes   map[string]importBinding
	Functions map[string]importBinding
	Consts    map[string]importBinding
}

// Resolver resolves names in one file.
type Resolver struct {
	scopes          []*Scope // ordered by Span.Start
	anonymousPrefix string
}

// New builds a resolver for f.
func New(f *syntax.File) *Resolver {
	r := &Resolver{anonymousPrefix: fmt.Sprintf("@anonymous:%x:", sha256.Sum256([]byte(f.Path)))}
	global := newScope(syntax.Span{Start: 0, End: uint32(len(f.Src))}, "")
	braced := map[*Scope]bool{}
	r.scopes = append(r.scopes, global)
	for _, st := range f.Stmts {
		switch n := st.(type) {
		case *syntax.Namespace:
			name := ""
			if n.Name != nil {
				name = strings.Trim(n.Name.Value, `\`)
			}
			sc := newScope(n.Span(), name)
			braced[sc] = n.Braced
			for _, inner := range n.Stmts {
				if u, ok := inner.(*syntax.Use); ok {
					sc.addUse(u)
				}
			}
			r.scopes = append(r.scopes, sc)
		case *syntax.Use:
			global.addUse(n)
		}
	}
	sort.SliceStable(r.scopes[1:], func(i, j int) bool { return r.scopes[1+i].Span.Start < r.scopes[1+j].Span.Start })
	// An unbraced namespace extends to the next namespace (or end of file).
	for i := 1; i < len(r.scopes); i++ {
		if braced[r.scopes[i]] {
			continue
		}
		end := uint32(len(f.Src))
		if i+1 < len(r.scopes) {
			end = r.scopes[i+1].Span.Start
		}
		r.scopes[i].Span.End = end
	}
	return r
}

type importBinding struct {
	start   uint32
	fqn     string
	history *[]importBinding
}

func imported(entries map[string]importBinding, alias string, offset uint32) (string, bool) {
	entry, ok := entries[alias]
	if !ok {
		return "", false
	}
	if entry.start > offset {
		if entry.history == nil {
			return "", false
		}
		history := *entry.history
		i := sort.Search(len(history), func(i int) bool { return history[i].start > offset })
		if i == 0 {
			return "", false
		}
		entry = history[i-1]
	}
	return entry.fqn, true
}

func bindImport(entries map[string]importBinding, alias, fqn string, at uint32) {
	entry := importBinding{start: at, fqn: fqn}
	if previous, ok := entries[alias]; ok {
		if previous.history == nil {
			history := []importBinding{previous}
			entry.history = &history
		} else {
			entry.history = previous.history
			*entry.history = append(*entry.history, importBinding{start: previous.start, fqn: previous.fqn})
		}
	}
	entries[alias] = entry
}

func newScope(span syntax.Span, ns string) *Scope {
	return &Scope{Span: span, Namespace: ns, Classes: map[string]importBinding{}, Functions: map[string]importBinding{}, Consts: map[string]importBinding{}}
}

func (s *Scope) addUse(u *syntax.Use) {
	prefix := ""
	if u.Prefix != nil {
		prefix = strings.Trim(u.Prefix.Value, `\`) + `\`
	}
	for _, it := range u.Items {
		fqn := prefix + strings.Trim(it.Name.Value, `\`)
		alias := fqn[strings.LastIndexByte(fqn, '\\')+1:]
		if it.Alias != nil {
			alias = it.Alias.Value
		}
		switch it.Type {
		case syntax.UseFunction:
			bindImport(s.Functions, strings.ToLower(alias), fqn, u.Span().End)
		case syntax.UseConst:
			bindImport(s.Consts, alias, fqn, u.Span().End)
		default:
			bindImport(s.Classes, strings.ToLower(alias), fqn, u.Span().End)
		}
	}
}

// ScopeAt returns the namespace scope containing offset.
func (r *Resolver) ScopeAt(offset uint32) *Scope {
	best := r.scopes[0]
	for _, s := range r.scopes[1:] {
		if s.Span.Start <= offset && offset < s.Span.End {
			best = s
		}
	}
	return best
}

// Namespace returns the namespace at offset ("" for global).
func (r *Resolver) Namespace(offset uint32) string { return r.ScopeAt(offset).Namespace }

// special class names that are not resolved through imports.
var special = map[string]bool{"self": true, "static": true, "parent": true}

// builtinTypes are type keywords that are never class names in type positions.
var builtinTypes = map[string]bool{
	"int": true, "float": true, "string": true, "bool": true, "array": true, "callable": true,
	"iterable": true, "object": true, "mixed": true, "void": true, "never": true, "null": true,
	"false": true, "true": true, "static": true, "self": true, "parent": true, "integer": true,
	"boolean": true, "double": true,
}

// IsBuiltinType reports whether name (as written) is a builtin type keyword.
func IsBuiltinType(name string) bool { return builtinTypes[strings.ToLower(name)] }

// IsSpecialClass reports whether name is self, static or parent.
func IsSpecialClass(name string) bool { return special[strings.ToLower(name)] }

// Class resolves a class-like name written at offset to its FQN (without
// leading backslash). self/static/parent are returned lower-cased as-is.
//
// A name written with a leading "!" asks whether it is explicitly imported
// at offset (`use App\Number;` for "!number", case-insensitively): the FQN,
// or "" when it is not (types.FromDoc then reads a pseudo-type such as
// `number` as the class only when imported).
func (r *Resolver) Class(written string, offset uint32) string {
	if probe, ok := strings.CutPrefix(written, "!"); ok {
		fqn, _ := imported(r.ScopeAt(offset).Classes, strings.ToLower(probe), offset)
		return fqn
	}
	if strings.HasPrefix(written, `\`) {
		return written[1:]
	}
	if IsSpecialClass(written) {
		return strings.ToLower(written)
	}
	sc := r.ScopeAt(offset)
	if rest, ok := cutRelative(written); ok {
		return join(sc.Namespace, rest)
	}
	first, rest, qualified := strings.Cut(written, `\`)
	if fqn, ok := imported(sc.Classes, strings.ToLower(first), offset); ok {
		if qualified {
			return fqn + `\` + rest
		}
		return fqn
	}
	return join(sc.Namespace, written)
}

// Function resolves a function name. For unqualified names in a namespace,
// PHP falls back to the global function at runtime: fallback holds the
// global candidate ("" when there is no fallback).
func (r *Resolver) Function(written string, offset uint32) (fqn, fallback string) {
	if strings.HasPrefix(written, `\`) {
		return written[1:], ""
	}
	sc := r.ScopeAt(offset)
	if rest, ok := cutRelative(written); ok {
		return join(sc.Namespace, rest), ""
	}
	if strings.Contains(written, `\`) {
		first, rest, _ := strings.Cut(written, `\`)
		if ns, ok := imported(sc.Classes, strings.ToLower(first), offset); ok {
			return ns + `\` + rest, ""
		}
		return join(sc.Namespace, written), ""
	}
	if fqn, ok := imported(sc.Functions, strings.ToLower(written), offset); ok {
		return fqn, ""
	}
	if sc.Namespace == "" {
		return written, ""
	}
	return join(sc.Namespace, written), written
}

// Const resolves a constant name, with the same fallback rule as Function.
func (r *Resolver) Const(written string, offset uint32) (fqn, fallback string) {
	if strings.HasPrefix(written, `\`) {
		return written[1:], ""
	}
	sc := r.ScopeAt(offset)
	if rest, ok := cutRelative(written); ok {
		return join(sc.Namespace, rest), ""
	}
	if strings.Contains(written, `\`) {
		first, rest, _ := strings.Cut(written, `\`)
		if ns, ok := imported(sc.Classes, strings.ToLower(first), offset); ok {
			return ns + `\` + rest, ""
		}
		return join(sc.Namespace, written), ""
	}
	if fqn, ok := imported(sc.Consts, written, offset); ok {
		return fqn, ""
	}
	if sc.Namespace == "" {
		return written, ""
	}
	return join(sc.Namespace, written), written
}

// IsGlobalFunction reports whether a call written as `written` at offset
// targets the global function `global` (case-insensitive), assuming no
// namespaced function of that name exists (callers with an index can refine).
func (r *Resolver) IsGlobalFunction(written string, offset uint32, global string) bool {
	fqn, fallback := r.Function(written, offset)
	if strings.EqualFold(fqn, global) {
		return true
	}
	return fallback != "" && strings.EqualFold(fallback, global)
}

func cutRelative(written string) (string, bool) {
	if len(written) > 10 && strings.EqualFold(written[:10], `namespace\`) {
		return written[10:], true
	}
	return "", false
}

func join(ns, name string) string {
	if ns == "" {
		return name
	}
	return ns + `\` + name
}

// DeclFQN returns the FQN (without leading backslash) of a named class-like
// declaration of the file, "" for anonymous classes and nil.
func (r *Resolver) DeclFQN(c *syntax.ClassLike) string {
	if c == nil || c.Name == nil {
		return ""
	}
	if ns := r.Namespace(c.Span().Start); ns != "" {
		return ns + `\` + c.Name.Value
	}
	return c.Name.Value
}

// SymbolFQN returns a semantic identity for named and anonymous declarations.
// Anonymous identities are internal and cannot be emitted as PHP class names.
func (r *Resolver) SymbolFQN(c *syntax.ClassLike) string {
	if c == nil || c.Name != nil {
		return r.DeclFQN(c)
	}
	return r.anonymousPrefix + strconv.FormatUint(uint64(c.Span().Start), 10)
}

// IsAnonymousClassName reports whether name denotes an internal anonymous class.
func IsAnonymousClassName(name string) bool {
	name = strings.TrimPrefix(name, `\`)
	const prefix = "@anonymous:"
	if !strings.HasPrefix(name, prefix) || len(name) <= len(prefix)+65 || name[len(prefix)+64] != ':' {
		return false
	}
	for _, c := range name[len(prefix) : len(prefix)+64] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	offset := name[len(prefix)+65:]
	for _, c := range offset {
		if c < '0' || c > '9' {
			return false
		}
	}
	_, err := strconv.ParseUint(offset, 10, 32)
	return err == nil
}

// ParentFQN resolves the `extends` clause of a class (not interface)
// declaration of the file; "" when there is none.
func (r *Resolver) ParentFQN(c *syntax.ClassLike) string {
	if c == nil || c.ClassKind == syntax.KindInterface || len(c.Extends) == 0 {
		return ""
	}
	return r.Class(c.Extends[0].Value, c.Span().Start)
}
