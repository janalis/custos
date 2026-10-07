package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// nullCoalescingOperatorCanBeUsed reports ternaries and small if constructs
// that can be written with `??`.
type nullCoalescingOperatorCanBeUsed struct{}

func init() { register(nullCoalescingOperatorCanBeUsed{}) }

func (nullCoalescingOperatorCanBeUsed) ID() string { return "NullCoalescingOperatorCanBeUsed" }

func (nullCoalescingOperatorCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KTernary, syntax.KIf}
}

func (r nullCoalescingOperatorCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP70 { // E1
		return
	}
	switch n := n.(type) {
	case *syntax.Ternary:
		if ctx.Bool("SUGGEST_SIMPLIFYING_TERNARIES") {
			r.checkTernary(ctx, n)
		}
	case *syntax.If:
		if ctx.Bool("SUGGEST_SIMPLIFYING_IFS") {
			r.checkIf(ctx, n)
		}
	}
}

// ncoCond is a classified condition.
type ncoCond struct {
	kind    byte // 't' truthy, 'i' isset, 'e' empty, 'n' null comparison, 'a' array_key_exists
	negated bool
	k       syntax.Expr // the subject condition (K)
	subject syntax.Expr // S for isset/empty, the compared value for K-null
	op      syntax.TokenKind
	call    *syntax.FuncCall
}

func ncoClassify(ctx *analysis.Context, c0 syntax.Expr) (ncoCond, bool) {
	var c ncoCond
	k := syntax.UnwrapParens(c0)
	if u, ok := k.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		c.negated = true
		k = syntax.UnwrapParens(u.Expr)
	}
	c.k = k
	switch x := k.(type) {
	case *syntax.Variable, *syntax.ArrayDimFetch, *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		c.kind = 't'
	case *syntax.Isset:
		if len(x.Vars) != 1 {
			return c, false
		}
		c.kind, c.subject = 'i', x.Vars[0]
	case *syntax.Empty:
		c.kind, c.subject = 'e', x.Expr
	case *syntax.Binary:
		if x.Op.Kind != syntax.TIsIdentical && x.Op.Kind != syntax.TIsNotIdentical {
			return c, false
		}
		switch {
		case syntax.IsNullConst(x.Left):
			c.subject = x.Right
		case syntax.IsNullConst(x.Right):
			c.subject = x.Left
		default:
			return c, false
		}
		c.kind, c.op = 'n', x.Op.Kind
	case *syntax.FuncCall:
		if ctx.GlobalFunctionName(x) != "array_key_exists" { // any case, global function only
			return c, false
		}
		c.kind, c.call = 'a', x
	default:
		return c, false
	}
	return c, true
}

// ncoWrap parenthesises low-precedence expressions.
func ncoWrap(ctx *analysis.Context, e syntax.Expr) string {
	switch x := e.(type) {
	case *syntax.Ternary, *syntax.Assign, *syntax.Instanceof:
		return "(" + ctx.Text(e) + ")"
	case *syntax.Binary:
		if x.Op.Kind != syntax.TCoalesce {
			return "(" + ctx.Text(e) + ")"
		}
	}
	return ctx.Text(e)
}

func ncoAlt(ctx *analysis.Context, alt syntax.Expr) string {
	if alt == nil {
		return "null"
	}
	return ncoWrap(ctx, alt)
}

// ncoPropBase returns the base of a property access (static when allowed).
func ncoPropBase(e syntax.Expr, allowStatic bool) (syntax.Expr, bool) {
	switch x := e.(type) {
	case *syntax.PropertyFetch:
		return x.Var, true
	case *syntax.StaticPropertyFetch:
		if allowStatic {
			return x.Class, true
		}
	}
	return nil, false
}

// ncoSafeFallback implements G6. With a null (or absent) fallback the
// rewrite yields the same value in every case. With any other fallback it is
// only equivalent when the probe holds objects (or null, when allowNull) —
// a truthy scalar or array probe made the original read a property on a
// non-object and yield null — and when the property can never be null, since
// a set object with a null property yielded null, not the fallback.
func ncoSafeFallback(ctx *analysis.Context, probe syntax.Expr, allowNull bool, cand, alt syntax.Expr) bool {
	if alt == nil || syntax.IsNullConst(alt) {
		return true
	}
	if pt := ctx.TypeOf(cand); pt.IsUnknown() || pt.Has("null") {
		return false
	}
	return ncoObjectsOnly(ctx, probe, allowNull)
}

