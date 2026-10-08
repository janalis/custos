package performance

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// cascadeStringReplacement reports consecutive or nested str_replace() calls
// on the same subject that can be one call, and search arrays made of one
// repeated literal.
type cascadeStringReplacement struct{}

func init() { register(cascadeStringReplacement{}) }

// Semantic marks the rule as needing the project index (symbols or types
// declared in other files).
func (cascadeStringReplacement) Semantic() {}

func (cascadeStringReplacement) ID() string { return "CascadeStringReplacement" }

func (cascadeStringReplacement) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KReturn, syntax.KAssign}
}

// csrCall is a str_replace() call with exactly three plain arguments.
type csrCall struct {
	call *syntax.FuncCall
	args []syntax.Expr // search, replace, subject
}

func csrMatch(ctx *analysis.Context, e syntax.Expr) (csrCall, bool) {
	f, ok := csrMatchN(ctx, e)
	return f, ok && len(f.args) == 3
}

// csrMatchN matches a str_replace() call with three plain arguments or four
// (the by-reference $count); args holds them all.
func csrMatchN(ctx *analysis.Context, e syntax.Expr) (csrCall, bool) {
	call, _ := perfCall(ctx, e, "str_replace")
	if call == nil {
		return csrCall{}, false
	}
	args, ok := perfPlainArgs(call.Args)
	if !ok || (len(args) != 3 && len(args) != 4) {
		return csrCall{}, false
	}
	return csrCall{call: call, args: args}, true
}

func (cascadeStringReplacement) Check(ctx *analysis.Context, n syntax.Node) {
	var f csrCall
	var ok bool
	switch x := n.(type) {
	case *syntax.Return: // D1
		if x.Expr == nil {
			return
		}
		f, ok = csrMatchN(ctx, syntax.UnwrapParens(x.Expr))
	case *syntax.Assign:
		if x.Op.Kind != syntax.TEqual || x.ByRef || x.Value == nil {
			return
		}
		f, ok = csrMatchN(ctx, syntax.UnwrapParens(x.Value))
	}
	if !ok { // D2
		return
	}
	c := &csrCtx{ctx: ctx, opt: ctx.Bool("USE_SHORT_ARRAYS_SYNTAX")}
	if len(f.args) == 4 {
		// A $count argument: merging would change (or lose) the count, so
		// only part C applies; collapsing identical searches keeps it.
		if simplified := c.simplifiedSearch(f); simplified != nil {
			c.reportSimplified(f, simplified, nil)
		}
		return
	}

	// A. cascading assignments (D3–D5), statement level only
	if stmt := csrStmtOf(n); stmt != nil {
		if _, _, ok := c.link(stmt); ok {
			c.reportCascade(f, stmt)
		}
	}

	// B. nested call (D6). When part C also applies to F, its
	// simplification is folded into the merge (C's fix comes first when the
	// fixes are applied in order), and both findings share that fix.
	var nestedFix *analysis.Fix
	simplified := c.simplifiedSearch(f)
	if inner, ok := csrMatch(c.ctx, f.args[2]); ok && c.mergeable(inner, f) {
		span := f.call.Span()
		strip, fixable := csrKeyMode([]csrCall{f, inner})
		nestedFix = &analysis.Fix{
			Title: "Merge the str_replace() calls",
			Edits: func() []analysis.TextEdit {
				c.strip = strip
				pm := c.model(f)
				if simplified != nil {
					pm.search = exprVal(simplified)
				}
				return []analysis.TextEdit{{Span: span, NewText: c.rebuild(f, c.mergeInto(f, pm, c.model(inner)))}}
			},
		}
		if fixable {
			ctx.Report(inner.call.Span(), "Fold this nested str_replace() into the enclosing call.", *nestedFix)
		} else {
			ctx.Report(inner.call.Span(), "Fold this nested str_replace() into the enclosing call.")
			nestedFix = nil
		}
	}

	// C. redundant search array (D7, D8)
	if simplified != nil {
		c.reportSimplified(f, simplified, nestedFix)
	}
}

// reportSimplified reports part C; nestedFix, when set, replaces the
// simplification fix.
func (c *csrCtx) reportSimplified(f csrCall, simplified syntax.Expr, nestedFix *analysis.Fix) {
	span := f.args[0].Span()
	text := c.ctx.Text(simplified)
	fx := analysis.Fix{
		Title: "Pass the single string",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: text}} },
	}
	if nestedFix != nil {
		fx = *nestedFix
	}
	c.ctx.ReportSeverity(span, meta.SeverityInfo, "All searched items are identical; pass the single string.", fx)
}

