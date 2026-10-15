package returntypecanbedeclared

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/types"
)

// returnTypeCanBeDeclared suggests a native return type for methods whose
// returned values and @return documentation agree on one declarable type.
type returnTypeCanBeDeclared struct{}

func (returnTypeCanBeDeclared) ID() string               { return "ReturnTypeCanBeDeclared" }
func (returnTypeCanBeDeclared) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }
func (returnTypeCanBeDeclared) Semantic()                {}

// rtdMagic holds the magic method names, lower-case (PHP method names are
// case-insensitive).
var rtdMagic = map[string]bool{
	"__construct": true, "__destruct": true, "__call": true, "__callstatic": true, "__get": true,
	"__set": true, "__isset": true, "__unset": true, "__sleep": true, "__wakeup": true,
	"__tostring": true, "__invoke": true, "__set_state": true, "__clone": true, "__debuginfo": true,
	"__serialize": true, "__unserialize": true, // custos: mandated return types
}

var rtdBuiltin = map[string]bool{
	"array": true, "iterable": true, "string": true, "bool": true, "int": true, "float": true,
	"number": true, "null": true, "void": true, "mixed": true, "callable": true, "resource": true,
	"static": true, "self": true, "object": true, "never": true, "parent": true,
}

var rtdScalarOK = map[string]bool{
	"self": true, "array": true, "callable": true, "bool": true, "float": true, "int": true, "string": true,
}