// ncoObjectsOnly reports whether probe's type is known and every component
// is a class type, `object` or `static` (or null, when allowNull).
func ncoObjectsOnly(ctx *analysis.Context, probe syntax.Expr, allowNull bool) bool {
	t := ctx.TypeOf(probe)
	if t.IsUnknown() {
		return false
	}
	for _, a := range t.Atoms() {
		switch {
		case a == "null":
			if !allowNull {
				return false
			}
		case a == "object", a == "static":
		case strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]"):
		default:
			return false
		}
	}
	return true
}

// ncoGenerate implements G1–G5. f may be nil (absent).
func ncoGenerate(ctx *analysis.Context, c ncoCond, t, f syntax.Expr) (string, bool) {
	eq := func(a, b syntax.Expr) bool { return a != nil && b != nil && util.EquivalentFoldNames(ctx.File, a, b) }
	pick := func(trueIsCand bool) (cand, alt syntax.Expr) {
		if trueIsCand {
			return t, f
		}
		return f, t
	}
	switch c.kind {
	case 'i': // G1
		cand, alt := pick(!c.negated)
		if cand == nil || !eq(cand, c.subject) {
			return "", false
		}
		return ncoWrap(ctx, cand) + " ?? " + ncoAlt(ctx, alt), true
	case 'e': // G2
		cand, alt := pick(c.negated)
		if cand == nil {
			return "", false
		}
		base, ok := ncoPropBase(cand, true)
		if !ok || !eq(base, c.subject) {
			return "", false
		}
		_, static := cand.(*syntax.StaticPropertyFetch)
		// `$c::$p ?? x` throws when $c is '' or null (?? does not guard the
		// class lookup), where `!empty($c)` skipped it: objects only (spec
		// Divergences).
		if static && !ncoObjectsOnly(ctx, c.subject, false) {
			return "", false
		}
		if !ncoSafeFallback(ctx, c.subject, !static, cand, alt) { // G6
			return "", false
		}
		return ctx.Text(cand) + " ?? " + ncoAlt(ctx, alt), true
	case 't': // G3
		cand, alt := pick(!c.negated)
		if cand == nil {
			return "", false
		}
		base, ok := ncoPropBase(cand, false)
		if !ok || !eq(base, c.k) {
			return "", false
		}
		if ctx.TypeOf(cand).IsUnknown() || !ncoSafeFallback(ctx, c.k, true, cand, alt) { // G6
			return "", false
		}
		return ctx.Text(cand) + " ?? " + ncoAlt(ctx, alt), true
	case 'a': // G4
		if f == nil {
			return "", false
		}
		args, ok := util.CallArgValues(c.call)
		if !ok || len(args) != 2 {
			return "", false
		}
		cand, alt := pick(!c.negated)
		dim, ok := cand.(*syntax.ArrayDimFetch)
		if !ok || dim.Dim == nil || !eq(dim.Var, args[1]) || !eq(dim.Dim, args[0]) || !syntax.IsNullConst(alt) {
			return "", false
		}
		return ncoWrap(ctx, cand) + " ?? " + ncoWrap(ctx, alt), true
	default: // 'n', G5
		if f == nil {
			return "", false
		}
		set := (c.op == syntax.TIsNotIdentical) != c.negated
		cand, alt := pick(set)
		if !eq(cand, c.subject) {
			return "", false
		}
		return ncoWrap(ctx, cand) + " ?? " + ncoWrap(ctx, alt), true
	}
}

func (nullCoalescingOperatorCanBeUsed) checkTernary(ctx *analysis.Context, t *syntax.Ternary) {
	if t.Then == nil { // E2
		return
	}
	c, ok := ncoClassify(ctx, t.Cond)
	if !ok {
		return
	}
	r, ok := ncoGenerate(ctx, c, t.Then, t.Else)
	if !ok {
		return
	}
	span := t.Span()
	ctx.Report(span, "Simplify to '"+r+"' using the null coalescing operator.", analysis.Fix{
		Title: "Use '??'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: r}} },
	})
}

