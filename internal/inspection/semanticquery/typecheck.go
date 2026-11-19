package semanticquery

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// indexRelations answers the class questions of types.Accepts from the
// project index. Not safe for concurrent use (it caches resolutions).
type indexRelations struct {
	ix       *index.Index
	ver      phpversion.Version
	resolved map[string]bool
}

// newIndexRelations returns the class relations of ix at PHP version ver.
func newIndexRelations(ix *index.Index, ver phpversion.Version) *indexRelations {
	return &indexRelations{ix: ix, ver: ver}
}

// acceptContext returns the acceptance context of a value passed in the file
// of ctx: its strict_types mode and the class relations of its index;
// internal marks a parameter of a builtin function.
func acceptContext(ctx *analysis.Context, internal bool) types.AcceptContext {
	return types.AcceptContext{
		Classes:  newIndexRelations(ctx.Index(), ctx.PHP),
		Strict:   syntax.StrictTypes(ctx.File),
		Internal: internal,
	}
}

// declaredTypeArgs returns given, the type of value x, keeping the generic
// arguments of its classes only when a declaration states them for x. Elsewhere
// inference may have filled them from a template's bound, which says what they
// could be and not what they are, so they cannot prove a mismatch.
func declaredTypeArgs(ctx *analysis.Context, x syntax.Expr, given types.Type) types.Type {
	if typeArgsDeclared(ctx, x, 0) {
		return given
	}
	return given.WithoutClassTypeArgs()
}

// maxTypeArgSources bounds the assignments followed back from one value.
const maxTypeArgSources = 4

