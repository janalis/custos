// Package undefinedclass implements the native UndefinedClass inspection.
package undefinedclass

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/names"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "UndefinedClass" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{
		syntax.KNew, syntax.KStaticCall, syntax.KClassConstFetch, syntax.KStaticPropertyFetch,
		syntax.KClassLike, syntax.KTraitUse,
	}
}

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var refs []syntax.Expr
	switch n := n.(type) {
	case *syntax.New:
		refs = append(refs, n.Class)
	case *syntax.StaticCall:
		refs = append(refs, n.Class)
	case *syntax.StaticPropertyFetch:
		refs = append(refs, n.Class)
	case *syntax.ClassConstFetch:
		// `Name::class` only spells the name; it loads nothing.
		if id, ok := n.Name.(*syntax.Identifier); !ok || !strings.EqualFold(id.Value, "class") {
			refs = append(refs, n.Class)
		}
	case *syntax.ClassLike:
		for _, name := range n.Extends {
			refs = append(refs, name)
		}
		for _, name := range n.Implements {
			refs = append(refs, name)
		}
	case *syntax.TraitUse:
		for _, name := range n.Traits {
			refs = append(refs, name)
		}
	}
	for _, ref := range refs {
		name, ok := ref.(*syntax.Name)
		if !ok || names.IsSpecialClass(name.Value) {
			continue
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		if !semanticquery.UndeclaredEverywhere(ctx, false, fqn) || semanticquery.ExistenceChecked(ctx, name, fqn) {
			continue
		}
		ctx.ReportNode(name, "Fix the name or install the package declaring "+fqn+".")
	}
}
