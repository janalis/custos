package pattern

import (
	"reflect"
	"strings"
	"sync"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/astquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/names"
)

// binding is what a metavariable matched: one node, or a run of call
// arguments for `...$REST` (possibly empty, at span's start).
type binding struct {
	node syntax.Node
	args []syntax.Expr
	seq  bool
	span syntax.Span
	role nameRole // how node resolves, when it is a name
}

// matcher matches one pattern against code of one file.
type matcher struct {
	ctx    *analysis.Context
	binds  []named // in binding order; backtracking truncates
	budget *budget // shared by every matcher of the rule in the file
}

// budget bounds the statement comparisons one rule makes in one file: each
// `$_;` may try every split of a statement list, and patterns come from
// the analyzed project. Every matcher of the rule draws from it; once it
// is spent, the rule reports nothing more in the file.
type budget struct {
	steps    int
	spent    bool
	reported bool    // the file's analysis was marked incomplete
	binds    []named // storage the rule's top-level matchers reuse, one at a time
}

// maxSteps is a rule's budget per file.
const maxSteps = 1000000

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
	if m.budget.steps++; m.budget.steps > maxSteps {
		m.budget.spent = true
		return 0, false
	}
	if len(ps) == 0 {
		return 0, prefix || len(ts) == 0
	}
	if stmtMeta(ps[0]) == wildcard {
		for k := 0; k <= len(ts); k++ {
			saved := len(m.binds)
			if n, ok := m.stmts(ps[1:], ts[k:], prefix); ok {
				return k + n, true
			}
			m.binds = m.binds[:saved]
			if m.budget.spent {
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

// enclosed reports whether a node enclosing n matches one of patterns,
// keeping the bindings of the first that does.
func (m *matcher) enclosed(patterns []*snippet, n syntax.Node) bool {
	for a := n.Parent(); a != nil; a = a.Parent() {
		for _, p := range patterns {
			if p.root.Kind() != a.Kind() {
				continue
			}
			saved := len(m.binds)
			if m.match(p.root, a, roleClass, false) {
				return true
			}
			m.binds = m.binds[:saved]
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
		return m.bind(name, binding{node: t, span: t.Span(), role: role})
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
// Fields holding a metavariable are matched last: the other fields reject
// most candidates without binding anything.
func (m *matcher) fields(pv, tv reflect.Value) bool {
	plan := planOf(pv.Type())
	deferred := false
	for _, f := range plan {
		if metaVarField(pv.Field(f.index)) {
			deferred = true
			continue
		}
		if !m.field(pv, tv, f) {
			return false
		}
	}
	for i := 0; deferred && i < len(plan); i++ {
		if metaVarField(pv.Field(plan[i].index)) && !m.field(pv, tv, plan[i]) {
			return false
		}
	}
	return true
}

// fieldKind is how a field of a node compares.
type fieldKind uint8

const (
	compareValue     fieldKind = iota // nodes, lists, tokens and plain values
	compareStmts                      // a statement list, with wildcards
	compareModifiers                  // modifiers, in any order
	compareUses                       // a closure's use list, any when the pattern has none
)

// fieldPlan is one compared field of a node type.
type fieldPlan struct {
	index int
	kind  fieldKind
	role  nameRole
	fold  bool
}

// plans caches the fields each node type compares (reflection on types is
// slow, and every match walks them).
var plans sync.Map // reflect.Type -> []fieldPlan

// planOf returns the compared fields of node type typ.
func planOf(typ reflect.Type) []fieldPlan {
	if p, ok := plans.Load(typ); ok {
		return p.([]fieldPlan)
	}
	var plan []fieldPlan
	for i := range typ.NumField() {
		f := typ.Field(i)
		if f.Anonymous || !f.IsExported() || f.Type == spanType || spelling(typ.Name(), f.Name) {
			continue
		}
		fp := fieldPlan{index: i}
		switch {
		case typ.Name() == "Closure" && f.Name == "Uses":
			fp.kind = compareUses
		case f.Type == stmtsType:
			fp.kind = compareStmts
		case f.Type == modifiersType:
			fp.kind = compareModifiers
		}
		fp.role, fp.fold = fieldRole(typ.Name(), f.Name)
		plan = append(plan, fp)
	}
	plans.Store(typ, plan)
	return plan
}

// field matches one field of two nodes of the same type.
func (m *matcher) field(pv, tv reflect.Value, f fieldPlan) bool {
	p, t := pv.Field(f.index), tv.Field(f.index)
	switch f.kind {
	case compareUses:
		if p.Len() == 0 {
			return true // a closure written without `use` stands for any
		}
	case compareStmts:
		_, ok := m.stmts(p.Interface().([]syntax.Stmt), t.Interface().([]syntax.Stmt), false)
		return ok
	case compareModifiers:
		// `public static` and `static public` are the same.
		return sameModifiers(p.Interface().(syntax.Modifiers), t.Interface().(syntax.Modifiers))
	}
	return m.value(p, t, f.role, f.fold)
}

// metaVarField reports a field holding a metavariable.
func metaVarField(v reflect.Value) bool {
	if v.Kind() != reflect.Interface && v.Kind() != reflect.Pointer || v.IsNil() {
		return false
	}
	n, ok := v.Interface().(syntax.Node)
	return ok && metaVarName(n) != ""
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

// named is a metavariable's binding.
type named struct {
	name string
	b    binding
}

// get returns what metavariable name bound.
func (m *matcher) get(name string) (binding, bool) {
	for _, n := range m.binds {
		if n.name == name {
			return n.b, true
		}
	}
	return binding{}, false
}

// lookup returns what metavariable name, which compile checked every
// pattern binds, bound.
func (m *matcher) lookup(name string) binding {
	b, _ := m.get(name)
	return b
}

// bind records a metavariable; a repeated one must bind equivalent code.
func (m *matcher) bind(name string, b binding) bool {
	if name == wildcard {
		return true
	}
	prev, ok := m.get(name)
	if !ok {
		m.binds = append(m.binds, named{name, b})
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
	if pn, ok := prev.node.(*syntax.Name); ok {
		if tn, ok := b.node.(*syntax.Name); ok {
			// Two spellings of one name: an alias and the full name. A
			// function is not a class of the same name, and self, static
			// and parent only equal themselves: static binds late, and self
			// in a trait is the using class.
			switch {
			case prev.role != b.role:
				return false
			case names.IsSpecialClass(pn.Value) || names.IsSpecialClass(tn.Value):
				return strings.EqualFold(pn.Value, tn.Value)
			}
			return strings.EqualFold(m.fqn(pn, prev.role), m.fqn(tn, b.role))
		}
	}
	return astquery.EquivalentFoldNames(m.ctx.File, prev.node, b.node)
}

// fqn returns the fully qualified name, without a leading backslash, that
// name n at a position of role resolves to: for an unqualified call in a
// namespace, the namespaced function when one is declared, else the global
// one; "" for a class name nothing resolves (static outside a class). A
// metavariable never stands for a constant's name: the pattern `$K` is the
// whole constant fetch.
func (m *matcher) fqn(n *syntax.Name, role nameRole) string {
	if role == roleFunction {
		fqn, fallback := m.resolved(functionKey{n})
		if fallback != "" && m.ctx.Index().Function(fqn, m.ctx.PHP) == nil {
			return fallback
		}
		return fqn
	}
	fqn, _ := m.resolved(classKey{n})
	return strings.TrimPrefix(fqn, `\`)
}

// name matches a fully qualified pattern name against a target name as PHP
// resolves it.
func (m *matcher) name(p, t *syntax.Name, role nameRole) bool {
	want := strings.TrimPrefix(p.Value, `\`)
	if !m.mayResolveTo(t, want, role) {
		return false
	}
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

// mayResolveTo reports whether target name t may resolve to want, without
// resolving it: resolving keeps a name's last segment, except for an
// unqualified name an import aliases, and for self, static and parent.
func (m *matcher) mayResolveTo(t *syntax.Name, want string, role nameRole) bool {
	if strings.EqualFold(lastSegment(t.Value), lastSegment(want)) {
		return true
	}
	if strings.Contains(t.Value, `\`) {
		return false
	}
	switch role {
	case roleFunction:
		return m.ctx.Names().Aliases(syntax.UseFunction, t.Value)
	case roleConst:
		return m.ctx.Names().Aliases(syntax.UseConst, t.Value)
	}
	return names.IsSpecialClass(t.Value) || m.ctx.Names().Aliases(syntax.UseNormal, t.Value)
}

// lastSegment returns the part of a name after its last backslash.
func lastSegment(name string) string {
	return name[strings.LastIndexByte(name, '\\')+1:]
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
