package infer

import (
	"sort"
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// Types of parameters, inline @var documentation and branch assignments.

func (e *Env) paramType(scope syntax.Node, p *syntax.Param) types.Type {
	return e.bindAtClass(e.paramTypeUnbound(scope, p), p)
}

// bindAtClass binds the relative atoms of a type written at n: `self` and
// `parent` name the class enclosing n (unknown in a trait, whose self is
// the using class, or outside a class); `static` (documented only) any
// declaredParam is the type p's declaration names, built once per parameter.
func (e *Env) declaredParam(p *syntax.Param) types.Type {
	if t, ok := e.declared[p]; ok {
		return t
	}
	if e.declared == nil {
		e.declared = map[*syntax.Param]types.Type{}
	}
	t := types.FromNode(p.Type, e.resolver(p.Span().Start))
	e.declared[p] = t
	return t
}

// class extending it, which that class bounds.
func (e *Env) bindAtClass(t types.Type, n syntax.Node) types.Type {
	if !t.HasRelative() {
		return t
	}
	var c *syntax.ClassLike
	if n != nil {
		c = syntax.EnclosingClass(n)
	}
	self, parent := e.selfClass(c), ""
	if self != "" {
		parent = e.Names.ParentFQN(c)
	}
	return t.BindRelative(self, parent, self)
}

// collected is the array a variadic parameter of element type t holds.
func collected(t types.Type) types.Type {
	atoms := make([]string, 0, len(t.Atoms()))
	for _, a := range t.Atoms() {
		atoms = append(atoms, a+"[]")
	}
	return types.Of(atoms...)
}

// variadicDoc is the array a variadic parameter holds, from its tag type t.
// A tag without an array type (`string ...$names`) types each collected
// value. An array or iterable tag (`string[] ...$names`) is ambiguous: the
// PHPStan and Psalm convention reads it as each value's type, while much
// code writes the collected array. Either guess is wrong for some code and
// would report its correct uses, so such a tag gives an unknown type.
func variadicDoc(t types.Type) types.Type {
	for _, a := range t.Atoms() {
		if types.IsArrayAtom(a) || a == "iterable" {
			return types.Unknown
		}
	}
	return collected(t) // unknown stays unknown
}

func (e *Env) paramTypeUnbound(scope syntax.Node, p *syntax.Param) types.Type {
	at := p.Span().Start
	declared := e.declaredParam(p)
	if p.Variadic && !declared.IsUnknown() {
		declared = collected(declared)
	}
	if lit, ok := p.Default.(*syntax.ConstFetch); ok && strings.EqualFold(lit.Name.Value, "null") && !declared.IsUnknown() {
		declared = types.Union(declared, types.Null)
	}
	doc := ""
	if d := e.DocOf(scope); d != nil && !e.native {
		for _, dp := range d.EffectiveParams() {
			if dp.Name == p.Var.Name {
				doc = dp.Type
			}
		}
	}
	if doc != "" {
		dt := e.docType(doc, e.resolverFor(scope, at))
		if p.Variadic {
			dt = variadicDoc(dt)
		}
		if narrowed, ok := e.narrowerDoc(declared, dt); ok {
			return narrowed
		}
		if declared.IsUnknown() || declared.HasAny("array", "iterable", "mixed") {
			return dt
		}
		return declared.WithTypeArgsFrom(dt)
	}
	if m, ok := scope.(*syntax.Method); ok && !e.native && (declared.IsUnknown() || declared.HasAny("array", "iterable", "mixed")) {
		if dt := e.inheritedParamDoc(m, nodeIndex(m.Params, p)); !dt.IsUnknown() && (!declared.HasAny("array", "iterable") || dt.Without("null").IsArrayLike()) {
			return dt
		}
	}
	if e.native || p.Variadic {
		return declared
	}
	// A callback's untyped parameter receives what its caller passes; one
	// declared `array` (or `iterable`, maybe nullable) keeps the passed
	// array's facts, with null only when it admits null.
	var ct types.Type
	switch fn := scope.(type) {
	case *syntax.Closure:
		if declared.IsUnknown() || declared.HasAny("array", "iterable") && declared.OnlyOf("array", "iterable", "null") {
			ct = e.contextParamType(fn, nodeIndex(fn.Params, p))
		}
	case *syntax.ArrowFunction:
		if declared.IsUnknown() || declared.HasAny("array", "iterable") && declared.OnlyOf("array", "iterable", "null") {
			ct = e.contextParamType(fn, nodeIndex(fn.Params, p))
		}
	}
	if declared.IsUnknown() || ct.Without("null").IsArrayLike() && (declared.Has("null") || !ct.Has("null")) {
		return ct
	}
	return declared
}

func (e *Env) inlineVarDocs(scope syntax.Node, addDoc func(string, uint32, uint32, func() types.Type)) {
	var span syntax.Span
	if scope == nil {
		span = syntax.Span{Start: 0, End: uint32(len(e.File.Src))}
	} else {
		span = scope.Span()
	}
	toks := e.File.Tokens
	first := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= span.Start })
	excluded := promotedHookDocSpans(scope)
	hook := 0
	for _, t := range toks[first:] {
		if t.Start >= span.End {
			break
		}
		for hook < len(excluded) && excluded[hook].End <= t.Start {
			hook++
		}
		if hook < len(excluded) && excluded[hook].Start <= t.Start {
			continue
		}
		if t.Kind != syntax.TDocComment && t.Kind != syntax.TComment {
			continue
		}
		text := string(e.File.Src[t.Start:t.End])
		if !strings.Contains(text, "@var") && !strings.Contains(text, "@phpstan-var") && !strings.Contains(text, "@psalm-var") {
			continue
		}
		d := phpdoc.Parse(text)
		for _, hint := range d.EffectiveVars() {
			typ, name := hint.Type, hint.Name
			if name == "" {
				continue
			}
			at := e.nodeAt(scope, t.Start)
			ty := e.bindAtClass(e.docType(typ, e.resolverFor(at, t.Start)), at)
			addDoc(name, t.Start, t.End, func() types.Type { return ty })
		}
	}
}

