package architecture

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// propertyCanBeStatic reports non-public instance properties defaulting to
// a sizeable array literal.
type propertyCanBeStatic struct{}

func init() { register(propertyCanBeStatic{}) }

func (propertyCanBeStatic) ID() string { return "PropertyCanBeStatic" }

func (propertyCanBeStatic) Semantic() {}

func (propertyCanBeStatic) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KProperty} }

func (propertyCanBeStatic) Check(ctx *analysis.Context, n syntax.Node) {
	prop := n.(*syntax.Property)
	if prop.Modifiers.Has(syntax.TStatic) || !(prop.Modifiers.Has(syntax.TPrivate) || prop.Modifiers.Has(syntax.TProtected)) { // D2
		return
	}
	cl, ok := prop.Parent().(*syntax.ClassLike)
	if !ok {
		return
	}
	parent := ""
	if p := ctx.Names().ParentFQN(cl); p != "" {
		if pc := ctx.Index().Class(p, ctx.PHP); pc != nil {
			parent = pc.FQN
		}
	}
	for _, item := range prop.Props { // D1
		arr, ok := item.Default.(*syntax.Array) // D3
		if !ok || item.Var == nil || item.Var.Span().Len() == 0 {
			continue
		}
		if parent != "" && util.PropertyInChain(ctx.Index(), parent, item.Var.Name, ctx.PHP) != nil { // D4
			continue
		}
		count := 0
		for _, el := range arr.Items { // D5
			if el == nil || el.Unpack || el.Value == nil {
				continue
			}
			switch v := syntax.UnwrapParens(el.Value).(type) {
			case *syntax.Array:
				count++
			case *syntax.InterpolatedString:
				if !v.Backtick {
					count++
				}
			case *syntax.Literal:
				if v.LitKind == syntax.LitString {
					count++
				}
			}
			if count >= 3 {
				break
			}
		}
		if count < 3 {
			continue
		}
		msg := "Large array default on an instance property; consider a static property."
		if ctx.PHP >= phpver.PHP56 {
			msg = "Large array default on an instance property; consider a static property or a class constant."
		}
		ctx.ReportNode(item.Var, msg)
	}
}
