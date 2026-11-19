package semanticquery

import (
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
// arguments of its classes only when x is a parameter nothing has written:
// those are the ones its declaration states. Elsewhere inference may have
// filled them from a template's bound, which says what they could be and
// not what they are, so they cannot prove a mismatch.
func declaredTypeArgs(ctx *analysis.Context, x syntax.Expr, given types.Type) types.Type {
	if variable, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok {
		scope := syntax.EnclosingVariableScope(x)
		params, _ := astquery.ScopeParts(scope)
		for _, p := range params {
			if p.Var != nil && p.Var.Name == variable.Name &&
				nativeContractScope(ctx, scope).localWrites(ctx, scope).untouched(variable.Name, x.Span().Start) {
				return given
			}
		}
	}
	return given.WithoutClassTypeArgs()
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
