package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/phpdoc"
	"custos/internal/syntax"
	"custos/internal/types"
)

// nullPointerException reports dereferences of possibly-null values
// (nullable object parameters, nullable locals, chained nullable calls).
// It runs per file so that the chained-call strategy, which also scans nested
// closures from their enclosing unit, reports each range only once.
type nullPointerException struct{}

func init() { register(nullPointerException{}) }

// Semantic marks the rule as needing the project symbol index.
func (nullPointerException) Semantic() {}

func (nullPointerException) ID() string { return "NullPointerException" }

func (nullPointerException) Kinds() []syntax.NodeKind { return nil }

func (nullPointerException) Check(*analysis.Context, syntax.Node) {}

const npeMessage = "Possible null dereference."

func (nullPointerException) CheckFile(ctx *analysis.Context) {
	reported := map[syntax.Span]bool{}
	report := func(s syntax.Span) {
		if !reported[s] && s.Len() > 0 {
			reported[s] = true
			ctx.Report(s, npeMessage)
		}
	}
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		switch f := n.(type) {
		case *syntax.Method:
			if f.Body == nil || util.InTestContext(ctx, f) {
				return true
			}
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction:
		default:
			return true
		}
		u := &npeUnit{ctx: ctx, fn: n, body: syntax.FuncLikeBody(n), report: report}
		u.strategyParams()
		u.strategyChains()
		u.strategyLocals()
		return true
	})
}

type npeUnit struct {
	ctx    *analysis.Context
	fn     syntax.Node
	body   *syntax.Block
	report func(syntax.Span)
}

// npeObjectOnly reports whether t (minus null/void) is a non-empty set of
// class names, self, static or object.
func npeObjectOnly(t types.Type) bool {
	n := 0
	for _, a := range t.Atoms() {
		switch {
		case a == "null" || a == "void":
			continue
		case a == "self" || a == "static" || a == "object":
		case strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]"):
		default:
			return false
		}
		n++
	}
	return n > 0
}

func (u *npeUnit) declaredType(n syntax.Expr) types.Type {
	at := n.Span().Start
	return types.FromNode(n, func(w string) string { return u.ctx.Names().Class(w, at) })
}

// ---- strategy A ----------------------------------------------------------------------

func (u *npeUnit) strategyParams() {
	if u.body == nil { // D1
		return
	}
	for _, p := range syntax.FuncLikeParams(u.fn) { // D2
		if p.Var == nil || p.Var.Name == "" || p.Type == nil {
			continue
		}
		t := u.declaredType(p.Type)
		if !t.Has("null") && !(p.Default != nil && syntax.IsNullConst(p.Default)) {
			continue
		}
		if !npeObjectOnly(t.Without("void")) || t.Has("void") {
			continue
		}
		u.walk(p.Var.Name, nil)
	}
}

// ---- strategy B ----------------------------------------------------------------------

func (u *npeUnit) strategyChains() {
	var root syntax.Node = u.body
	if af, ok := u.fn.(*syntax.ArrowFunction); ok {
		root = af.Expr
	}
	// root is never nil: methods without a body are skipped, functions and
	// closures always have one, arrow functions an expression.
	var tested []syntax.Node
	syntax.Inspect(root, func(n syntax.Node) bool {
		mc, ok := n.(*syntax.MethodCall)
		if !ok || mc.NullSafe {
			return true
		}
		switch mc.Var.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
			t := u.chainBaseType(mc.Var)
			if t.HasAny("null", "void") {
				seen := false
				for _, e := range tested {
					if util.Equivalent(u.ctx.File, e, mc.Var) {
						seen = true
						break
					}
				}
				if !seen {
					if tok, ok := util.NextSignificant(u.ctx.File, mc.Var.Span().End); ok && tok.Kind == syntax.TObjectOperator {
						u.report(syntax.Span{Start: tok.Start, End: tok.End})
					}
				}
			}
		}
		if npeNullTested(mc) {
			tested = append(tested, mc)
		}
		return true
	})
}