// ncoSingle returns the only statement of a braced, non-alternative block.
func ncoSingle(s syntax.Stmt) (syntax.Stmt, bool) {
	b, ok := s.(*syntax.Block)
	if !ok || b.Alt || len(b.Stmts) != 1 {
		return nil, false
	}
	return b.Stmts[0], true
}

// ncoAssign returns the plain `=` assignment to a plain variable held by an
// expression statement.
func ncoAssign(s syntax.Stmt) (*syntax.Assign, bool) {
	es, ok := s.(*syntax.ExprStmt)
	if !ok {
		return nil, false
	}
	a, ok := es.Expr.(*syntax.Assign)
	if !ok || a.Op.Kind != syntax.TEqual {
		return nil, false
	}
	if v, ok := a.Var.(*syntax.Variable); !ok || v.NameExpr != nil {
		return nil, false
	}
	return a, true
}

func (nullCoalescingOperatorCanBeUsed) checkIf(ctx *analysis.Context, n *syntax.If) {
	if len(n.ElseIfs) > 0 || n.Alt { // D3
		return
	}
	c, ok := ncoClassify(ctx, n.Cond) // D4
	if !ok {
		return
	}
	body, ok := ncoSingle(n.Body) // D5
	if !ok {
		return
	}
	ret, isRet := body.(*syntax.Return)
	asg, isAsg := ncoAssign(body)
	if !isRet && !isAsg {
		return
	}
	var t, f syntax.Expr
	replace := n.Span()
	switch {
	case n.Else != nil:
		other, ok := ncoSingle(n.Else.Body)
		if !ok {
			return
		}
		if r2, ok := other.(*syntax.Return); ok && isRet { // S1
			t, f = ret.Expr, r2.Expr
		} else if a2, ok := ncoAssign(other); ok && isAsg && util.EquivalentFoldNames(ctx.File, asg.Var, a2.Var) { // S2
			t, f = asg.Value, a2.Value
		} else {
			return
		}
	case isAsg: // S3
		prev, ok := util.PrevStmt(ctx.File, n)
		if !ok {
			return
		}
		pa, ok := ncoAssign(prev)
		if !ok || !util.EquivalentFoldNames(ctx.File, pa.Var, asg.Var) || pa.ByRef {
			return
		}
		if _, ok := syntax.UnwrapParens(pa.Value).(*syntax.Assign); ok {
			return
		}
		reads := false
		syntax.Inspect(pa.Value, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && x != syntax.Node(pa.Value) && util.EquivalentFoldNames(ctx.File, v, pa.Var) {
				reads = true
			}
			return !reads
		})
		if reads || util.MayHaveSideEffects(pa.Value) { // the fallback is not always evaluated
			return
		}
		t, f = asg.Value, pa.Value
		replace.Start = prev.Span().Start
	default: // isRet
		next, ok := util.NextStmt(ctx.File, n)
		if ok {
			r2, ok := next.(*syntax.Return)
			if !ok { // S4 does not apply, and S5 requires no next statement
				return
			}
			t, f = ret.Expr, r2.Expr
			replace.End = next.Span().End
		} else { // S5
			blk, ok := n.Parent().(*syntax.Block)
			if !ok {
				return
			}
			switch blk.Parent().(type) { // a block directly under a function is its body
			case *syntax.Function, *syntax.Method, *syntax.Closure:
			default:
				return
			}
			t = ret.Expr
		}
	}
	if t == nil { // D7
		return
	}
	r, ok := ncoGenerate(ctx, c, t, f)
	if !ok {
		return
	}
	q := "return " + r
	if isAsg {
		q = ctx.Text(asg.Var) + " = " + r
	}
	text := q + ";"
	if _, ok := n.Parent().(*syntax.Else); ok {
		text = "{ " + q + "; }"
	}
	ctx.Report(syntax.Span{Start: n.Span().Start, End: n.Span().Start + 2}, "Simplify to '"+q+"' using the null coalescing operator.", analysis.Fix{
		Title: "Use '??'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: replace, NewText: text}} },
	})
}
