// Package overrideattributewithouttarget implements the native OverrideAttributeWithoutTarget inspection.
package overrideattributewithouttarget

import (
	"slices"
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
)

const message = "Apply Override to an inherited member."

type rule struct{}

// New returns the OverrideAttributeWithoutTarget inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OverrideAttributeWithoutTarget" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAttribute} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Attribute)
	if ctx.PHP.Below(phpversion.PHP83) {
		return
	}
	name := ctx.Names().Class(a.Name.Value, a.Name.Span().Start)
	bc := ctx.Index().Class(name, ctx.PHP)
	if !strings.EqualFold(name, "Override") || bc == nil || !strings.HasPrefix(bc.File, "stubs/") {
		return
	}
	owner := a.Parent().Parent()
	decl := syntax.EnclosingClass(owner)
	// A trait member is validated by PHP in each using class, never in the trait.
	if decl == nil || decl.ClassKind == syntax.KindTrait {
		return
	}
	fqn := ctx.Names().DeclFQN(decl)
	c := ctx.Index().Class(fqn, ctx.PHP)
	if c == nil || !ctx.Index().AncestorsComplete(fqn, ctx.PHP) {
		return
	}
	var members []string
	domain := methods
	switch x := owner.(type) {
	case *syntax.Method:
		// Declaring __toString implicitly implements Stringable, which is the target.
		if strings.EqualFold(x.Name.Value, "__toString") {
			return
		}
		members = []string{x.Name.Value}
	case *syntax.Property:
		if ctx.PHP.Below(phpversion.PHP85) || len(x.Props) != 1 {
			return
		}
		members, domain = []string{x.Props[0].Var.Name}, properties
	case *syntax.ClassConst:
		if ctx.PHP.Below(phpversion.PHP86) {
			return
		}
		for _, item := range x.Consts {
			members = append(members, item.Name.Value)
		}
		domain = constants
	case *syntax.EnumCase:
		if ctx.PHP.Below(phpversion.PHP86) {
			return
		}
		members, domain = []string{x.Name.Value}, constants
	default:
		return
	}
	// Each member of a grouped constant declaration needs its own target.
	inherited := make([]bool, len(members))
	for _, anc := range ctx.Index().Ancestors(fqn, ctx.PHP) {
		if anc.Parent != "" && ctx.Index().Class(anc.Parent, ctx.PHP) == nil {
			return
		}
		for _, super := range anc.Interfaces {
			if ctx.Index().Class(super, ctx.PHP) == nil {
				return
			}
		}
		if strings.EqualFold(anc.FQN, fqn) {
			continue
		}
		if len(anc.Traits) > 0 {
			return
		}
		for i, member := range members {
			inherited[i] = inherited[i] || overridable(ctx, domain, anc.FQN, member)
		}
	}
	if len(c.Traits) > 0 || !slices.Contains(inherited, false) {
		return
	}
	ctx.ReportNode(a.Parent(), message)
}

// The separate lookup domains of an overridden member.
const (
	methods = iota
	properties
	constants // class constants and enum cases (PHP 8.6)
)

// overridable reports whether class declares or inherits a member of that
// name in that domain that a subclass may override: any but a private one.
func overridable(ctx *analysis.Context, domain int, class, member string) bool {
	vis := index.Private
	switch domain {
	case methods:
		if m := ctx.Index().FindMethod(class, member, ctx.PHP); m != nil {
			vis = m.Visibility
		}
	case properties:
		if p := ctx.Index().FindProperty(class, member, ctx.PHP); p != nil {
			vis = p.Visibility
		}
	default:
		if k := ctx.Index().FindConst(class, member, ctx.PHP); k != nil {
			vis = k.Visibility
		}
	}
	return vis != index.Private
}
