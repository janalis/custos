package pattern

import (
	"maps"
	"reflect"
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// binding is what a metavariable matched: one node, or a run of call
// arguments for `...$REST` (possibly empty, at span's start).
type binding struct {
	node syntax.Node
	args []syntax.Expr
	seq  bool
	span syntax.Span
}

// matcher matches one pattern against code of one file.
type matcher struct {
	ctx   *analysis.Context
	binds map[string]binding
	// steps counts the statement comparisons of this match; past
	// maxSteps it stops, unmatched, and exhausted is set.
	steps     int
	exhausted bool
}

// maxSteps bounds the statement comparisons of one match: each `$_;` may
// try every split of a statement list, and patterns come from the
// analyzed project.
const maxSteps = 100000

// nameRole is how a name position resolves.
type nameRole uint8

const (
	roleClass nameRole = iota
	roleFunction
	roleConst
)

var (
	spanType      = reflect.TypeFor[syntax.Span]()
	modifiersType = reflect.TypeFor[syntax.Modifiers]()
	stmtsType     = reflect.TypeFor[[]syntax.Stmt]()
)

// stmts matches pattern statements against target statements, all of them
// or, with prefix, a first run of them; it returns how many statements
// matched. `$_;` matches any run of statements, as short as possible.
func (m *matcher) stmts(ps, ts []syntax.Stmt, prefix bool) (int, bool) {
	if m.steps++; m.steps > maxSteps {
		m.exhausted = true
		return 0, false
	}
	if len(ps) == 0 {
		return 0, prefix || len(ts) == 0
	}
	if stmtMeta(ps[0]) == wildcard {
		for k := 0; k <= len(ts); k++ {
			saved := maps.Clone(m.binds)
			if n, ok := m.stmts(ps[1:], ts[k:], prefix); ok {
				return k + n, true
			}
			m.binds = saved
			if m.exhausted {
				break
			}
		}
		return 0, false
	}
	if len(ts) == 0 || !m.match(ps[0], ts[0], roleClass, false) {
		return 0, false
	}
	n, ok := m.stmts(ps[1:], ts[1:], prefix)
	return n + 1, ok
}

// gap records that the match at span stopped at its step budget: the
// file's analysis is then incomplete.
func (m *matcher) gap(span syntax.Span) {
	if m.exhausted {
		m.ctx.Gap("query-budget", span)
	}
}

// enclosed reports whether a node enclosing n matches one of patterns,
// keeping the bindings of the first that does.
func (m *matcher) enclosed(patterns []*snippet, n syntax.Node) bool {
	for a := n.Parent(); a != nil; a = a.Parent() {
		for _, p := range patterns {
			if p.root.Kind() != a.Kind() {
				continue
			}
			saved := maps.Clone(m.binds)
			if m.match(p.root, a, roleClass, false) {
				return true
			}
			m.binds = saved
		}
	}
	return false
}

// sameModifiers compares modifier lists regardless of their order.
func sameModifiers(a, b syntax.Modifiers) bool {
	if len(a) != len(b) {
		return false
	}
	for _, t := range a {
		if !b.Has(t.Kind) {
			return false
		}
	}
	return true
}

// match reports whether target matches pattern p, recording bindings.
func (m *matcher) match(p, t syntax.Node, role nameRole, foldIdent bool) bool {
	if name := metaVarName(p); name != "" {
		return m.bind(name, binding{node: t, span: t.Span()})
	}
	if name := stmtMeta(p); name != "" {
		// `$S;` stands for any one statement.
		if _, ok := t.(syntax.Stmt); ok {
			return m.bind(name, binding{node: t, span: t.Span()})
		}
	}
	if p.Kind() != syntax.KParen {
		for t.Kind() == syntax.KParen {
			t = t.(*syntax.Paren).Expr
		}
	}
	if p.Kind() != t.Kind() {
		return false
	}
	switch pn := p.(type) {
	case *syntax.Name:
		return m.name(pn, t.(*syntax.Name), role)
	case *syntax.Identifier:
		tn := t.(*syntax.Identifier)
		if foldIdent {
			return strings.EqualFold(pn.Value, tn.Value)
		}
		return pn.Value == tn.Value
	case *syntax.Variable:
		tv := t.(*syntax.Variable)
		if pn.NameExpr == nil || tv.NameExpr == nil {
			return pn.NameExpr == nil && tv.NameExpr == nil && pn.Name == tv.Name
		}
		return m.match(pn.NameExpr, tv.NameExpr, role, false)
	case *syntax.Literal:
		return literalEqual(pn, t.(*syntax.Literal))
	case *syntax.ArgList:
		return m.args(pn, t.(*syntax.ArgList))
	case *syntax.New:
		// `new Foo` and `new Foo()` are the same call.
		tn := t.(*syntax.New)
		if !m.match(pn.Class, tn.Class, roleClass, false) {
			return false
		}
		switch {
		case pn.Args == nil:
			return tn.Args == nil || len(tn.Args.Args) == 0
		case tn.Args == nil:
			return m.argsAt(pn.Args, nil, tn.Span().End)
		}
		return m.args(pn.Args, tn.Args)
	}
	return m.fields(reflect.ValueOf(p).Elem(), reflect.ValueOf(t).Elem())
}

// fields matches the exported fields of two nodes of the same type.
func (m *matcher) fields(pv, tv reflect.Value) bool {
	typ := pv.Type()
	for i := range typ.NumField() {
		f := typ.Field(i)
		if f.Anonymous || !f.IsExported() || f.Type == spanType || spelling(typ.Name(), f.Name) {
			continue
		}
		if typ.Name() == "Closure" && f.Name == "Uses" && pv.Field(i).Len() == 0 {
			continue // a closure written without `use` stands for any
		}
		if f.Type == stmtsType {
			if _, ok := m.stmts(pv.Field(i).Interface().([]syntax.Stmt), tv.Field(i).Interface().([]syntax.Stmt), false); !ok {
				return false
			}
			continue
		}
		if f.Type == modifiersType {
			// `public static` and `static public` are the same.
			if !sameModifiers(pv.Field(i).Interface().(syntax.Modifiers), tv.Field(i).Interface().(syntax.Modifiers)) {
				return false
			}
			continue
		}
		role, fold := fieldRole(typ.Name(), f.Name)
		if !m.value(pv.Field(i), tv.Field(i), role, fold) {
			return false
		}
	}
	return true
}

// spelling reports fields that tell only how equivalent code is written:
// `[]` or `array()`, `exit` or `die`, braces or the alternative syntax.
func spelling(owner, field string) bool {
	switch field {
	case "Short", "Alt":
		return true
	case "Keyword":
		return owner == "Exit"
	}
	return false
}

// fieldRole tells how names in a field resolve and whether its identifiers
// compare as PHP compares method names.
func fieldRole(owner, field string) (nameRole, bool) {
	switch {
	case owner == "FuncCall" && field == "Name":
		return roleFunction, false
	case owner == "ConstFetch":
		return roleConst, false
	case (owner == "MethodCall" || owner == "StaticCall") && field == "Name":
		return roleClass, true
	}
	return roleClass, false
}

func (m *matcher) value(p, t reflect.Value, role nameRole, fold bool) bool {
	switch p.Kind() {
	case reflect.Interface, reflect.Pointer:
		// Every pointer and interface field of the syntax tree holds a node.
		pn, _ := p.Interface().(syntax.Node)
		tn, _ := t.Interface().(syntax.Node)
		if isNil(pn) || isNil(tn) {
			return isNil(pn) && isNil(tn)
		}
		return m.match(pn, tn, role, fold)
	case reflect.Slice:
		if p.Len() != t.Len() {
			return false
		}
		for i := range p.Len() {
			if !m.value(p.Index(i), t.Index(i), role, fold) {
				return false
			}
		}
		return true
	case reflect.Struct:
		// Spans are skipped: the only other struct is a token reference,
		// whose first field is its kind.
		return p.Field(0).Uint() == t.Field(0).Uint()
	}
	return p.Equal(t)
}

// args matches argument lists: `...` ends a pattern list matching any
// remaining arguments, and `...$REST` binds them.
func (m *matcher) args(p, t *syntax.ArgList) bool {
	return m.argsAt(p, t.Args, t.Span().End-1)
}

// argsAt matches pattern arguments against targets; an empty run of
// arguments binds at end.
func (m *matcher) argsAt(p *syntax.ArgList, targets []syntax.Expr, end uint32) bool {
	for i, pa := range p.Args {
		if pa.Kind() == syntax.KVariadicPlaceholder && i == len(p.Args)-1 {
			return true
		}
		if a, ok := pa.(*syntax.Arg); ok && a.Unpack && a.Name == nil && i == len(p.Args)-1 {
			if name := metaVarName(a.Value); name != "" {
				rest := targets[min(i, len(targets)):]
				span := syntax.Span{Start: end, End: end}
				if len(rest) > 0 {
					span = syntax.Span{Start: rest[0].Span().Start, End: rest[len(rest)-1].Span().End}
				}
				return m.bind(name, binding{args: rest, seq: true, span: span})
			}
		}
		if i >= len(targets) || !m.match(pa, targets[i], roleClass, false) {
			return false
		}
	}
	return len(p.Args) == len(targets)
}

// bind records a metavariable; a repeated one must bind equivalent code.
func (m *matcher) bind(name string, b binding) bool {
	if name == wildcard {
		return true
	}
	prev, ok := m.binds[name]
	if !ok {
		m.binds[name] = b
		return true
	}
	if prev.seq || b.seq {
		if !prev.seq || !b.seq || len(prev.args) != len(b.args) {
			return false
		}
		for i := range prev.args {
			if !astquery.EquivalentFoldNames(m.ctx.File, prev.args[i], b.args[i]) {
				return false
			}
		}
		return true
	}
	return astquery.EquivalentFoldNames(m.ctx.File, prev.node, b.node)
}

// name matches a fully qualified pattern name against a target name as PHP
// resolves it.
func (m *matcher) name(p, t *syntax.Name, role nameRole) bool {
	want := strings.TrimPrefix(p.Value, `\`)
	ctx := m.ctx
	switch role {
	case roleFunction:
		fqn, fallback := m.resolved(functionKey{t})
		if fallback == "" {
			return strings.EqualFold(fqn, want)
		}
		// An unqualified call reaches the namespaced function when one is
		// declared, else the global one.
		if !strings.EqualFold(fqn, want) && !strings.EqualFold(fallback, want) {
			return false
		}
		declared := ctx.Index().Function(fqn, ctx.PHP) != nil
		return declared == strings.EqualFold(fqn, want)
	case roleConst:
		fqn, fallback := m.resolved(constKey{t})
		if fallback == "" {
			return constEqual(fqn, want)
		}
		if !constEqual(fqn, want) && !constEqual(fallback, want) {
			return false
		}
		declared := ctx.Index().Constant(fqn, ctx.PHP) != nil
		return declared == constEqual(fqn, want)
	}
	fqn, _ := m.resolved(classKey{t})
	return fqn != "" && strings.EqualFold(strings.TrimPrefix(fqn, `\`), want)
}

// Keys of the name resolutions shared by every custom rule of a file.
type (
	functionKey struct{ n *syntax.Name }
	constKey    struct{ n *syntax.Name }
	classKey    struct{ n *syntax.Name }
)

// resolution is a resolved name and its global fallback.
type resolution struct{ fqn, fallback string }

// resolved resolves a name once per file for all custom rules.
func (m *matcher) resolved(key any) (string, string) {
	r := m.ctx.Shared(key, func() any {
		names := m.ctx.Names()
		switch k := key.(type) {
		case functionKey:
			fqn, fallback := names.Function(k.n.Value, k.n.Span().Start)
			return resolution{fqn, fallback}
		case constKey:
			fqn, fallback := names.Const(k.n.Value, k.n.Span().Start)
			return resolution{fqn, fallback}
		}
		t := key.(classKey).n
		switch strings.ToLower(t.Value) {
		case "self", "static":
			return resolution{fqn: names.DeclFQN(syntax.EnclosingClass(t))}
		case "parent":
			return resolution{fqn: names.ParentFQN(syntax.EnclosingClass(t))}
		}
		return resolution{fqn: names.Class(t.Value, t.Span().Start)}
	}).(resolution)
	return r.fqn, r.fallback
}

// constEqual compares constant names: namespaces fold case, as do true,
// false and null; other constant names are case-sensitive.
func constEqual(a, b string) bool {
	ia, ib := strings.LastIndexByte(a, '\\'), strings.LastIndexByte(b, '\\')
	if !strings.EqualFold(a[:ia+1], b[:ib+1]) {
		return false
	}
	na, nb := a[ia+1:], b[ib+1:]
	switch strings.ToLower(na) {
	case "true", "false", "null":
		return strings.EqualFold(na, nb)
	}
	return na == nb
}

// literalEqual compares literals by value: strings decoded, numbers as
// written.
func literalEqual(p, t *syntax.Literal) bool {
	if p.LitKind != t.LitKind {
		return false
	}
	if p.LitKind != syntax.LitString {
		return p.Raw == t.Raw
	}
	pd, ok1 := syntax.DecodeLiteralString(p)
	td, ok2 := syntax.DecodeLiteralString(t)
	return ok1 == ok2 && pd.Value == td.Value && (ok1 || p.Raw == t.Raw)
}

func isNil(n syntax.Node) bool {
	if n == nil {
		return true
	}
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Pointer && v.IsNil()
}
