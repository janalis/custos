package codestyle

import (
	"sort"
	"strings"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// properNullCoalescingOperatorUsage reports `call() ?? null` and `??`
// operands with unrelated types.
type properNullCoalescingOperatorUsage struct{}

func init() { register(properNullCoalescingOperatorUsage{}) }

func (properNullCoalescingOperatorUsage) ID() string { return "ProperNullCoalescingOperatorUsage" }

func (properNullCoalescingOperatorUsage) Semantic() {}

func (properNullCoalescingOperatorUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary}
}

func pncoIsCoalesce(n syntax.Node) bool {
	b, ok := n.(*syntax.Binary)
	return ok && b.Op.Kind == syntax.TCoalesce
}

func (r properNullCoalescingOperatorUsage) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TCoalesce || b.Span().Len() == 0 || b.Left == nil || b.Right == nil || ctx.PHP < phpver.PHP70 {
		return
	}
	if pncoIsCoalesce(b.Parent()) { // P1
		return
	}
	if p, ok := b.Parent().(*syntax.Paren); ok { // P2
		if u, ok := p.Parent().(*syntax.Unary); ok && u.Op.Kind.IsCast() {
			return
		}
	}
	if syntax.IsNullConst(b.Right) { // Case A
		switch b.Left.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			left := ctx.Text(b.Left)
			span := b.Span()
			ctx.Report(span, "'"+left+"' alone is equivalent; drop the '?? null' fallback.", analysis.Fix{
				Title: "Drop the '?? null' fallback",
				Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: left}} },
			})
		}
		return
	}
	if !ctx.Bool("ANALYZE_TYPES") || syntax.EnclosingFuncLike(b) == nil { // D2
		return
	}
	lt, ok := r.typeSet(ctx, b.Left)
	if !ok {
		return
	}
	rt, ok := r.typeSet(ctx, b.Right)
	if !ok {
		return
	}
	complementary := false // D4
	if ctx.Bool("ALLOW_OVERLAPPING_TYPES") {
		for t := range rt {
			if lt[t] {
				complementary = true
				break
			}
		}
	} else {
		complementary = true
		for t := range lt {
			if !rt[t] {
				complementary = false
				break
			}
		}
	}
	if complementary || pncoScalarOnly(lt) && pncoScalarOnly(rt) || r.related(ctx, lt, rt) || r.iterable(ctx, lt) && r.iterable(ctx, rt) {
		return
	}
	ctx.Report(b.Span(), "Operand types of '??' do not match ("+pncoSetString(lt)+" vs "+pncoSetString(rt)+").")
}

// typeSet returns the normalised types of e without null/static (D3).
func (properNullCoalescingOperatorUsage) typeSet(ctx *analysis.Context, e syntax.Expr) (map[string]bool, bool) {
	if v, ok := syntax.UnwrapParens(e).(*syntax.Variable); ok && v.Name == "this" {
		return nil, false // `$this` is static
	}
	t := ctx.TypeOf(e)
	if t.IsUnknown() || t.HasAny("mixed", "object") {
		return nil, false
	}
	out := map[string]bool{}
	for _, a := range t.Atoms() {
		switch {
		case a == "null" || a == "static" || a == "never":
			// never (`?? throw …`) is the bottom type: it fits any side.
			continue
		case strings.EqualFold(a, "self") || strings.EqualFold(a, "parent"):
			// Bind self/parent to the enclosing class (or its parent).
			fqn := ctx.Types().ClassFQN(syntax.EnclosingClass(e))
			if fqn != "" && strings.EqualFold(a, "parent") {
				c := ctx.Index().Class(fqn, ctx.PHP)
				fqn = ""
				if c != nil {
					fqn = c.Parent
				}
			}
			if fqn == "" {
				return nil, false
			}
			a = `\` + strings.TrimPrefix(fqn, `\`)
		case a == "true" || a == "false":
			a = "bool"
		case strings.HasSuffix(a, "[]"):
			a = "array"
		case strings.EqualFold(a, `\Closure`):
			a = "callable"
		case a == "iterable":
			// iterable is array|\Traversable: an array fallback fits it.
			out["array"] = true
			out[`\Traversable`] = true
			continue
		}
		if strings.HasPrefix(a, `\`) && ctx.Index().Class(a, ctx.PHP) == nil {
			return nil, false // unresolved class: unusable set (D3)
		}
		out[a] = true
	}
	return out, len(out) > 0
}

// related reports whether the class types of both sides share an element of
// their inheritance closures (D5).
func (properNullCoalescingOperatorUsage) related(ctx *analysis.Context, lt, rt map[string]bool) bool {
	closure := func(set map[string]bool) map[string]bool {
		out := map[string]bool{}
		for t := range set {
			if !strings.HasPrefix(t, `\`) {
				continue
			}
			out[strings.ToLower(strings.TrimPrefix(t, `\`))] = true
			for _, c := range ctx.Index().Ancestors(t, ctx.PHP) {
				out[strings.ToLower(c.FQN)] = true
			}
		}
		return out
	}
	lc := closure(lt)
	for k := range closure(rt) {
		if lc[k] {
			return true
		}
	}
	// An invokable class fits a callable side (custos).
	invokable := func(set map[string]bool) bool {
		for t := range set {
			if strings.HasPrefix(t, `\`) && ctx.Index().FindMethod(t, "__invoke", ctx.PHP) != nil {
				return true
			}
		}
		return false
	}
	return lt["callable"] && invokable(rt) || rt["callable"] && invokable(lt)
}

// pncoScalarOnly reports whether every type of the set is a scalar (D5a).
func pncoScalarOnly(set map[string]bool) bool {
	for t := range set {
		switch t {
		case "int", "float", "string", "bool":
		default:
			return false
		}
	}
	return true
}

// iterable reports whether the set holds a member of the iterable family:
// array, \Traversable or one of its implementors (D5b).
func (properNullCoalescingOperatorUsage) iterable(ctx *analysis.Context, set map[string]bool) bool {
	for t := range set {
		if t == "array" || strings.HasPrefix(t, `\`) && ctx.Index().IsSubtype(t, `\Traversable`, ctx.PHP) {
			return true
		}
	}
	return false
}

func pncoSetString(s map[string]bool) string {
	list := make([]string, 0, len(s))
	for k := range s {
		list = append(list, k)
	}
	sort.Strings(list)
	return "[" + strings.Join(list, ", ") + "]"
}