// simplifiedSearch implements D7/D8: the shared string literal of F's search
// array when F's replace is a string literal, else nil.
func (c *csrCtx) simplifiedSearch(f csrCall) syntax.Expr {
	if !isStringLit(f.args[1]) {
		return nil
	}
	arr, ok := f.args[0].(*syntax.Array)
	if !ok {
		return nil
	}
	var first syntax.Expr
	text := ""
	for _, it := range arr.Items {
		if it == nil {
			return nil
		}
		// str_replace() searches the values; keys are ignored (custos
		// diverges: upstream compares the key of `k => v` elements).
		e := it.Value
		if it.ByRef || !isStringLit(e) {
			return nil
		}
		t := c.ctx.Text(e)
		if first != nil && t != text {
			return nil
		}
		if first == nil {
			first, text = e, t
		}
	}
	return first
}

// csrStmtOf returns the statement for part A: the return itself, or the
// expression statement holding the assignment.
func csrStmtOf(n syntax.Node) syntax.Stmt {
	switch x := n.(type) {
	case *syntax.Return:
		return x
	case *syntax.Assign:
		if es, ok := x.Parent().(*syntax.ExprStmt); ok && es.Expr == syntax.Expr(x) {
			return es
		}
	}
	return nil
}

// csrStmtCall returns the str_replace() call of a `return F;` or `$v = F;`
// statement and, for assignments, the assigned variable name.
func csrStmtCall(ctx *analysis.Context, s syntax.Stmt) (f csrCall, target string, isReturn, ok bool) {
	switch x := s.(type) {
	case *syntax.Return:
		if x.Expr == nil {
			return csrCall{}, "", false, false
		}
		f, ok = csrMatch(ctx, syntax.UnwrapParens(x.Expr))
		return f, "", true, ok
	case *syntax.ExprStmt:
		a, isAssign := x.Expr.(*syntax.Assign)
		if !isAssign || a.Op.Kind != syntax.TEqual || a.ByRef || a.Value == nil {
			return csrCall{}, "", false, false
		}
		f, ok = csrMatch(ctx, syntax.UnwrapParens(a.Value))
		if !ok {
			return csrCall{}, "", false, false
		}
		target, _ = perfPlainVar(a.Var)
		return f, target, false, true
	}
	return csrCall{}, "", false, false
}

// link reports whether s cascades onto its preceding statement (D3, D4, D9)
// and returns that statement and its call.
func (c *csrCtx) link(s syntax.Stmt) (prev syntax.Stmt, g csrCall, ok bool) {
	f, target, isReturn, ok := csrStmtCall(c.ctx, s)
	if !ok {
		return nil, csrCall{}, false
	}
	prev, ok = util.PrevStmt(c.ctx.File, s)
	if !ok {
		return nil, csrCall{}, false
	}
	g, p, isRet, ok := csrStmtCall(c.ctx, prev)
	if !ok || isRet || p == "" {
		return nil, csrCall{}, false
	}
	if subj, ok := perfPlainVar(f.args[2]); !ok || subj != p {
		return nil, csrCall{}, false
	}
	if !isReturn && target != p {
		return nil, csrCall{}, false
	}
	// custos: a search/replace argument reading the variable sees the
	// first replacement's result; merged, it would see the original.
	if util.MentionsVariable(f.args[0], p) || util.MentionsVariable(f.args[1], p) {
		return nil, csrCall{}, false
	}
	if !c.mergeable(f, g) {
		return nil, csrCall{}, false
	}
	return prev, g, true
}

// reportCascade reports F (part A). The fix folds the whole chain of
// cascading statements around stmt into its last statement, which is what
// applying every cascade fix in turn produces.
func (c *csrCtx) reportCascade(f csrCall, stmt syntax.Stmt) {
	ctx := c.ctx
	const msg = "Fold this str_replace() into the preceding one on the same variable."
	stmts, calls := c.chain(stmt)
	strip, fixable := csrKeyMode(calls)
	if !fixable || len(calls) > csrMaxFixChain {
		ctx.Report(f.call.Span(), msg)
		return
	}
	ctx.Report(f.call.Span(), msg, analysis.Fix{
		Title: "Merge the str_replace() calls",
		Edits: func() []analysis.TextEdit {
			c.strip = strip
			// calls[len-1] is the head; fold forward
			acc := c.model(calls[len(calls)-1])
			for i := len(calls) - 2; i > 0; i-- {
				acc = c.mergeInto(calls[i], c.model(calls[i]), acc)
			}
			final := calls[0]
			edits := make([]analysis.TextEdit, 0, len(stmts))
			for i := len(stmts) - 1; i > 0; i-- {
				edits = append(edits, analysis.TextEdit{Span: util.WithTrailingWhitespace(ctx.File, stmts[i].Span())})
			}
			edits = append(edits, analysis.TextEdit{Span: final.call.Span(), NewText: c.rebuild(final, c.mergeInto(final, c.model(final), acc))})
			return edits
		},
	})
}