// chainBaseType is the return type of call x as seen by a chained `->`. The
// null added by a nullsafe `?->` call does not count: PHP short-circuits the
// rest of the chain, so only the callee's own declared null/void does.
func (u *npeUnit) chainBaseType(x syntax.Expr) types.Type {
	t := u.ctx.TypeOf(x)
	mc, ok := x.(*syntax.MethodCall)
	if !ok || !mc.NullSafe || !t.Has("null") {
		return t
	}
	name := npeMemberName(mc.Name) // "" (no method) for a dynamic name
	for _, cls := range u.ctx.TypeOf(mc.Var).Classes() {
		m := u.ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), name, u.ctx.PHP)
		if m == nil || types.FromDoc(m.Return, nil).HasAny("null", "void") || types.FromDoc(m.DocReturn, nil).HasAny("null", "void") {
			return t
		}
	}
	return t.Without("null")
}

// npeNullTested reports whether call is used as a null test.
func npeNullTested(call syntax.Expr) bool {
	switch p := call.Parent().(type) {
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
			other := p.Left
			if other == call {
				other = p.Right
			}
			return syntax.IsNullConst(other)
		case syntax.TBooleanAnd, syntax.TAnd:
			return true
		}
		return false
	case *syntax.Instanceof:
		return p.Expr == call
	}
	return npeLogicalOperand(call, false)
}

// npeLogicalOperand reports whether e (through parentheses) is a condition
// of if/elseif/while/do-while, the operand of `!`, the condition of a full
// ternary, or (withBinary) an operand of a logical binary operator.
func npeLogicalOperand(e syntax.Node, withBinary bool) bool {
	cur := e
	for {
		p := cur.Parent()
		switch p := p.(type) {
		case *syntax.Paren:
			cur = p
			continue
		case *syntax.If:
			return p.Cond == cur
		case *syntax.ElseIf:
			return p.Cond == cur
		case *syntax.While:
			return p.Cond == cur
		case *syntax.DoWhile:
			return p.Cond == cur
		case *syntax.Unary:
			return p.Op.Kind == syntax.TExclaim
		case *syntax.Ternary:
			return p.Then != nil && p.Cond == cur
		case *syntax.Binary:
			if !withBinary {
				return false
			}
			switch p.Op.Kind {
			case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
				return true
			}
		}
		return false
	}
}

// ---- strategy C ----------------------------------------------------------------------

func (u *npeUnit) strategyLocals() {
	if u.body == nil { // D4
		return
	}
	params := map[string]bool{}
	for _, p := range syntax.FuncLikeParams(u.fn) {
		if p.Var != nil {
			params[p.Var.Name] = true
		}
	}
	firsts := map[string]*syntax.Assign{}
	var order []string
	syntax.Inspect(u.body, func(n syntax.Node) bool {
		a, ok := n.(*syntax.Assign)
		if !ok || a.Op.Kind != syntax.TEqual {
			return true
		}
		if _, ok := a.Parent().(*syntax.ExprStmt); !ok {
			return true
		}
		v, ok := a.Var.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name == "" || params[v.Name] {
			return true
		}
		switch a.Value.(type) {
		case *syntax.PropertyFetch, *syntax.Unary, *syntax.Clone:
			return true
		}
		if _, ok := firsts[v.Name]; !ok {
			firsts[v.Name] = a
			order = append(order, v.Name)
		}
		return true
	})
	for _, name := range order { // D5, D6
		a := firsts[name]
		if u.nullableAssign(a, name) {
			u.walk(name, a)
		}
	}
}

