package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// IndexRelations answers the class questions of types.Accepts from the
// project index. Not safe for concurrent use (it caches resolutions).
type IndexRelations struct {
	ix       *index.Index
	ver      phpversion.Version
	resolved map[string]bool
}

// NewIndexRelations returns the class relations of ix at PHP version ver.
func NewIndexRelations(ix *index.Index, ver phpversion.Version) *IndexRelations {
	return &IndexRelations{ix: ix, ver: ver}
}

// AcceptContext returns the acceptance context of a value passed in the file
// of ctx: its strict_types mode and the class relations of its index;
// internal marks a parameter of a builtin function.
func AcceptContext(ctx *analysis.Context, internal bool) types.AcceptContext {
	return types.AcceptContext{
		Classes:  NewIndexRelations(ctx.Index(), ctx.PHP),
		Strict:   syntax.StrictTypes(ctx.File),
		Internal: internal,
	}
}

// class returns the class-like named by class atom a (nil when unknown).
func (r *IndexRelations) class(a string) *index.Class {
	return r.ix.Class(strings.TrimPrefix(a, `\`), r.ver)
}

// Resolved reports a class known together with its whole hierarchy.
func (r *IndexRelations) Resolved(a string) bool {
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
func (r *IndexRelations) IsSubtype(child, parent string) types.Verdict {
	if !r.Resolved(child) || !r.Resolved(parent) {
		return types.Maybe
	}
	if r.ix.IsSubtype(strings.TrimPrefix(child, `\`), strings.TrimPrefix(parent, `\`), r.ver) {
		return types.Yes
	}
	return types.No
}

// IsFinal reports a final class or an enum.
func (r *IndexRelations) IsFinal(a string) bool {
	c := r.class(a)
	return c != nil && (c.Final || c.Kind == syntax.KindEnum)
}

// IsInterface reports an interface.
func (r *IndexRelations) IsInterface(a string) bool {
	c := r.class(a)
	return c != nil && c.Kind == syntax.KindInterface
}
