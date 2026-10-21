package nullpointerexception

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// nullableObject is what a reported variable may hold (see U12).
var nullableObject = types.Of("object", "null")

// nullPointerException reports dereferences of possibly-null values
// (nullable object parameters, nullable locals, chained nullable calls).
// It runs per file so that the chained-call strategy, which also scans nested
// closures from their enclosing unit, reports each range only once.
type nullPointerException struct{}

// Semantic marks the rule as needing the project symbol index.
func (nullPointerException) Semantic()                            {}
func (nullPointerException) ID() string                           { return "NullPointerException" }
func (nullPointerException) Kinds() []syntax.NodeKind             { return nil }
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
			if f.Body == nil || semanticquery.InTestContext(ctx, f) {
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

func (u *npeUnit) strategyParams() {
	if u.body == nil { // D1
		return
	}
	for _, p := range syntax.FuncLikeParams(u.fn) { // D2
		// A variadic parameter is always an array: only its elements are nullable.
		if p.Var == nil || p.Var.Name == "" || p.Type == nil || p.Variadic {
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
					if astquery.Equivalent(u.ctx.File, e, mc.Var) {
						seen = true
						break
					}
				}
				if !seen {
					if tok, ok := astquery.NextSignificant(u.ctx.File, mc.Var.Span().End); ok && tok.Kind == syntax.TObjectOperator {
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
// null added by nullsafe short-circuiting does not count, even through
// intervening ordinary accesses; only the evaluated result's null/void does.
func (u *npeUnit) chainBaseType(x syntax.Expr) types.Type {
	t, _ := u.ctx.Types().ChainTypeOf(x)
	return t
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
	return astquery.LogicalOperand(call, false)
}

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
	tok, ok := astquery.TokenBefore(u.ctx.File, stmt.Span().Start)
	for ok && tok.Kind == syntax.TWhitespace {
		tok, ok = astquery.TokenBefore(u.ctx.File, tok.Start)
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

func npeVarsNamed(root syntax.Node, name string, strict bool, fn func(*syntax.Variable)) {
	syntax.Inspect(root, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name && !astquery.IsStaticPropName(v) {
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
		// U12: a variable whose native type is unknown is read as the
		// nullable object D5 found (strategy C); a declared one keeps its type.
		if rep && semanticquery.NativeNarrowedExcludes(u.ctx, v, nullableObject, "null") {
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
	if astquery.LogicalOperand(v, true) { // U3
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
				strings.EqualFold(astquery.LastNamePart(n.Value), "is_null") {
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
	switch call.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	default:
		return nil
	}
	c, ok := semanticquery.ResolveCallee(u.ctx, call)
	if !ok {
		return nil
	}
	if name != "" {
		for i := range c.Params {
			if c.Params[i].Name == name {
				return &c.Params[i]
			}
		}
		return nil
	}
	if idx >= len(c.Params) {
		return nil
	}
	return &c.Params[idx]
}
