package propernullcoalescingoperatorusage

import (
	"sort"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// properNullCoalescingOperatorUsage reports `call() ?? null` and `??`
// operands with unrelated types.
type properNullCoalescingOperatorUsage struct{}

func (properNullCoalescingOperatorUsage) ID() string { return "ProperNullCoalescingOperatorUsage" }
func (properNullCoalescingOperatorUsage) Semantic()  {}
func (properNullCoalescingOperatorUsage) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary}
}

func pncoIsCoalesce(n syntax.Node) bool {
	b, ok := n.(*syntax.Binary)
	return ok && b.Op.Kind == syntax.TCoalesce
}

func (r properNullCoalescingOperatorUsage) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TCoalesce || b.Span().Len() == 0 || b.Left == nil || b.Right == nil || ctx.PHP < phpversion.PHP70 {
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
		switch x := b.Left.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			if m, ok := x.(*syntax.MethodCall); ok && m.NullSafe && !pncoDefined(m.Var, b) {
				return // `??` also silences the undefined-variable warning
			}
			left := ctx.Text(b.Left)
			span := b.Span()
			ctx.Report(span, "'"+left+"' alone is equivalent; drop the '?? null' fallback.", diagnostic.Fix{
				Title: "Drop the '?? null' fallback",
				Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: left}} },
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
	if complementary || pncoScalarOnly(lt) && pncoScalarOnly(rt) && !pncoNeverNull(b.Left) || r.related(ctx, lt, rt) || r.sameFamily(ctx, lt, rt) ||
		pncoCallableValue(ctx, b.Left, rt) || pncoCallableValue(ctx, b.Right, lt) {
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

// pncoCallableValue reports whether e is declared `callable` (not
// \Closure, which only an object satisfies) while the other side may be a
// callable string (`'trim'`) or a callable array (`[$this, 'm']`).
func pncoCallableValue(ctx *analysis.Context, e syntax.Expr, other map[string]bool) bool {
	return (other["string"] || other["array"]) && ctx.TypeOf(e).HasAny("callable")
}

// pncoDefined reports whether the receiver of a nullsafe call is certainly
// set at the `??` expression at: anything but a local variable (those warn
// with or without `??`), `$this`, a parameter, or a variable assigned on
// every path before at. A variable a closure imports with `use` is set
// inside it (to null when the outer one was not).
func pncoDefined(recv syntax.Expr, at syntax.Node) bool {
	v, ok := syntax.UnwrapParens(recv).(*syntax.Variable)
	if !ok || v.Name == "this" {
		return true
	}
	scope := syntax.EnclosingFuncLike(at)
	if c, ok := scope.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if v.Name != "" && u.Var != nil && u.Var.Name == v.Name {
				return true
			}
		}
	}
	for _, p := range syntax.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name == v.Name {
			return true
		}
	}
	_, entry := flowquery.ReachingAssignments(scope, at, v.Name)
	return !entry
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

// pncoNeverNull reports whether e is an operator expression whose value is
// never null (comparison, &&/||, instanceof, !): a scalar fallback after it
// is a precedence mistake, not a placeholder (D5a).
func pncoNeverNull(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TLess, syntax.TIsSmallerOrEqual, syntax.TGreater, syntax.TIsGreaterOrEqual, syntax.TSpaceship,
			syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
			syntax.TBooleanAnd, syntax.TBooleanOr:
			return true
		}
	case *syntax.Instanceof:
		return true
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim
	}
	return false
}

// sameFamily reports whether both sides are used through the same
// operation: both iterable (array, \Traversable or an implementor), an array
// next to a \Countable or \ArrayAccess implementor, or a string next to the
// bare \Stringable interface (D5b, D5c).
func (r properNullCoalescingOperatorUsage) sameFamily(ctx *analysis.Context, lt, rt map[string]bool) bool {
	if r.family(ctx, lt, "array", `\Traversable`) && r.family(ctx, rt, "array", `\Traversable`) {
		return true
	}
	for _, p := range [][2]map[string]bool{{lt, rt}, {rt, lt}} {
		if p[0]["array"] && r.family(ctx, p[1], "", `\Countable`, `\ArrayAccess`) || p[0]["string"] && r.family(ctx, p[1], `\Stringable`) {
			return true
		}
	}
	return false
}

// family reports whether the set holds the type exact (compared
// case-insensitively) or a subtype of one of the bases.
func (properNullCoalescingOperatorUsage) family(ctx *analysis.Context, set map[string]bool, exact string, bases ...string) bool {
	for t := range set {
		if strings.EqualFold(t, exact) {
			return true
		}
		for _, base := range bases {
			if strings.HasPrefix(t, `\`) && ctx.Index().IsSubtype(t, base, ctx.PHP) {
				return true
			}
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