// csrMaxFixChain bounds the chains the merge fix folds: each finding of a
// chain carries a fix rebuilding the whole chain (quadratic text), so a
// huge generated chain is reported without a fix.
const csrMaxFixChain = 256

type csrChain struct {
	stmts []syntax.Stmt
	calls []csrCall
}

// chain returns the whole chain of cascading statements around stmt, last
// statement first, with their calls. Every statement of a chain shares it,
// so it is computed once per chain and file (per statement it was
// quadratic on long chains).
func (c *csrCtx) chain(stmt syntax.Stmt) ([]syntax.Stmt, []csrCall) {
	memo := c.ctx.Memo("chains", func() any { return map[syntax.Stmt]*csrChain{} }).(map[syntax.Stmt]*csrChain)
	if ch, ok := memo[stmt]; ok {
		return ch.stmts, ch.calls
	}
	stmts, calls := c.computeChain(stmt)
	ch := &csrChain{stmts, calls}
	for _, s := range stmts {
		memo[s] = ch
	}
	memo[stmt] = ch
	return stmts, calls
}

func (c *csrCtx) computeChain(stmt syntax.Stmt) ([]syntax.Stmt, []csrCall) {
	last := stmt
	for {
		next, ok := util.NextStmt(c.ctx.File, last)
		if !ok {
			break
		}
		if _, _, linked := c.link(next); !linked {
			break
		}
		last = next
	}
	var stmts []syntax.Stmt
	var calls []csrCall
	cur := last
	for {
		fc, _, _, _ := csrStmtCall(c.ctx, cur)
		stmts = append(stmts, cur)
		calls = append(calls, fc)
		prev, _, ok := c.link(cur)
		if !ok {
			break
		}
		cur = prev
	}
	return stmts, calls
}

// csrKeyMode decides how the merge fix treats explicit array keys (F3).
// Without any keyed element in the search/replace array literals of calls,
// elements are merged verbatim (strip = false). Otherwise keys are dropped
// from merged elements (strip = true), which str_replace() does not observe,
// provided each keyed literal has only int/string literal keys and no two
// elements of it share an effective key; else the merge cannot be written
// safely and no fix is offered (fixable = false).
func csrKeyMode(calls []csrCall) (strip, fixable bool) {
	for _, call := range calls {
		for _, a := range call.args[:2] {
			arr, ok := a.(*syntax.Array)
			if !ok {
				continue
			}
			keyed := false
			for _, it := range arr.Items {
				if it != nil && it.Key != nil {
					keyed = true
				}
			}
			if !keyed {
				continue
			}
			strip = true
			if !csrDistinctKeys(arr) {
				return true, false
			}
		}
	}
	return strip, true
}