// rtdNormalize implements D7. Type atoms are already normalised by the
// type model (integer -> int, boolean -> bool, $this -> static).
func rtdNormalize(a string) string {
	low := strings.ToLower(a)
	switch {
	case strings.Contains(a, "[]"):
		return "array"
	case low == "boolean" || low == "true" || low == "false":
		return "bool"
	case low == `\closure` || low == "closure":
		return "callable"
	}
	if rtdBuiltin[strings.TrimPrefix(low, `\`)] {
		return strings.TrimPrefix(low, `\`)
	}
	return `\` + strings.TrimPrefix(a, `\`)
}

// rtdWalkOwn visits the nodes of a method body that belong to the method
// itself (nested functions, closures and classes are skipped).
func rtdWalkOwn(body syntax.Node, fn func(syntax.Node)) {
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		}
		fn(n)
		return true
	})
}

// rtdFirstReturn returns the first return statement in document order at any
// depth (nested closures included).
func rtdFirstReturn(body syntax.Node) *syntax.Return {
	var first *syntax.Return
	syntax.Inspect(body, func(n syntax.Node) bool {
		if first != nil {
			return false
		}
		if r, ok := n.(*syntax.Return); ok {
			first = r
			return false
		}
		return true
	})
	return first
}

// rtdExprType is the inferred type of a returned expression, with `$this`
// and `new static` typed as static.
func rtdExprType(ctx *analysis.Context, e syntax.Expr) types.Type {
	return rtdExprTypeIn(ctx.Types(), e)
}

// rtdExprTypeIn is rtdExprType in env (the native env ignores PHPDoc).
func rtdExprTypeIn(env *infer.Env, e syntax.Expr) types.Type {
	switch x := e.(type) {
	case *syntax.Paren:
		return rtdExprTypeIn(env, x.Expr)
	case *syntax.Variable:
		if x.Name == "this" {
			return types.Of("static")
		}
	case *syntax.New:
		if n, ok := x.Class.(*syntax.Name); ok && strings.EqualFold(n.Value, "static") {
			return types.Of("static")
		}
	}
	return env.TypeOf(e)
}

func (r returnTypeCanBeDeclared) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 { // E1
		return
	}
	m := n.(*syntax.Method)
	class, ok := m.Parent().(*syntax.ClassLike)
	if !ok || m.Name == nil || m.ReturnType != nil || rtdMagic[strings.ToLower(m.Name.Value)] { // D1-D3
		return
	}
	// custos: constructors take no return type below 8.0; from 8.0
	// DeprecatedConstructorStyle renames such a method to __construct, and
	// both fixes together would give `__construct(): void` (fatal).
	if rtdLegacyConstructor(class, m) {
		return
	}
	abstract := m.Body == nil
	var doc *phpdoc.Doc
	if c := index.DocComment(ctx.File, m); c != "" {
		doc = phpdoc.Parse(c)
	}
	hasReturnTag := false
	if doc != nil {
		_, hasReturnTag = doc.Tag("return")
	}
	if abstract && !hasReturnTag { // D4
		return
	}

	// D5: R0.
	known := map[string]bool{}
	unknown := false
	add := func(t types.Type) {
		if t.IsUnknown() {
			unknown = true
			return
		}
		for _, a := range t.Atoms() {
			known[a] = true
		}
	}
	at := m.Span().Start
	resolve := func(w string) string { return ctx.Names().Class(w, at) }
	docRet := ""
	if hasReturnTag {
		docRet = doc.ReturnType()
		add(types.FromDoc(docRet, resolve))
	}
	hasYield := false
	if !abstract {
		rtdWalkOwn(m.Body, func(x syntax.Node) {
			switch x.(type) {
			case *syntax.Yield, *syntax.YieldFrom:
				hasYield = true
			}
		})
	}
	unknownReturn := false                  // a `return expr;` of unknown type
	docReturn := false                      // a `return expr;` typed through PHPDoc
	bareReturn, valueReturn := false, false // own `return;` / `return expr;`
	var vars *rtdVars                       // computed on the first `return expr;`
	if !abstract {
		rtdWalkOwn(m.Body, func(x syntax.Node) {
			if x, ok := x.(*syntax.Return); ok {
				if hasYield {
					// custos: a generator's return value is not what the
					// call returns (always a Generator object).
					return
				}
				if x.Expr == nil {
					bareReturn = true
					add(types.Void)
				} else {
					valueReturn = true
					t := rtdExprType(ctx, x.Expr)
					if vars == nil {
						vars = rtdScanVars(ctx, class, m)
					}
					// custos: a value that a reference or a subclass constant
					// may replace is unknown, whatever its initialiser says.
					unstable, power := rtdOperandFacts(ctx, class, vars, x.Expr)
					if t.Has("mixed") || unstable {
						// custos: `mixed` (Doctrine's getResult()) says no more
						// than an unknown value: the @return tag decides.
						t = types.Type{}
					}
					if t.IsUnknown() && !unstable {
						t = rtdInheritedParamType(ctx, class, m, x.Expr)
					}
					if power {
						add(types.Of("float")) // custos: int ** int may be a float
					}
					if rtdImplicitNullProp(ctx, class, x.Expr) { // D5b
						if t.IsUnknown() {
							t = types.Null // the @return tag covers the written values
						} else {
							add(types.Null)
						}
					}
					if t.IsUnknown() {
						unknownReturn = true
					} else if !rtdExprTypeIn(ctx.Types().Native(), x.Expr).Equal(t) {
						// custos: the type rests on PHPDoc somewhere (a
						// callee's @return, a @param, a property's @var).
						docReturn = true
					}
					add(t)
				}
			}
		})
	}
	// D6. Without a @return tag, the "one known member" allowance does not
	// cover returned values (custos diverges: an unknown returned value may
	// be anything, so declaring the other member's type would throw or
	// coerce).
	if unknownReturn && !hasReturnTag || unknown && len(known) != 1 {
		return
	}
	set := map[string]bool{} // D7
	for a := range known {
		// An anonymous identity supports lookup, but cannot be written in PHP.
		if names.IsAnonymousClassName(a) {
			return
		}
		set[rtdNormalize(a)] = true
		// custos: Doctrine hydrates mapped collections as
		// PersistentCollection and matching() returns a lazy collection;
		// neither is an ArrayCollection, so `: ArrayCollection` throws.
		if strings.EqualFold(rtdNormalize(a), `\Doctrine\Common\Collections\ArrayCollection`) {
			return
		}
	}
	if hasYield && !set[`\Generator`] { // D8
		set[`\Generator`] = true
		if rtdFirstReturn(m.Body) == nil {
			delete(set, "null")
		}
	}
	if len(set) > 0 && !abstract && !hasYield { // D9 (a generator call never returns null)
		if !set["null"] && !set["void"] {
			last := syntax.Stmt(nil)
			if len(m.Body.Stmts) > 0 {
				last = m.Body.Stmts[len(m.Body.Stmts)-1]
			}
			isExit := false
			switch l := last.(type) {
			case *syntax.Return:
				isExit = true
			case *syntax.ExprStmt:
				_, isExit = l.Expr.(*syntax.Throw)
			}
			if !isExit {
				set["null"] = true
			}
		}
		if len(set) == 1 && set["null"] {
			if fr := rtdFirstReturn(m.Body); fr != nil && fr.Expr != nil && !syntax.IsNullConst(fr.Expr) {
				delete(set, "null")
			}
		}
	}

	l71 := ctx.PHP >= phpversion.PHP71
	suggestion := ""
	switch len(set) {
	case 0: // D10
		if !l71 {
			return
		}
		if !abstract {
			fr := rtdFirstReturn(m.Body)
			if fr != nil && syntax.EnclosingFuncLike(fr) != syntax.Node(m) {
				fr = nil
			}
			if fr != nil {
				return
			}
		}
		suggestion = "void"
	case 1: // D11
		var t string
		for a := range set {
			t = a
		}
		s := ""
		if l71 && (t == "null" || t == "void") {
			s = "void"
		} else {
			s = r.compact(ctx, class, doc, docRet, t)
		}
		switch {
		case rtdTrait(ctx, t):
			return
		case strings.HasPrefix(t, `\`) || rtdScalarOK[t] || s == "self" || s == "static":
		case l71 && s == "void":
		default:
			return
		}
		if hasReturnTag {
			if tag, _ := doc.Tag("return"); strings.TrimSpace(tag.Text) == "static" { // static guard
				if ctx.PHP < phpversion.PHP80 {
					return
				}
				s = "static"
			}
		}
		if s == "static" && ctx.PHP < phpversion.PHP80 {
			return // see Divergences: never suggest static below 8.0
		}
		suggestion = s
	case 2: // D12
		if !l71 {
			return
		}
		if set["void"] {
			delete(set, "void")
		} else if set["null"] {
			delete(set, "null")
		}
		if len(set) != 1 {
			return
		}
		var t string
		for a := range set {
			t = a
		}
		s := ""
		if t == "null" || t == "void" {
			s = "void"
		} else {
			s = r.compact(ctx, class, doc, docRet, t)
		}
		switch {
		case rtdTrait(ctx, t):
			return
		case strings.HasPrefix(t, `\`) || rtdScalarOK[t] || s == "self":
			suggestion = "?" + s
		case s == "void":
			suggestion = "void"
		default:
			return
		}
	default:
		return
	}

	// D14: the declaration must accept the method's own return statements:
	// `return null;` is a compile error under `: void`, and a bare `return;`
	// under any other type (generators excepted).
	if !hasYield && (suggestion == "void" && valueReturn || suggestion != "void" && bareReturn) {
		return
	}
	span := m.Name.Span()
	if rtdOverridden(ctx, class, m) { // D13
		ctx.Report(span, "Declare ': "+suggestion+"' as the return type (update the whole hierarchy with a signature refactoring).")
		return
	}
	pos, ok := rtdParamsEnd(ctx, m)
	if !ok || unknownReturn || docReturn || rtdTraitExposed(class, m) {
		// custos: a returned value of unknown type means the suggestion
		// rests on the @return tag alone; a wrong doc would make the added
		// native type throw, so it is reported without a fixing. Nor is a
		// trait method fixed when the subclasses of the classes using the
		// trait (not enumerable) may override it.
		ctx.Report(span, "Declare ': "+suggestion+"' as the return type.")
		return
	}
	text := ": " + suggestion
	if int(pos) < len(ctx.Src) && ctx.Src[pos] == '{' {
		text += " "
	}
	ctx.Report(span, "Declare ': "+suggestion+"' as the return type.", diagnostic.Fix{
		Title: "Declare the return type",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: syntax.Span{Start: pos, End: pos}, NewText: text}}
		},
	})
}

// compact implements compact(t).
func (returnTypeCanBeDeclared) compact(ctx *analysis.Context, class *syntax.ClassLike, doc *phpdoc.Doc, docRet, t string) string {
	if !strings.HasPrefix(t, `\`) && t != "static" {
		return t
	}
	if docRet != "" && ctx.Bool("LOOKUP_PHPDOC_RETURN_DECLARATIONS") { // 1
		for _, p := range strings.Split(docRet, "|") {
			if p = strings.TrimSpace(p); p == "self" || p == "$this" {
				return "self"
			}
		}
	}
	if t == "static" { // 2
		return "static"
	}
	// 3: imports in the enclosing statement lists.
	for p := syntax.Node(class); ; {
		parent := p.Parent()
		var list []syntax.Stmt
		switch x := parent.(type) {
		case *syntax.File:
			list = x.Stmts
		case *syntax.Namespace:
			list = x.Stmts
		case *syntax.Block:
			list = x.Stmts
		case nil:
			list = ctx.File.Stmts
		}
		for _, s := range list {
			u, ok := s.(*syntax.Use)
			if !ok {
				continue
			}
			for _, it := range u.Items {
				kind := u.Type
				if u.Prefix != nil {
					kind = it.Type
				}
				if kind != syntax.UseNormal {
					continue
				}
				name := strings.TrimPrefix(it.Name.Value, `\`)
				if u.Prefix != nil {
					name = strings.TrimPrefix(strings.TrimSuffix(u.Prefix.Value, `\`), `\`) + `\` + name
				}
				if `\`+name == t {
					if it.Alias != nil {
						return it.Alias.Value
					}
					return astquery.LastNamePart(name)
				}
			}
		}
		if parent == nil {
			break
		}
		p = parent
	}
	// 4: same namespace prefix.
	if ns := ctx.Names().Namespace(class.Span().Start); ns != "" && strings.HasPrefix(t, `\`+ns+`\`) {
		return t[len(ns)+2:]
	}
	return t
}

// rtdTrait reports whether t names a trait: no value is ever an instance
// of a trait, so `: Singleton` always throws (custos; a doc type may still
// name one, though `new static` / `clone $this` in a trait are unknown).
func rtdTrait(ctx *analysis.Context, t string) bool {
	if !strings.HasPrefix(t, `\`) {
		return false
	}
	c := ctx.Index().Class(t, ctx.PHP)
	return c != nil && c.Kind == syntax.KindTrait
}

// rtdOverridden implements D13.
func rtdOverridden(ctx *analysis.Context, class *syntax.ClassLike, m *syntax.Method) bool {
	if class.ClassKind == syntax.KindTrait && m.Body == nil {
		return true // every class using the trait declares the method
	}
	if class.Modifiers.Has(syntax.TFinal) || m.Modifiers.Has(syntax.TFinal) || m.Modifiers.Has(syntax.TPrivate) {
		return false
	}
	fqn := ctx.Types().ClassFQN(class)
	ix := ctx.Index()
	name := strings.ToLower(m.Name.Value)
	self := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	for _, c := range ix.Ancestors(fqn, ctx.PHP) {
		if strings.ToLower(strings.TrimPrefix(c.FQN, `\`)) == self {
			continue
		}
		if _, ok := c.Methods[name]; ok {
			return true
		}
	}
	if class.ClassKind == syntax.KindTrait {
		return rtdTraitUserDeclares(ctx, self, name)
	}
	below := ctx.Memo("descendant-methods\x00"+self, func() any {
		return semanticquery.DescendantMethods(ix, fqn, ctx.PHP)
	}).(map[string]bool)
	return below[name]
}

// rtdTraitExposed reports whether trait method m is composed into classes
// whose subclasses may override it: the index does not list the classes
// using a trait, so such a method never gets a fix.
func rtdTraitExposed(class *syntax.ClassLike, m *syntax.Method) bool {
	return class.ClassKind == syntax.KindTrait && !m.Modifiers.Has(syntax.TFinal) && !m.Modifiers.Has(syntax.TPrivate)
}

// rtdTraitUserDeclares reports whether a class-like of the current file
// that composes trait (lower-case FQN, no leading backslash), or inherits
// from one that does, declares method name itself, or a method named like
// an alias a `use` block gives to name.
func rtdTraitUserDeclares(ctx *analysis.Context, trait, name string) bool {
	ix := ctx.Index()
	var users []*index.Class
	wanted := map[string]bool{name: true}
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		cl, ok := n.(*syntax.ClassLike)
		if !ok {
			return true
		}
		fqn := ctx.Types().ClassFQN(cl)
		c := ix.Class(fqn, ctx.PHP)
		if c == nil || strings.ToLower(strings.TrimPrefix(c.FQN, `\`)) == trait {
			return true
		}
		for _, a := range ix.Ancestors(fqn, ctx.PHP) {
			if strings.ToLower(strings.TrimPrefix(a.FQN, `\`)) != trait {
				continue
			}
			users = append(users, c)
			for _, ad := range c.TraitAdaptations {
				if ad.Alias != "" && strings.ToLower(ad.Method) == name {
					wanted[strings.ToLower(ad.Alias)] = true
				}
			}
			break
		}
		return true
	})
	for _, c := range users {
		for w := range wanted {
			if _, ok := c.Methods[w]; ok {
				return true
			}
		}
	}
	return false
}

// Levels of rtdVars.refs.
const (
	rtdMaybeRef uint8 = iota + 1 // passed to a callee that cannot be resolved
	rtdRef                       // bound by reference
)

// rtdVars describes the variables of a method whose inferred type cannot
// be taken at face value.
type rtdVars struct {
	refs   map[string]uint8 // may be written by something else than a tracked assignment
	powers map[string]bool  // assigned an integer power (D5d)
}

// rtdScanVars collects the rtdVars of m. refs: variables bound by reference
// (`=&`, `[&$v]`, by-reference foreach value, closure `use (&$v)`,
// `global`, `static`), passed to a by-reference parameter of a user callee
// or to settype(), or passed to a callee that cannot be resolved. Builtin
// functions are left to the type inference, which knows their output
// parameters.
func rtdScanVars(ctx *analysis.Context, class *syntax.ClassLike, m *syntax.Method) *rtdVars {
	out := map[string]uint8{}
	powers := map[string]bool{}
	mark := func(e syntax.Expr, level uint8) {
		for {
			d, ok := e.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			e = d.Var // `f($v['k'])` may turn a null $v into an array
		}
		if v, ok := e.(*syntax.Variable); ok && v.Name != "" && out[v.Name] < level {
			out[v.Name] = level
		}
	}
	syntax.Inspect(m.Body, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Function, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Closure:
			for _, u := range x.Uses {
				if u.ByRef {
					mark(u.Var, rtdRef)
				}
			}
			return false
		case *syntax.Assign:
			if x.ByRef {
				mark(x.Var, rtdRef)
				mark(x.Value, rtdRef)
			}
			if v, ok := x.Var.(*syntax.Variable); ok && rtdAssignsPower(x) {
				powers[v.Name] = true
			}
		case *syntax.ArrayItem:
			if x.ByRef {
				mark(x.Value, rtdRef)
			}
		case *syntax.Foreach:
			if x.ByRef {
				mark(x.Value, rtdRef)
			}
		case *syntax.Global:
			for _, v := range x.Vars {
				mark(v, rtdRef)
			}
		case *syntax.StaticVar:
			mark(x.Var, rtdRef)
		case *syntax.FuncCall:
			if f := ctx.Types().ResolveFunction(x); f == nil {
				rtdMarkArgs(x.Args, nil, false, mark)
			} else if !f.Builtin {
				rtdMarkArgs(x.Args, f.Params, true, mark)
			} else if strings.EqualFold(strings.TrimPrefix(f.FQN, `\`), "settype") {
				rtdMarkArgs(x.Args, f.Params, true, mark)
			}
		case *syntax.MethodCall:
			c, ok := semanticquery.ResolveCallee(ctx, x)
			rtdMarkArgs(x.Args, c.Params, ok, mark)
		case *syntax.StaticCall:
			c, ok := semanticquery.ResolveCallee(ctx, x)
			rtdMarkArgs(x.Args, c.Params, ok, mark)
		case *syntax.New:
			if x.Args != nil {
				params, ok := rtdCtorParams(ctx, x)
				rtdMarkArgs(x.Args, params, ok, mark)
			}
		}
		return true
	})
	return &rtdVars{refs: out, powers: powers}
}

// rtdAssignsPower reports a compound power whose exponent may be negative.
// Ordinary power expressions already retain float in their inferred type.
func rtdAssignsPower(a *syntax.Assign) bool {
	return a.Op.Kind == syntax.TPowEqual && !rtdNaturalLiteral(a.Value)
}

// rtdNaturalLiteral reports whether e is a non-negative integer literal (a
// negative one is a unary minus).
func rtdNaturalLiteral(e syntax.Expr) bool {
	lit, ok := e.(*syntax.Literal)
	return ok && lit.LitKind == syntax.LitInt
}

// rtdCtorParams resolves the constructor parameters of a `new` expression;
// false when the class or part of its hierarchy is unknown.
func rtdCtorParams(ctx *analysis.Context, n *syntax.New) ([]index.Param, bool) {
	cls := semanticquery.StaticCallClass(ctx, n.Class)
	if cls == "" || ctx.Index().Class(cls, ctx.PHP) == nil || !semanticquery.HierarchyResolved(ctx.Index(), cls, ctx.PHP) {
		return nil, false
	}
	if c := ctx.Index().FindMethod(cls, "__construct", ctx.PHP); c != nil {
		return c.Params, true
	}
	return nil, true
}

// rtdMarkArgs marks the arguments of a call bound by reference (all of
// them, one level lower, when the callee is not resolved).
func rtdMarkArgs(al *syntax.ArgList, params []index.Param, resolved bool, mark func(syntax.Expr, uint8)) {
	for i, x := range al.Args {
		a, ok := x.(*syntax.Arg)
		if !ok || a.Unpack {
			continue
		}
		switch {
		case a.ByRef:
			mark(a.Value, rtdRef)
		case !resolved:
			mark(a.Value, rtdMaybeRef)
		case rtdParamByRef(params, a, i):
			mark(a.Value, rtdRef)
		}
	}
}

// rtdParamByRef reports whether argument a (at position i) binds a
// by-reference parameter.
func rtdParamByRef(params []index.Param, a *syntax.Arg, i int) bool {
	if a.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), a.Name.Value) {
				return p.ByRef
			}
		}
		return false
	}
	if i >= len(params) {
		i = len(params) - 1 // a variadic takes the remaining arguments
		if i < 0 || !params[i].Variadic {
			return false
		}
	}
	return params[i].ByRef
}

// rtdOperands calls fn on e and on every sub-expression whose type decides
// the type of e (the branches of a conditional, arithmetic operands, the
// assigned value, the array an element is read from).
func rtdOperands(e syntax.Expr, fn func(syntax.Expr)) {
	fn(e)
	switch x := e.(type) {
	case *syntax.Paren:
		rtdOperands(x.Expr, fn)
	case *syntax.Ternary:
		if x.Then != nil {
			rtdOperands(x.Then, fn)
		} else {
			rtdOperands(x.Cond, fn)
		}
		rtdOperands(x.Else, fn)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TCoalesce, syntax.TPlus, syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow:
			rtdOperands(x.Left, fn)
			rtdOperands(x.Right, fn)
		}
	case *syntax.Unary:
		switch x.Op.Kind {
		case syntax.TPlus, syntax.TMinus, syntax.TTilde, syntax.TAt:
			rtdOperands(x.Expr, fn)
		}
	case *syntax.Assign:
		if x.Op.Kind == syntax.TEqual {
			rtdOperands(x.Value, fn)
		}
	case *syntax.ArrayDimFetch:
		rtdOperands(x.Var, fn)
	case *syntax.Match:
		for _, arm := range x.Arms {
			rtdOperands(arm.Body, fn)
		}
	}
}

// rtdOperandFacts inspects the operands of expression e. unstable: one of
// them is a variable of vars.refs whose value may have been replaced, or a
// late-bound class constant a subclass redeclares. power: one of them is
// an integer `**` (or a variable assigned one), which gives a float for a
// negative exponent.
func rtdOperandFacts(ctx *analysis.Context, class *syntax.ClassLike, vars *rtdVars, e syntax.Expr) (unstable, power bool) {
	rtdOperands(e, func(o syntax.Expr) {
		switch x := o.(type) {
		case *syntax.Variable:
			power = power || vars.powers[x.Name]
			switch vars.refs[x.Name] {
			case rtdRef:
				unstable = true
			case rtdMaybeRef:
				// An object handed to an unknown callee stays that object
				// in practice; scalars and arrays are what callees fill.
				t := ctx.TypeOf(x)
				if len(t.Classes()) == 0 || len(t.Classes()) != len(t.Without("null").Atoms()) {
					unstable = true
				}
			}
		case *syntax.ClassConstFetch:
			if rtdLateConst(ctx, class, x) {
				unstable = true
			}
		}
	})
	return unstable, power
}

// rtdLateConst reports whether c reads a constant through `static::` or
// `$this::` that a known subclass redeclares (always in a trait, whose
// using classes are not known): the value is then not the one declared
// here. A final or typed constant keeps its type.
func rtdLateConst(ctx *analysis.Context, class *syntax.ClassLike, c *syntax.ClassConstFetch) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || strings.EqualFold(id.Value, "class") {
		return false
	}
	switch x := c.Class.(type) {
	case *syntax.Name:
		if !strings.EqualFold(x.Value, "static") {
			return false
		}
	case *syntax.Variable:
		if x.Name != "this" {
			return false
		}
	default:
		return false
	}
	if class.ClassKind == syntax.KindTrait {
		return true
	}
	fqn := ctx.Types().ClassFQN(class)
	ix := ctx.Index()
	if k := ix.FindConst(fqn, id.Value, ctx.PHP); k != nil && (k.Final || k.Type != "") {
		return false
	}
	for _, sub := range ix.Subclasses(fqn) {
		if sc := ix.Class(sub, ctx.PHP); sc != nil && sc.Consts[id.Value] != nil {
			return true
		}
	}
	return false
}

// rtdParamsEnd returns the offset right after the `)` closing the parameter
// list of m.
func rtdParamsEnd(ctx *analysis.Context, m *syntax.Method) (uint32, bool) {
	end := m.Span().End
	if m.Body != nil {
		end = m.Body.Span().Start
	}
	toks := ctx.File.Tokens
	lo, hi := 0, len(toks)
	for lo < hi {
		mid := (lo + hi) / 2
		if toks[mid].Start < end {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	for i := lo - 1; i >= 0 && toks[i].Start >= m.Span().Start; i-- {
		if toks[i].Kind == syntax.TRParen {
			return toks[i].End, true
		}
	}
	return 0, false
}

// rtdInheritedParamType types a returned, untyped and undocumented parameter
// of m from the same parameter of the method it overrides (declared type,
// else doc type), as PhpStorm inherits parameter types.
func rtdInheritedParamType(ctx *analysis.Context, class *syntax.ClassLike, m *syntax.Method, e syntax.Expr) types.Type {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return types.Unknown
	}
	idx := -1
	for i, p := range m.Params {
		if p.Var != nil && p.Var.Name == v.Name && p.Type == nil {
			idx = i
		}
	}
	if idx < 0 {
		return types.Unknown
	}
	// Only when the variable is never reassigned in the method.
	reassigned := false
	rtdWalkOwn(m.Body, func(n syntax.Node) {
		if a, ok := n.(*syntax.Assign); ok {
			if t, ok := a.Var.(*syntax.Variable); ok && t.Name == v.Name {
				reassigned = true
			}
		}
	})
	if reassigned {
		return types.Unknown
	}
	fqn := ctx.Types().ClassFQN(class)
	self := strings.ToLower(strings.TrimPrefix(fqn, `\`))
	name := strings.ToLower(m.Name.Value)
	for _, c := range ctx.Index().Ancestors(fqn, ctx.PHP) {
		if strings.ToLower(strings.TrimPrefix(c.FQN, `\`)) == self {
			continue
		}
		pm, ok := c.Methods[name]
		if !ok || idx >= len(pm.Params) {
			continue
		}
		p := pm.Params[idx]
		if p.Type != "" {
			return types.FromDoc(p.Type, nil)
		}
		if p.DocType != "" {
			return types.FromDoc(p.DocType, nil)
		}
		return types.Unknown
	}
	return types.Unknown
}

// rtdImplicitNullProp implements D5b: e reads `$this->p` where p has no
// native type and no non-null default, so it holds null until written —
// unless p is a promoted parameter or this class's constructor assigns it
// in a top-level statement. (Its @var documentation does not prevent null:
// `@var int` on a nullable ORM column is the common case.)
func rtdImplicitNullProp(ctx *analysis.Context, class *syntax.ClassLike, e syntax.Expr) bool {
	pf, ok := syntax.UnwrapParens(e).(*syntax.PropertyFetch)
	if !ok || pf.NullSafe {
		return false
	}
	v, ok := pf.Var.(*syntax.Variable)
	id, ok2 := pf.Name.(*syntax.Identifier)
	if !ok || !ok2 || v.NameExpr != nil || v.Name != "this" {
		return false
	}
	fqn := strings.TrimPrefix(ctx.Types().ClassFQN(class), `\`)
	p := ctx.Index().FindProperty(fqn, id.Value, ctx.PHP)
	if p == nil || p.Type != "" || p.Static || p.Promoted || p.Magic {
		return false
	}
	if p.HasDefault && !strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(p.Default), `\`), "null") {
		return false
	}
	return !flowquery.CtorAssignsProperty(class, id.Value)
}

// rtdLegacyConstructor reports whether m is shaped like a PHP 4 style
// constructor (one below PHP 8.0): a method named like its class (outside
// a named namespace) in a class without __construct.
func rtdLegacyConstructor(class *syntax.ClassLike, m *syntax.Method) bool {
	if class.ClassKind != syntax.KindClass || class.Name == nil || !strings.EqualFold(m.Name.Value, class.Name.Value) || astquery.InNamedNamespace(class) {
		return false
	}
	for _, mem := range class.Members {
		if o, ok := mem.(*syntax.Method); ok && o.Name != nil && strings.EqualFold(o.Name.Value, "__construct") {
			return false
		}
	}
	return true
}
