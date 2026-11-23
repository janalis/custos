// Package attributetargetmismatch implements the native AttributeTargetMismatch inspection.
package attributetargetmismatch

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Use the attribute on a permitted target."

type rule struct{}

// New returns the AttributeTargetMismatch inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AttributeTargetMismatch" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAttribute} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Attribute)
	if ctx.PHP.Below(phpversion.PHP80) {
		return
	}
	fqn, flags, known := semanticquery.NativeAttributeFlags(ctx, a)
	if !known {
		return
	}
	// DelayedTargetValidation (PHP 8.5) postpones the check of PHP's own
	// attributes to reflection, where it reports nothing.
	if _, builtin := semanticquery.BuiltinAttributeTargets(ctx, fqn); builtin && delayed(ctx, a) {
		return
	}
	var bit int64
	switch owner := a.Parent().Parent().(type) {
	case *syntax.ClassLike:
		bit = 1
		// PHP's own Deprecated applies to traits alone among class-likes.
		if _, builtin := semanticquery.BuiltinAttributeTargets(ctx, fqn); builtin && strings.EqualFold(strings.TrimPrefix(fqn, `\`), "Deprecated") && owner.ClassKind != syntax.KindTrait {
			bit = 0
		}
	case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction:
		bit = 2
	case *syntax.Method, *syntax.PropertyHook:
		bit = 4
	case *syntax.Property:
		bit = 8
	case *syntax.ClassConst, *syntax.EnumCase:
		bit = 16
	case *syntax.Param:
		bit = 32
		if promoted(owner) {
			bit |= 8
		}
	default: // the parser attaches attributes to no other node than a global const
		if ctx.PHP.Below(phpversion.PHP85) {
			return // a constant takes attributes from PHP 8.5: newer code, not a misplaced one
		}
		bit = 64
	}
	if flags&bit == 0 {
		ctx.ReportNode(a.Parent(), message)
	}
}

// promoted reports whether a parameter also declares a property: any
// visibility, set-visibility, readonly or final modifier promotes it.
func promoted(param *syntax.Param) bool {
	for _, m := range param.Modifiers {
		switch m.Kind {
		case syntax.TPublic, syntax.TProtected, syntax.TPrivate,
			syntax.TPublicSet, syntax.TProtectedSet, syntax.TPrivateSet,
			syntax.TReadonly, syntax.TFinal:
			return true
		}
	}
	return false
}

// delayed reports whether the declaration a applies to also carries PHP's
// DelayedTargetValidation attribute.
func delayed(ctx *analysis.Context, a *syntax.Attribute) bool {
	found := false
	syntax.Children(a.Parent().Parent(), func(n syntax.Node) {
		g, ok := n.(*syntax.AttributeGroup)
		if !ok {
			return
		}
		for _, other := range g.Attrs {
			name := ctx.Names().Class(other.Name.Value, other.Name.Span().Start)
			if strings.EqualFold(strings.TrimPrefix(name, `\`), "DelayedTargetValidation") {
				_, builtin := semanticquery.BuiltinAttributeTargets(ctx, name)
				found = found || builtin
			}
		}
	})
	return found
}