// typeArgsDeclared reports a value whose generic arguments come from a
// declaration: a parameter nothing has written, a property of $this, a call
// of a function or method whose documented return mentions no template, or a
// local whose only assignment is one of those.
func typeArgsDeclared(ctx *analysis.Context, x syntax.Expr, depth int) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		return localTypeArgsDeclared(ctx, n, depth)
	case *syntax.PropertyFetch:
		return ownPropertyTypeArgs(ctx, n)
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(n)
		return f != nil && f.Tpl == nil && plainDocReturn(f.DocReturn)
	case *syntax.MethodCall:
		id, named := n.Name.(*syntax.Identifier)
		classes := ctx.TypeOf(n.Var).Classes()
		if !named || len(classes) == 0 {
			return false
		}
		for _, class := range classes {
			if !plainMethodReturn(ctx.Index().FindMethod(strings.TrimPrefix(class, `\`), id.Value, ctx.PHP)) {
				return false
			}
		}
		return true
	case *syntax.StaticCall:
		id, named := n.Name.(*syntax.Identifier)
		name, written := n.Class.(*syntax.Name)
		if !named || !written {
			return false
		}
		class := StaticCallClass(ctx, name)
		return class != "" && plainMethodReturn(ctx.Index().FindMethod(class, id.Value, ctx.PHP))
	}
	return false
}

// localTypeArgsDeclared is typeArgsDeclared for a local variable.
func localTypeArgsDeclared(ctx *analysis.Context, v *syntax.Variable, depth int) bool {
	scope := syntax.EnclosingVariableScope(v)
	params, _ := astquery.ScopeParts(scope)
	untouched := nativeContractScope(ctx, scope).localWrites(ctx, scope).untouched(v.Name, v.Span().Start)
	for _, p := range params {
		if p.Var != nil && p.Var.Name == v.Name && untouched {
			return true
		}
	}
	origins, single := ctx.Flow().Reaching(v)
	return single && depth < maxTypeArgSources && typeArgsDeclared(ctx, origins[0].Expr, depth+1)
}

// ownPropertyTypeArgs reports a property of $this documented without templates.
func ownPropertyTypeArgs(ctx *analysis.Context, n *syntax.PropertyFetch) bool {
	this, isVar := n.Var.(*syntax.Variable)
	id, named := n.Name.(*syntax.Identifier)
	class := syntax.EnclosingClass(n)
	if !isVar || this.Name != "this" || !named || class == nil {
		return false
	}
	p := ctx.Index().FindProperty(strings.TrimPrefix(ctx.Types().ClassFQN(class), `\`), id.Value, ctx.PHP)
	return p != nil && plainDocReturn(p.DocType)
}

// plainMethodReturn reports a method whose documented return mentions no
// template of its own or of its class.
func plainMethodReturn(m *index.Method) bool {
	return m != nil && m.Tpl == nil && m.GenReturn == "" && plainDocReturn(m.DocReturn)
}

// plainDocReturn reports a documented type that mentions no template.
func plainDocReturn(doc string) bool {
	return doc != "" && !strings.Contains(doc, `\~`)
}

// Limits on one walk up a hierarchy, which analyzed code controls: each
// @extends tag may repeat a template, doubling the text of its arguments.
const (
	maxTypeArgClasses = 64
	maxTypeArgWork    = 16 * types.MaxDocTypeLen // bytes of arguments rendered
)

// TypeArgsAs returns the generic arguments class a, instantiated with args,
// gives its ancestor through @extends, @implements and @use tags; nil when it
// gives none or the walk runs out of budget. A template of a that args does
// not bind, or binds to an overlong type, reads as mixed.
func (r *indexRelations) TypeArgsAs(a string, args []types.Type, ancestor string) []types.Type {
	seen := map[string]bool{}
	work := maxTypeArgWork
	var walk func(fqn string, args []types.Type) []types.Type
	walk = func(fqn string, args []types.Type) []types.Type {
		k := strings.ToLower(strings.TrimPrefix(fqn, `\`))
		if strings.EqualFold(k, strings.TrimPrefix(ancestor, `\`)) {
			return args
		}
		c := r.class(fqn)
		if c == nil || seen[k] || len(seen) >= maxTypeArgClasses {
			return nil
		}
		seen[k] = true
		bound := boundTemplates(c, args, &work)
		for _, s := range c.Supers {
			if work < 0 {
				return nil
			}
			given := make([]types.Type, len(s.Args))
			for i, text := range s.Args {
				given[i] = types.FromDoc(text, bound)
			}
			if found := walk(s.Class, given); found != nil {
				return found
			}
		}
		for _, next := range slices.Concat([]string{c.Parent}, c.Interfaces, c.Traits) {
			if next == "" {
				continue
			}
			if found := walk(next, nil); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(a, args)
}

// boundTemplates resolves the templates of c in a doc type to args, by
// position, each rendered once and charged to work; a template without a
// usable argument reads as mixed.
func boundTemplates(c *index.Class, args []types.Type, work *int) types.Resolver {
	rendered := make(map[string]string, len(args))
	for i, tp := range c.Templates {
		if i >= len(args) || args[i].IsUnknown() {
			continue
		}
		text := args[i].DocString()
		*work -= len(text)
		if len(text) <= types.MaxDocTypeLen {
			rendered[tp.Name] = "=" + text
		}
	}
	return func(written string) string {
		name := strings.TrimPrefix(written, `\`)
		if template, isTemplate := strings.CutPrefix(name, "~"); isTemplate {
			return rendered[template]
		}
		return name
	}
}

// docConstants resolves the class constants a doc type names to their values.
type docConstants struct {
	ix     *index.Index
	ver    phpversion.Version
	parsed map[string]types.Type
}

// typeOf is doc type text with its class constants read as their values,
// nullable when asked; plain when one of them has no value known from its
// declaration.
func (d docConstants) typeOf(text string, plain types.Type, nullable bool) types.Type {
	t, ok := d.parsed[text]
	if !ok {
		known := true
		t = types.FromDocResolved(text, nil, func(expr string) types.Type {
			v := d.constant(expr)
			known = known && !v.IsUnknown()
			return v
		})
		if !known {
			t = types.Unknown
		}
		d.parsed[text] = t
	}
	if t.IsUnknown() {
		return plain
	}
	if nullable {
		return types.Union(t, types.Of("null"))
	}
	return t
}

// constant is the type of `\Foo::BAR` or `\Foo::*`; unknown unless every
// constant it names is declared with a literal value. Only a class without
// ancestors lists all its constants for `*`.
func (d docConstants) constant(expr string) types.Type {
	owner, member, _ := strings.Cut(strings.TrimPrefix(expr, `\`), "::")
	if owner == "" {
		return d.globalConstant(member)
	}
	if member != "*" {
		if k := d.ix.FindConst(owner, member, d.ver); k != nil {
			return literalValue(k.Value)
		}
		return types.Unknown
	}
	c := d.ix.Class(owner, d.ver)
	if c == nil || c.Parent != "" || len(c.Interfaces) > 0 || len(c.Traits) > 0 || len(c.Consts) == 0 {
		return types.Unknown
	}
	values := make([]types.Type, 0, len(c.Consts))
	for _, k := range c.Consts {
		v := literalValue(k.Value)
		if v.IsUnknown() {
			return types.Unknown
		}
		values = append(values, v)
	}
	return types.Union(values...)
}

// globalConstant is the type of a global constant written unqualified: the
// one of its namespace, else the global one.
func (d docConstants) globalConstant(fqn string) types.Type {
	k := d.ix.Constant(fqn, d.ver)
	if k == nil {
		k = d.ix.Constant(fqn[strings.LastIndexByte(fqn, '\\')+1:], d.ver)
	}
	if k == nil {
		return types.Unknown
	}
	return literalValue(k.Value)
}

// literalValue is the exact type of int or plain string literal source text.
func literalValue(text string) types.Type {
	text = strings.TrimSpace(text)
	if v, err := strconv.ParseInt(text, 0, 64); err == nil {
		return types.IntValue(v)
	}
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') && text[len(text)-1] == text[0] && !strings.ContainsAny(text[1:len(text)-1], `\$'"`) {
		return types.StringLiteral(text[1 : len(text)-1])
	}
	return types.Unknown
}

// class returns the class-like named by class atom a (nil when unknown).
func (r *indexRelations) class(a string) *index.Class {
	return r.ix.Class(strings.TrimPrefix(a, `\`), r.ver)
}

// Resolved reports a class known together with its whole hierarchy.
func (r *indexRelations) Resolved(a string) bool {
	if v, ok := r.resolved[a]; ok {
		return v
	}
	v := r.class(a) != nil && HierarchyResolved(r.ix, strings.TrimPrefix(a, `\`), r.ver)
	if r.resolved == nil {
		r.resolved = map[string]bool{}
	}
	r.resolved[a] = v
	return v
}

// IsSubtype reports whether child is, extends or implements parent; Maybe
// when either is not resolved.
func (r *indexRelations) IsSubtype(child, parent string) types.Verdict {
	if !r.Resolved(child) || !r.Resolved(parent) {
		return types.Maybe
	}
	if r.ix.IsSubtype(strings.TrimPrefix(child, `\`), strings.TrimPrefix(parent, `\`), r.ver) {
		return types.Yes
	}
	return types.No
}

// IsFinal reports a final class or an enum.
func (r *indexRelations) IsFinal(a string) bool {
	c := r.class(a)
	return c != nil && (c.Final || c.Kind == syntax.KindEnum)
}

// IsInterface reports an interface.
func (r *indexRelations) IsInterface(a string) bool {
	c := r.class(a)
	return c != nil && c.Kind == syntax.KindInterface
}

// TemplateVariance returns the variance of the template of class a at
// position i; false when a is unknown or declares none there.
func (r *indexRelations) TemplateVariance(a string, i int) (types.Variance, bool) {
	c := r.class(a)
	if c == nil || i >= len(c.Templates) {
		return types.Invariant, false
	}
	switch c.Templates[i].Variance {
	case phpdoc.Covariant:
		return types.Covariant, true
	case phpdoc.Contravariant:
		return types.Contravariant, true
	}
	return types.Invariant, true
}