// nullableAssign applies D5 to an assignment of $name.
func (u *npeUnit) nullableAssign(a *syntax.Assign, name string) bool {
	t := u.ctx.TypeOf(a.Value)
	if !t.HasAny("null", "void") || !npeObjectOnly(t) {
		return false
	}
	// a is a plain `=` assignment (both callers check it)
	stmt, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return true
	}
	tok, ok := util.TokenBefore(u.ctx.File, stmt.Span().Start)
	for ok && tok.Kind == syntax.TWhitespace {
		tok, ok = util.TokenBefore(u.ctx.File, tok.Start)
	}
	if !ok || tok.Kind != syntax.TDocComment {
		return true
	}
	vars := phpdoc.Parse(u.ctx.SpanText(syntax.Span{Start: tok.Start, End: tok.End})).All("var")
	if len(vars) != 1 {
		return true
	}
	typ, rest := phpdoc.SplitType(vars[0].Text)
	if strings.HasPrefix(typ, "$") {
		if phpdoc.VarName(typ) != name {
			return true
		}
		typ, _ = phpdoc.SplitType(rest)
	} else if phpdoc.VarName(rest) != name {
		return true
	}
	return types.FromDoc(typ, nil).Has("null")
}

// ---- usage walk ----------------------------------------------------------------------

func npeVarsNamed(root syntax.Node, name string, strict bool, fn func(*syntax.Variable)) {
	syntax.Inspect(root, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name && !util.IsStaticPropName(v) {
			if !strict || n != root {
				fn(v)
			}
		}
		return true
	})
}

func (u *npeUnit) usages(name string) []*syntax.Variable {
	var out []*syntax.Variable
	npeVarsNamed(u.body, name, false, func(v *syntax.Variable) {
		if syntax.EnclosingFuncLike(v) != u.fn {
			return
		}
		if a, ok := v.Parent().(*syntax.Assign); ok {
			added := map[*syntax.Variable]bool{}
			npeVarsNamed(a.Value, name, true, func(x *syntax.Variable) {
				added[x] = true
				out = append(out, x)
			})
			npeVarsNamed(a, name, false, func(x *syntax.Variable) {
				if !added[x] {
					out = append(out, x)
				}
			})
			return
		}
		npeVarsNamed(v.Parent(), name, true, func(x *syntax.Variable) { out = append(out, x) })
	})
	return out
}

func (u *npeUnit) walk(name string, decl *syntax.Assign) {
	list := u.usages(name)
	i := 0
	if decl != nil {
		for i < len(list) && list[i].Parent() != syntax.Node(decl) {
			i++
		}
		i++
	}
	for ; i < len(list); i++ {
		v := list[i]
		stop, rep := u.evaluate(v, name, decl)
		if rep && u.guardedByCondition(v, name) { // U12
			rep = false
		}
		if rep {
			u.report(v.Span())
		}
		if stop {
			return
		}
	}
}

// npeAssertNames are matched case-insensitively (lower-case keys), like PHP
// method names.
var npeAssertNames = map[string]bool{
	"assertnotnull": true, "assertinstanceof": true, "notnull": true, "isinstanceof": true, "isinstanceofany": true,
}

// npeMemberName returns the exact (case-preserved) member name of a call.
func npeMemberName(n syntax.Expr) string {
	if id, ok := n.(*syntax.Identifier); ok {
		return id.Value
	}
	return ""
}

// npeArgCall returns the call having v directly as an argument value, its
// positional index and, for a named argument, the parameter name.
func npeArgCall(v syntax.Expr) (call syntax.Node, idx int, name string) {
	arg, ok := v.Parent().(*syntax.Arg)
	if !ok || arg.Value != v {
		return nil, -1, ""
	}
	if arg.Name != nil {
		name = arg.Name.Value
	}
	list := arg.Parent().(*syntax.ArgList) // arguments only live in argument lists
	idx = -1
	for i, a := range list.Args {
		if a == syntax.Expr(arg) {
			idx = i
		}
	}
	return list.Parent(), idx, name
}