// semicolonBetween reports whether a `;` token lies in [from, to).
func (e *Env) semicolonBetween(from, to uint32) bool {
	toks := e.File.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= from })
	for ; i < len(toks) && toks[i].Start < to; i++ {
		if toks[i].Kind == syntax.TSemicolon {
			return true
		}
	}
	return false
}

// nodeAt returns scope itself (or nil for file scope); used to resolve doc
// names relative to the enclosing declarations.
func (e *Env) nodeAt(scope syntax.Node, _ uint32) syntax.Node { return scope }

// branchAssigned returns the variables assigned (plain `$x = …;` at the top
// level of the branch) in every branch of a complete if/elseif/else chain;
// branches that always leave (return/throw/…) do not constrain the result.
func branchAssigned(n *syntax.If) []string {
	if n.Else == nil {
		return nil
	}
	bodies := []syntax.Stmt{n.Body}
	for _, ei := range n.ElseIfs {
		bodies = append(bodies, ei.Body)
	}
	bodies = append(bodies, n.Else.Body)
	var common map[string]bool
	assignedSomewhere := false
	for _, b := range bodies {
		if terminates(b) {
			continue
		}
		set := topLevelAssigns(b)
		if common == nil {
			common = set
		} else {
			for k := range common {
				if !set[k] {
					delete(common, k)
				}
			}
		}
		assignedSomewhere = true
	}
	if !assignedSomewhere {
		return nil
	}
	out := make([]string, 0, len(common))
	for k := range common {
		out = append(out, k)
	}
	return out
}

func topLevelAssigns(s syntax.Stmt) map[string]bool {
	set := map[string]bool{}
	stmts := []syntax.Stmt{s}
	if b, ok := s.(*syntax.Block); ok {
		stmts = b.Stmts
	}
	for _, st := range stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		if a, ok := es.Expr.(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual {
			if v, ok := a.Var.(*syntax.Variable); ok && v.Name != "" {
				set[v.Name] = true
			}
		}
	}
	return set
}

// joinSpan is the region after an if-chain, up to the end of the enclosing
// statement list.
func joinSpan(n *syntax.If, f *syntax.File) syntax.Span {
	end := uint32(len(f.Src))
	if p := n.Parent(); p != nil {
		end = p.Span().End
	}
	return syntax.Span{Start: n.Span().End, End: end}
}
