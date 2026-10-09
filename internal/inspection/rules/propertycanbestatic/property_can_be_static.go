package propertycanbestatic

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// propertyCanBeStatic reports non-public instance properties defaulting to
// a sizeable array literal.
type propertyCanBeStatic struct{}

func (propertyCanBeStatic) ID() string               { return "PropertyCanBeStatic" }
func (propertyCanBeStatic) Semantic()                {}
func (propertyCanBeStatic) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KProperty} }
func (propertyCanBeStatic) Check(ctx *analysis.Context, n syntax.Node) {
	prop := n.(*syntax.Property)
	if prop.Modifiers.Has(syntax.TStatic) || !(prop.Modifiers.Has(syntax.TPrivate) || prop.Modifiers.Has(syntax.TProtected)) { // D2
		return
	}
	cl := prop.Parent().(*syntax.ClassLike) // properties only appear in class-like bodies
	parent := ""
	if p := ctx.Names().ParentFQN(cl); p != "" {
		if pc := ctx.Index().Class(p, ctx.PHP); pc != nil {
			parent = pc.FQN
		}
	}
	for _, item := range prop.Props { // D1
		arr, ok := item.Default.(*syntax.Array) // D3
		if !ok || item.Var.Span().Len() == 0 {
			continue
		}
		if parent != "" && semanticquery.PropertyInChain(ctx.Index(), parent, item.Var.Name, ctx.PHP) != nil { // D4
			continue
		}
		if pcbsWritten(ctx, cl)[item.Var.Name] { // custos: per-instance state
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
		if ctx.PHP >= phpversion.PHP56 {
			msg = "Large array default on an instance property; consider a static property or a class constant."
		}
		ctx.ReportNode(item.Var, msg)
	}
}

// pcbsWritten returns the names of the properties the class body writes
// through `$this` (`$this->p = …`, `$this->p['k'] = …`, `$this->p[] = …`,
// compound assignments, `++`/`--`, `unset($this->p['k'])`): such a
// property holds per-instance state (Roundcube's rcube_db options, set
// per connection), and a static one would be shared by every instance
// (custos).
func pcbsWritten(ctx *analysis.Context, cl *syntax.ClassLike) map[string]bool {
	return ctx.Memo("written:"+strconv.Itoa(int(cl.Span().Start)), func() any {
		out := map[string]bool{}
		mark := func(e syntax.Expr) {
			for {
				switch x := syntax.UnwrapParens(e).(type) {
				case *syntax.ArrayDimFetch:
					e = x.Var
					continue
				case *syntax.PropertyFetch:
					v, isVar := syntax.UnwrapParens(x.Var).(*syntax.Variable)
					if id, ok := x.Name.(*syntax.Identifier); ok && isVar && v.Name == "this" {
						out[id.Value] = true
					}
				}
				return
			}
		}
		syntax.Inspect(cl, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Assign:
				mark(x.Var)
			case *syntax.IncDec:
				mark(x.Var)
			case *syntax.Unset:
				for _, v := range x.Vars {
					mark(v)
				}
			}
			return true
		})
		return out
	}).(map[string]bool)
}