// csrDistinctKeys reports whether every element of arr has an explicit
// int/string literal key or none (no spread), and no two elements end up
// with the same key under PHP's key casting and auto-indexing.
func csrDistinctKeys(arr *syntax.Array) bool {
	seen := map[string]bool{}
	next := int64(0)
	for _, it := range arr.Items {
		if it == nil || it.Unpack {
			return false
		}
		var k string
		if it.Key == nil {
			k = "i" + strconv.FormatInt(next, 10)
			next++
		} else {
			lit, ok := it.Key.(*syntax.Literal)
			if !ok {
				return false
			}
			var n int64
			isInt := false
			switch lit.LitKind {
			case syntax.LitInt:
				v, err := strconv.ParseInt(strings.ReplaceAll(lit.Raw, "_", ""), 0, 64)
				if err != nil {
					return false
				}
				n, isInt = v, true
			case syntax.LitString:
				v, ok := util.StringLiteralValue(lit.Raw)
				if !ok {
					return false
				}
				if i, err := strconv.ParseInt(v, 10, 64); err == nil && strconv.FormatInt(i, 10) == v {
					n, isInt = i, true
				} else {
					k = "s" + v
				}
			default:
				return false
			}
			if isInt {
				if n < 0 {
					return false
				}
				k = "i" + strconv.FormatInt(n, 10)
				if n >= next {
					next = n + 1
				}
			}
		}
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	return true
}

type csrCtx struct {
	ctx   *analysis.Context
	opt   bool // USE_SHORT_ARRAYS_SYNTAX
	strip bool // drop explicit keys of merged array elements (csrKeyMode)
}

// arrayOnly reports whether e's known types include array but not string.
func (c *csrCtx) arrayOnly(e syntax.Expr) bool {
	if _, ok := e.(*syntax.Array); ok {
		return false
	}
	arr, str := false, false
	for _, a := range c.ctx.TypeOf(e).Atoms() {
		switch {
		case a == "string":
			str = true
		case a == "array" || strings.HasSuffix(a, "[]"):
			arr = true
		}
	}
	return arr && !str
}

// mergeable implements D9.
func (c *csrCtx) mergeable(a, b csrCall) bool {
	if c.ctx.PHP >= phpver.PHP74 {
		return true
	}
	for _, e := range []syntax.Expr{a.args[0], a.args[1], b.args[0], b.args[1]} {
		if c.arrayOnly(e) {
			return false
		}
	}
	return true
}

// csrVal is a (possibly merged) search or replace argument.
type csrVal struct {
	expr     syntax.Expr   // unmodified source argument (nil when merged)
	src      *syntax.Array // source array literal with elements merged in front
	inserted []string      // elements merged in front of src
	elems    []string      // synthetic array (expr == nil && src == nil)
}

func exprVal(e syntax.Expr) csrVal { return csrVal{expr: e} }

func (v csrVal) isArray() bool {
	if v.expr != nil {
		_, ok := v.expr.(*syntax.Array)
		return ok
	}
	return true
}

// csrModel is a call reduced to its arguments, as merging goes on.
type csrModel struct {
	search, replace csrVal
	subject         string
}

func (c *csrCtx) model(x csrCall) csrModel {
	return csrModel{search: exprVal(x.args[0]), replace: exprVal(x.args[1]), subject: c.ctx.Text(x.args[2])}
}

func (c *csrCtx) elemTexts(arr *syntax.Array) []string {
	out := make([]string, 0, len(arr.Items))
	for _, it := range arr.Items {
		switch {
		case it == nil:
		case c.strip && it.Key != nil && it.Value != nil:
			t := c.ctx.Text(it.Value)
			if it.ByRef {
				t = "&" + t
			}
			out = append(out, t)
		default:
			out = append(out, c.ctx.Text(it))
		}
	}
	return out
}

// contribution returns the elements a value contributes to a merge (F3).
func (c *csrCtx) contribution(v csrVal) []string {
	switch {
	case v.src != nil:
		return append(append([]string(nil), v.inserted...), c.elemTexts(v.src)...)
	case v.expr == nil:
		return v.elems
	}
	if arr, ok := v.expr.(*syntax.Array); ok {
		return c.elemTexts(arr)
	}
	if c.arrayOnly(v.expr) {
		// F3: spread through array_values() so string keys neither throw
		// (7.4/8.0) nor collide and drop entries (8.1+); str_replace()
		// ignores keys, so the result is unchanged.
		return []string{"..." + util.QualifiedBuiltin(c.ctx, "array_values", v.expr.Span().Start) + "(" + c.ctx.Text(v.expr) + ")"}
	}
	return []string{c.ctx.Text(v.expr)}
}

func (c *csrCtx) count(v csrVal) int { return len(c.contribution(v)) }

// merge implements F3: from's elements followed by to's.
func (c *csrCtx) merge(to, from csrVal) csrVal {
	fe := c.contribution(from)
	if arr, ok := to.expr.(*syntax.Array); ok {
		if len(arr.Items) == 0 { // see spec Divergences: keep from's elements
			return csrVal{elems: fe}
		}
		if c.strip { // keys dropped: rebuild instead of inserting into to's text
			return csrVal{elems: append(fe, c.elemTexts(arr)...)}
		}
		return csrVal{src: arr, inserted: fe}
	}
	// to is never a src value here: merges always start from a call's own
	// (exprVal) argument or an expanded replace (elems).
	return csrVal{elems: append(append([]string(nil), fe...), c.contribution(to)...)}
}

// unbox returns the single unkeyed string literal of a one-element array
// literal, else e.
func csrUnbox(e syntax.Expr) syntax.Expr {
	arr, ok := e.(*syntax.Array)
	if !ok || len(arr.Items) != 1 || arr.Items[0] == nil || arr.Items[0].Key != nil || arr.Items[0].ByRef || arr.Items[0].Unpack {
		return e
	}
	if isStringLit(arr.Items[0].Value) {
		return arr.Items[0].Value
	}
	return e
}

// constValue resolves a global constant to its single in-file define() value.
func (c *csrCtx) constValue(e syntax.Expr) syntax.Expr {
	cf, ok := e.(*syntax.ConstFetch)
	if !ok || cf.Name == nil {
		return e
	}
	name := util.LastNamePart(cf.Name.Value)
	switch strings.ToLower(name) {
	case "true", "false", "null":
		return e
	}
	var found syntax.Expr
	count := 0
	syntax.InspectFile(c.ctx.File, func(n syntax.Node) bool {
		call, ok := n.(*syntax.FuncCall)
		if !ok || !c.ctx.IsGlobalFunctionCall(call, "define") {
			return true
		}
		args, ok := util.CallArgValues(call)
		if !ok || len(args) < 2 {
			return true
		}
		if lit, ok := args[0].(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
			if v, ok := util.StringLiteralValue(lit.Raw); ok && strings.TrimPrefix(v, `\`) == name {
				found = args[1]
				count++
			}
		}
		return true
	})
	if count == 1 {
		return found
	}
	return e
}

func isStringLit(e syntax.Expr) bool {
	lit, ok := e.(*syntax.Literal)
	return ok && lit.LitKind == syntax.LitString
}

// expand implements the F1 expansion of a replace argument.
func (c *csrCtx) expand(m csrModel) csrVal {
	if !m.search.isArray() || c.count(m.search) < 2 {
		return m.replace
	}
	if m.replace.expr == nil {
		return m.replace // already an array
	}
	if _, isArr := c.constValue(m.replace.expr).(*syntax.Array); isArr {
		return m.replace
	}
	n := c.count(m.search)
	text := c.ctx.Text(m.replace.expr)
	elems := make([]string, n)
	for i := range elems {
		elems[i] = text
	}
	return csrVal{elems: elems}
}

// mergeInto folds e (eliminated) into the call p (patched): F1, F2, F4.
func (c *csrCtx) mergeInto(p csrCall, pm, e csrModel) csrModel {
	out := csrModel{subject: e.subject}
	pu := csrUnbox(p.args[1])
	var eu syntax.Expr
	if e.replace.expr != nil {
		eu = csrUnbox(e.replace.expr)
	}
	if eu != nil && isStringLit(c.constValue(pu)) && isStringLit(c.constValue(eu)) &&
		pu.Kind() == eu.Kind() && c.ctx.Text(pu) == c.ctx.Text(eu) {
		out.replace = exprVal(pu)
	} else {
		out.replace = c.merge(c.expand(pm), c.expand(e))
	}
	out.search = c.merge(pm.search, e.search)
	return out
}

// render implements F5 for one argument.
func (c *csrCtx) render(v csrVal) string {
	switch {
	case v.src != nil:
		if v.src.Short != c.opt {
			return c.wrap(c.contribution(v))
		}
		src := c.ctx.Text(v.src)
		open := 1 // after `[`
		if !v.src.Short {
			open = strings.IndexByte(src, '(') + 1
		}
		for open < len(src) && strings.IndexByte(" \t\r\n", src[open]) >= 0 {
			open++
		}
		return src[:open] + strings.Join(v.inserted, ", ") + ", " + src[open:]
	case v.expr == nil:
		return c.wrap(v.elems)
	}
	// An unmerged value is only ever a replace kept as is (F2: a string
	// literal or constant), never an array literal.
	return c.ctx.Text(v.expr)
}

func (c *csrCtx) wrap(elems []string) string {
	if c.opt {
		return "[" + strings.Join(elems, ", ") + "]"
	}
	return "array(" + strings.Join(elems, ", ") + ")"
}

// rebuild re-emits the patched call p with merged arguments (F5).
func (c *csrCtx) rebuild(p csrCall, m csrModel) string {
	qual, _, _ := util.CallName(p.call)
	return qual + "str_replace(" + c.render(m.search) + ", " + c.render(m.replace) + ", " + m.subject + ")"
}