func (u *npeUnit) evaluate(v *syntax.Variable, name string, decl *syntax.Assign) (stop, report bool) {
	p := v.Parent()
	var g syntax.Node
	if p != nil {
		g = p.Parent()
	}
	switch p := p.(type) {
	case *syntax.Instanceof: // U1
		if p.Expr == syntax.Expr(v) {
			return true, false
		}
	case *syntax.Binary: // U2
		switch p.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
			other := p.Left
			if other == syntax.Expr(v) {
				other = p.Right
			}
			return syntax.IsNullConst(other), false
		}
	case *syntax.Isset, *syntax.Empty: // U3
		return true, false
	case *syntax.Catch: // U4
		return true, false
	}
	if npeLogicalOperand(v, true) { // U3
		return true, false
	}
	call, idx, argName := npeArgCall(v)
	if call != nil { // U5
		var nm string
		switch c := call.(type) {
		case *syntax.Clone:
			if c.Expr == syntax.Expr(v) {
				return false, true
			}
		case *syntax.MethodCall: // ->, ?->
			nm = npeMemberName(c.Name)
		case *syntax.StaticCall: // ::
			nm = npeMemberName(c.Name)
		case *syntax.FuncCall:
			if n, ok := c.Name.(*syntax.Name); ok && u.ctx.IsGlobalFunctionCall(c, "is_null") &&
				strings.EqualFold(util.LastNamePart(n.Value), "is_null") {
				return true, false
			}
		}
		if npeAssertNames[strings.ToLower(nm)] || (strings.EqualFold(nm, "that") && npeChainedNotNull(call)) {
			return true, false
		}
	}
	switch p := p.(type) {
	case *syntax.Assign: // U6
		// `$v ??= <non-null>` leaves $v non-null too (custos diverges).
		if p == decl || p.Op.Kind != syntax.TEqual && p.Op.Kind != syntax.TCoalesceEqual {
			return false, false
		}
		if t, ok := p.Var.(*syntax.Variable); ok && t.NameExpr == nil && t.Name == name && !u.nullableAssign(p, name) {
			return true, false
		}
		return false, false
	case *syntax.ArrayDimFetch: // U7
		return false, p.Var == syntax.Expr(v)
	case *syntax.PropertyFetch: // U8
		if p.Var != syntax.Expr(v) || p.NullSafe {
			return false, false
		}
		var top syntax.Node = p
		for {
			switch q := top.Parent().(type) {
			case *syntax.PropertyFetch:
				if q.Var == top {
					top = q
					continue
				}
			case *syntax.ArrayDimFetch:
				if q.Var == top {
					top = q
					continue
				}
			}
			break
		}
		if b, ok := top.Parent().(*syntax.Binary); ok && b.Op.Kind == syntax.TCoalesce && b.Left == top {
			return false, false
		}
		if is, ok := top.Parent().(*syntax.Isset); ok {
			for _, x := range is.Vars {
				if x == top {
					return false, false
				}
			}
		}
		return false, true
	case *syntax.MethodCall: // U8
		if p.Var != syntax.Expr(v) || p.NullSafe {
			return false, false
		}
		if decl != nil && g == syntax.Node(decl) {
			return false, false
		}
		return false, true
	case *syntax.FuncCall: // U9
		return false, p.Name == syntax.Expr(v)
	case *syntax.Clone: // U10
		return false, p.Expr == syntax.Expr(v)
	}
	if call != nil && idx >= 0 { // U11
		if prm := u.calleeParam(call, idx, argName); prm != nil {
			t := types.FromDoc(prm.Type, nil)
			if !t.Has("null") && !strings.EqualFold(strings.TrimPrefix(prm.Default, `\`), "null") && npeObjectOnly(t) {
				return false, true
			}
		}
	}
	return false, false
}

// npeChainedNotNull reports whether call is, climbing through directly
// enclosing method calls (any operator) having the previous as their object,
// the object of a call named `notNull` (case-insensitive).
func npeChainedNotNull(call syntax.Node) bool {
	cur := call
	for {
		var nm string
		object := false // cur is the object of its parent call
		switch mc := cur.Parent().(type) {
		case *syntax.MethodCall:
			object, nm = mc.Var == cur, npeMemberName(mc.Name)
		case *syntax.StaticCall:
			object, nm = mc.Class == cur, npeMemberName(mc.Name)
		}
		if !object {
			return false
		}
		if strings.EqualFold(nm, "notNull") {
			return true
		}
		cur = cur.Parent()
	}
}

// calleeParam returns the parameter of the resolved callee receiving the
// argument at position idx, or the one called name for a named argument.
func (u *npeUnit) calleeParam(call syntax.Node, idx int, name string) *index.Param {
	c, ok := resolveCallee(u.ctx, call)
	if !ok {
		return nil
	}
	if name != "" {
		for i := range c.params {
			if c.params[i].Name == name {
				return &c.params[i]
			}
		}
		return nil
	}
	if idx >= len(c.params) {
		return nil
	}
	return &c.params[idx]
}

// ---- U12: enclosing conditions -------------------------------------------------------

// guardedByCondition reports whether usage v of $name sits in a branch that
// an enclosing condition only enters when $name is not null (if/elseif/else
// bodies, while bodies, ternary branches, right operands of &&/||), and
// $name is not assigned between that condition and v.
func (u *npeUnit) guardedByCondition(v *syntax.Variable, name string) bool {
	var cur syntax.Node = v
	for p := cur.Parent(); p != nil && p != u.fn; cur, p = p, p.Parent() {
		var conds []syntax.Expr // conditions whose truth value is implied
		var wants []bool
		switch x := p.(type) {
		case *syntax.If:
			if cur == syntax.Node(x.Body) {
				conds, wants = append(conds, x.Cond), append(wants, true)
			}
		case *syntax.ElseIf:
			if cur == syntax.Node(x.Body) {
				conds, wants = append(conds, x.Cond), append(wants, true)
				if ifs, ok := x.Parent().(*syntax.If); ok {
					conds, wants = append(conds, ifs.Cond), append(wants, false)
					for _, ei := range ifs.ElseIfs {
						if ei == x {
							break
						}
						conds, wants = append(conds, ei.Cond), append(wants, false)
					}
				}
			}
		case *syntax.Else:
			if ifs, ok := x.Parent().(*syntax.If); ok && cur == syntax.Node(x.Body) {
				conds, wants = append(conds, ifs.Cond), append(wants, false)
				for _, ei := range ifs.ElseIfs {
					conds, wants = append(conds, ei.Cond), append(wants, false)
				}
			}
		case *syntax.While:
			if cur == syntax.Node(x.Body) {
				conds, wants = append(conds, x.Cond), append(wants, true)
			}
		case *syntax.Ternary:
			if x.Then != nil && cur == syntax.Node(x.Then) {
				conds, wants = append(conds, x.Cond), append(wants, true)
			} else if cur == syntax.Node(x.Else) {
				conds, wants = append(conds, x.Cond), append(wants, false)
			}
		case *syntax.Binary:
			if cur == syntax.Node(x.Right) {
				switch x.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					conds, wants = append(conds, x.Left), append(wants, true)
				case syntax.TBooleanOr, syntax.TOr:
					conds, wants = append(conds, x.Left), append(wants, false)
				}
			}
		case *syntax.MatchArm:
			// match (true) { c => v }: the arm runs only when c === true.
			if m, ok := x.Parent().(*syntax.Match); ok && cur == syntax.Node(x.Body) && len(x.Conds) == 1 && npeIsTrueConst(m.Cond) {
				conds, wants = append(conds, x.Conds[0]), append(wants, true)
			}
		default:
			// An earlier `if (c) { …exit… }` in the same statement list: the
			// statements after it run only when c was false.
			if list, ok := syntax.StmtListOf(p); ok {
				for _, st := range list[:max(syntax.StmtIndex(list, cur), 0)] {
					if ifs, ok := st.(*syntax.If); ok && syntax.Terminates(ifs.Body) {
						conds, wants = append(conds, ifs.Cond), append(wants, false)
					}
				}
			}
		}
		for i, c := range conds {
			if c != nil && u.implies(c, name, wants[i]) && !u.assignedBetween(name, c, v) {
				return true
			}
		}
	}
	return false
}

// npeIsTrueConst reports whether e is the constant true (any case).
func npeIsTrueConst(e syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "true")
}

// implies reports whether e evaluating to want guarantees $name is not null.
func (u *npeUnit) implies(e syntax.Expr, name string, want bool) bool {
	e = syntax.UnwrapParens(e)
	if un, ok := e.(*syntax.Unary); ok && un.Op.Kind == syntax.TExclaim {
		return u.implies(un.Expr, name, !want)
	}
	if b, ok := e.(*syntax.Binary); ok {
		switch b.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			return want && (u.implies(b.Left, name, true) || u.implies(b.Right, name, true))
		case syntax.TBooleanOr, syntax.TOr:
			return !want && (u.implies(b.Left, name, false) || u.implies(b.Right, name, false))
		}
	}
	if want {
		return u.impliesWhenTrue(e, name)
	}
	return u.impliesWhenFalse(e, name)
}

func npeIsVar(e syntax.Expr, name string) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
}

// npeRootedAt reports whether e is $name or a property/method/array access
// chain (nullsafe or not) starting at $name; depth counts the accesses.
func npeRootedAt(e syntax.Expr, name string) (ok bool, depth int) {
	for {
		e = syntax.UnwrapParens(e)
		switch x := e.(type) {
		case *syntax.Variable:
			return x.NameExpr == nil && x.Name == name, depth
		case *syntax.PropertyFetch:
			e = x.Var
		case *syntax.MethodCall:
			e = x.Var
		case *syntax.ArrayDimFetch:
			e = x.Var
		default:
			return false, 0
		}
		depth++
	}
}

func (u *npeUnit) globalCallOn(e syntax.Expr, fn, name string) bool {
	c, ok := e.(*syntax.FuncCall)
	if !ok || !u.ctx.IsGlobalFunctionCall(c, fn) {
		return false
	}
	args, ok := util.CallArgValues(c)
	return ok && len(args) == 1 && npeIsVar(args[0], name)
}

func (u *npeUnit) impliesWhenTrue(e syntax.Expr, name string) bool {
	switch x := e.(type) {
	case *syntax.Variable, *syntax.PropertyFetch, *syntax.MethodCall, *syntax.ArrayDimFetch:
		ok, _ := npeRootedAt(x, name) // truthy value or chain
		return ok
	case *syntax.Instanceof:
		return npeIsVar(x.Expr, name)
	case *syntax.Isset:
		for _, a := range x.Vars {
			if ok, _ := npeRootedAt(a, name); ok {
				return true
			}
		}
	case *syntax.FuncCall:
		return u.globalCallOn(x, "is_object", name)
	case *syntax.Binary:
		l, r := x.Left, x.Right
		if syntax.IsNullConst(syntax.UnwrapParens(l)) {
			l, r = r, l
		}
		switch x.Op.Kind {
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			if syntax.IsNullConst(syntax.UnwrapParens(r)) {
				ok, _ := npeRootedAt(l, name)
				return ok
			}
		case syntax.TIsIdentical, syntax.TIsEqual:
			// $n === E, or a chain rooted at $n (`$n->p === E`): a null $n
			// makes the chain null (or throws), so a non-null E proves it.
			if ok, _ := npeRootedAt(r, name); ok {
				l, r = r, l
			}
			if ok, _ := npeRootedAt(l, name); !ok || syntax.IsNullConst(syntax.UnwrapParens(r)) {
				return false
			}
			t := u.ctx.TypeOf(r)
			if t.IsUnknown() || t.HasAny("null", "void", "mixed") {
				return false
			}
			// loose == null also holds for '', 0, false and []: only objects
			// compared loosely prove non-null.
			return x.Op.Kind == syntax.TIsIdentical || npeObjectOnly(t)
		}
	}
	return false
}

func (u *npeUnit) impliesWhenFalse(e syntax.Expr, name string) bool {
	switch x := e.(type) {
	case *syntax.Empty:
		ok, _ := npeRootedAt(x.Expr, name)
		return ok
	case *syntax.FuncCall:
		return u.globalCallOn(x, "is_null", name)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsEqual:
			l, r := x.Left, x.Right
			if syntax.IsNullConst(syntax.UnwrapParens(l)) {
				l, r = r, l
			}
			if syntax.IsNullConst(syntax.UnwrapParens(r)) {
				ok, _ := npeRootedAt(l, name)
				return ok
			}
		}
	}
	return false
}

// assignedBetween reports whether $name is the target of an assignment (any
// operator, by reference, destructuring, or a foreach/catch variable)
// completed between the end of condition c and usage v in the unit body.
// An assignment in a statement list that then leaves for good (return,
// throw, exit) or jumps back through c (continue or break out of a loop
// holding c) never reaches v and does not count.
func (u *npeUnit) assignedBetween(name string, c syntax.Expr, v syntax.Node) bool {
	from, to := c.Span().End, v.Span().Start
	found := false
	syntax.Inspect(u.body, func(n syntax.Node) bool {
		if found {
			return false
		}
		sp := n.Span()
		if sp.End <= from || sp.Start >= to {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			if x.Span().Start >= from && x.Span().End <= to && !npeDiverted(x, c, v) { // completed before v
				// The variable itself or a destructuring slot: writing
				// `$n->p` or `$n['k']` leaves $n as it was.
				syntax.Inspect(x.Var, func(t syntax.Node) bool {
					switch t.(type) {
					case *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
						return false
					}
					if npeIsVar(util.AsExpr(t), name) {
						found = true
					}
					return !found
				})
			}
		case *syntax.Foreach:
			if x.Span().Start >= from && (util.MentionsVariable(x.Key, name) || util.MentionsVariable(x.Value, name)) {
				found = true
			}
		}
		return !found
	})
	return found
}

// npeDiverted reports whether control cannot flow from assignment a to v
// without re-evaluating c: a statement list holding a but not v ends, after
// a, in a return, throw or exit, or in a continue or break whose nearest
// loop holds c (and so v, which c guards): whatever the level, v runs again
// only after the loop is re-entered, which re-checks c first.
func npeDiverted(a *syntax.Assign, c syntax.Expr, v syntax.Node) bool {
	var cur syntax.Node = a
	for p := a.Parent(); p != nil && !syntax.IsFuncLike(p); cur, p = p, p.Parent() {
		if util.NodeContains(p, v) {
			return false
		}
		list, ok := syntax.StmtListOf(p)
		if !ok {
			continue
		}
		t := syntax.FirstTerminating(p)
		if t == len(list) || syntax.StmtIndex(list, cur) > t {
			continue
		}
		switch x := list[t].(type) {
		case *syntax.Return:
			return true
		case *syntax.ExprStmt:
			if _, ok := syntax.UnwrapParens(x.Expr).(*syntax.Throw); ok {
				return true
			}
			return syntax.ExitInvocation(x.Expr)
		case *syntax.Continue, *syntax.Break:
			return npeLoopHolds(x, c)
		}
		return false
	}
	return false
}

// npeLoopHolds reports whether the nearest loop a continue/break jump leaves
// holds c (a switch in between: no, the jump may only leave the switch).
func npeLoopHolds(jump syntax.Node, c syntax.Expr) bool {
	p := jump.Parent()
	for p != nil && !npeJumpTarget(p) {
		p = p.Parent()
	}
	_, sw := p.(*syntax.Switch)
	return p != nil && !sw && util.NodeContains(p, c)
}

// npeJumpTarget reports whether a continue/break may leave n: a loop or a
// switch.
func npeJumpTarget(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile, *syntax.Switch:
		return true
	}
	return false
}
