package unnecessaryassertion

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/phpunit"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// unnecessaryAssertion reports assertions already guaranteed by a declared
// return type, and `expects($this->any())`.
type unnecessaryAssertion struct{}

func (unnecessaryAssertion) ID() string { return "UnnecessaryAssertion" }
func (unnecessaryAssertion) Semantic()  {}
func (unnecessaryAssertion) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

const (
	uaTypedMsg = "The declared return type already guarantees this; the assertion can go."
	uaAnyMsg   = "expects(any()) verifies nothing; drop the expects() call."
)

func (unnecessaryAssertion) Check(ctx *analysis.Context, n syntax.Node) {
	c, ok := phpunit.AsPuCall(n)
	if !ok {
		return
	}
	if strings.HasPrefix(strings.ToLower(c.Name), "assert") {
		uaCheckTyped(ctx, c)
		return
	}
	if c.Name != "expects" { // D7
		return
	}
	args, ok := c.Values()
	if !ok || len(args) != 1 {
		return
	}
	if _, ok := phpunit.PuMethodCallNamed(args[0], "any"); !ok { // D8
		return
	}
	// F1 needs a plain `->` call: a class reference is not an expression, a
	// nullsafe receiver would lose its short-circuit, and a doubled class
	// declaring its own method() makes the expects() form mandatory.
	mc, ok := c.Node.(*syntax.MethodCall)
	if !ok || mc.NullSafe || uaDeclaresMethod(ctx, c.Recv) {
		ctx.ReportNode(args[0], uaAnyMsg)
		return
	}
	call, recv := c.Node, c.Recv
	f := ctx.File
	ctx.ReportNode(args[0], uaAnyMsg, diagnostic.Fix{
		Title: "Drop expects(any())",
		Edits: func() []diagnostic.TextEdit {
			rs := recv.Span()
			return []diagnostic.TextEdit{{Span: call.Span(), NewText: string(f.Src[rs.Start:rs.End])}}
		},
	})
}

// uaDeclaresMethod reports whether a class the receiver may double (its
// inferred classes, plus every `X::class` named by the expressions it is
// built from) declares a real method named `method` outside PHPUnit:
// `$double->method(...)` would then call that method.
func uaDeclaresMethod(ctx *analysis.Context, recv syntax.Expr) bool {
	classes := ctx.TypeOf(recv).Classes()
	for _, v := range flowquery.PossibleValues(ctx.File, recv) {
		syntax.Inspect(v, func(n syntax.Node) bool {
			if e, ok := n.(syntax.Expr); ok {
				if cn, ok := phpunit.PuClassConstClass(e); ok {
					classes = append(classes, phpunit.PuResolveClassName(ctx, cn))
				}
			}
			return true
		})
	}
	for _, cls := range classes {
		m := ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), "method", ctx.PHP)
		if m != nil && !m.Magic && !strings.HasPrefix(strings.ToLower(strings.TrimPrefix(m.Class, `\`)), `phpunit\`) {
			return true
		}
	}
	return false
}

// uaValues is flowquery.PossibleValues where a local variable that the
// scope also rebinds other than by a plain assignment makes the result
// unknown (nil).
func uaValues(ctx *analysis.Context, e syntax.Expr) []syntax.Expr {
	w := flowquery.NewValueWalk(ctx.File)
	w.Resolve = func(e syntax.Expr) bool {
		v, ok := e.(*syntax.Variable)
		if !ok {
			vals, known := flowquery.PossibleValuesKnown(ctx.File, e)
			w.Stop = !known
			w.Out = append(w.Out, vals...)
			return true
		}
		if uaRebound(ctx, v) {
			w.Stop = true
			return true
		}
		w.LocalVar(v)
		return true
	}
	w.Collect(e)
	vals, _ := w.Result()
	return vals
}

// uaRebound reports whether the scope of v may give it a value that the
// plain-assignment scan does not see: foreach/catch/global/static/unset
// targets, destructuring, array element writes, references taken to it and
// by-reference (or unresolvable) argument positions.
func uaRebound(ctx *analysis.Context, v *syntax.Variable) bool {
	for _, a := range flowquery.VarAccesses(ctx.File, syntax.EnclosingFuncLike(v), v.Name) {
		if a.Write {
			// Compound assignments and ++/-- are left to the shared scan.
			switch by := a.By.(type) {
			case *syntax.Assign:
				if by.Var != syntax.Expr(a.Var) {
					return true
				}
			case *syntax.IncDec:
			default:
				return true
			}
			continue
		}
		if u, ok := a.By.(*syntax.ClosureUse); ok && u.ByRef {
			return true
		}
		switch p := a.Var.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if a.ElemWrite {
				return true
			}
		case *syntax.Assign:
			if p.ByRef && p.Value == syntax.Expr(a.Var) {
				return true
			}
		case *syntax.ArrayItem:
			if p.ByRef {
				return true
			}
		case *syntax.Arg:
			if uaArgMayBindByRef(ctx, p) {
				return true
			}
		}
	}
	return false
}

// uaArgMayBindByRef reports whether the call owning arg takes an argument
// by reference, or cannot be resolved to tell. Unresolved `assert*` methods
// are taken as by-value: the assertion itself is never resolved.
func uaArgMayBindByRef(ctx *analysis.Context, arg *syntax.Arg) bool {
	list := arg.Parent().(*syntax.ArgList)
	call := list.Parent()
	if n, ok := call.(*syntax.New); ok {
		nm, ok := n.Class.(*syntax.Name)
		if !ok {
			return true
		}
		cls := phpunit.PuResolveClassName(ctx, nm)
		if ctx.Index().Class(cls, ctx.PHP) == nil {
			return true
		}
		ctor := ctx.Index().FindMethod(cls, "__construct", ctx.PHP)
		return ctor != nil && semanticquery.ArgBindsByRef(list, ctor.Params, true)
	}
	if callee, ok := semanticquery.ResolveCallee(ctx, call); ok {
		return semanticquery.ArgBindsByRef(list, callee.Params, true)
	}
	c, ok := phpunit.AsPuCall(call)
	return !ok || !strings.HasPrefix(strings.ToLower(c.Name), "assert")
}

func uaCheckTyped(ctx *analysis.Context, c phpunit.PuCall) {
	if ctx.PHP < phpversion.PHP70 { // E1
		return
	}
	pos := 0 // D1
	switch c.Name {
	case "assertNull", "assertEmpty":
	case "assertInstanceOf", "assertInternalType":
		pos = 1
	default:
		return
	}
	args, ok := c.Values()
	if !ok || len(args) < pos+1 {
		return
	}
	vals := uaValues(ctx, args[pos]) // D2
	if len(vals) != 1 {
		return
	}
	if uaNullSafeChain(vals[0]) { // E4b (custos)
		return
	}
	t, ok := uaDeclaredReturn(ctx, vals[0]) // D3–D5
	if !ok || len(t.Atoms()) != 1 {
		return
	}
	atom := t.Atoms()[0]
	switch c.Name { // D6
	case "assertNull", "assertEmpty":
		if !strings.EqualFold(strings.TrimPrefix(atom, `\`), "void") {
			return
		}
	case "assertInstanceOf":
		cn, ok := phpunit.PuClassConstClass(args[0])
		if !ok {
			return
		}
		fqn := phpunit.PuResolveClassName(ctx, cn)
		if fqn == "" || ctx.Index().Class(fqn, ctx.PHP) == nil {
			return
		}
		if !strings.HasPrefix(atom, `\`) || !strings.EqualFold(atom[1:], fqn) {
			return
		}
	case "assertInternalType":
		lit, ok := args[0].(*syntax.Literal)
		if !ok {
			return
		}
		name, ok := astquery.StringLiteralValue(lit.Raw)
		if !ok || !uaInternalTypeMatches(name, atom) {
			return
		}
	}
	ctx.ReportNode(c.Node, uaTypedMsg)
}

// uaNullSafeChain reports whether a nullsafe access (`?->`) anywhere in
// e's member-access chain can short-circuit the whole chain to null, whatever
// the final method declares (`$r->find()?->sidebar()`).
func uaNullSafeChain(e syntax.Expr) bool {
	for {
		switch x := e.(type) {
		case *syntax.MethodCall:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.PropertyFetch:
			if x.NullSafe {
				return true
			}
			e = x.Var
		case *syntax.ArrayDimFetch:
			e = x.Var
		default:
			return false
		}
	}
}

// uaDeclaredReturn resolves a function/method call to its declaration and
// returns its declared return type merged with the @return doc type. ok is
// false when the call does not resolve or has no native return type.
func uaDeclaredReturn(ctx *analysis.Context, e syntax.Expr) (types.Type, bool) {
	var ret, doc, cls, self string
	switch x := e.(type) {
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(x)
		if f == nil || uaFirstClassCallable(x.Args) {
			return types.Unknown, false
		}
		ret, doc = f.Return, f.DocReturn
	case *syntax.MethodCall, *syntax.StaticCall:
		c, ok := phpunit.AsPuCall(x)
		if !ok || uaFirstClassCallable(c.Args) {
			return types.Unknown, false
		}
		if mc, ok := x.(*syntax.MethodCall); ok {
			cs := ctx.TypeOf(mc.Var).Classes()
			if len(cs) != 1 {
				return types.Unknown, false
			}
			cls = strings.TrimPrefix(cs[0], `\`)
		} else if nm, ok := c.Recv.(*syntax.Name); ok {
			cls = phpunit.PuResolveClassName(ctx, nm)
		} else {
			cs := ctx.TypeOf(c.Recv).Classes()
			if len(cs) != 1 {
				return types.Unknown, false
			}
			cls = strings.TrimPrefix(cs[0], `\`)
		}
		if cls == "" {
			return types.Unknown, false
		}
		m := ctx.Index().FindMethod(cls, c.Name, ctx.PHP)
		if m == nil {
			return types.Unknown, false
		}
		ret, doc = m.Return, m.DocReturn
		// `self` names the declaring class, or the class importing the
		// declaring trait; a trait reached without an importing class
		// leaves it unresolved.
		self = strings.TrimPrefix(m.Class, `\`)
		if m.TypeClass != "" {
			self = strings.TrimPrefix(m.TypeClass, `\`)
		} else if d := ctx.Index().Class(self, ctx.PHP); d != nil && d.Kind == syntax.KindTrait {
			self = ""
		}
	default:
		return types.Unknown, false
	}
	if ret == "" { // D4
		return types.Unknown, false
	}
	t := types.FromDoc(ret, nil)
	if doc != "" {
		t = types.Union(t, types.FromDoc(doc, nil))
	}
	// An unparsable type has no atoms: the caller's single-type check (D5)
	// rejects it.
	if cls != "" {
		atoms := make([]string, 0, len(t.Atoms()))
		for _, a := range t.Atoms() {
			switch {
			case a == "static":
				a = `\` + cls
			case a == "self" && self != "":
				a = `\` + self
			}
			atoms = append(atoms, a)
		}
		t = types.Of(atoms...)
	}
	return t, true
}

// uaFirstClassCallable reports whether list is the `(...)` of a first-class
// callable: such an expression yields a Closure, not the call's result.
func uaFirstClassCallable(list *syntax.ArgList) bool {
	if list == nil || len(list.Args) != 1 {
		return false
	}
	_, ok := list.Args[0].(*syntax.VariadicPlaceholder)
	return ok
}

// uaInternalTypeMatches reports whether a value of the single declared type
// atom always satisfies assertInternalType(name, ...).
func uaInternalTypeMatches(name, atom string) bool {
	atom = strings.ToLower(atom)
	isClass := strings.HasPrefix(atom, `\`)
	isArray := atom == "array" || strings.HasSuffix(atom, "[]")
	switch strings.ToLower(name) {
	case "array":
		return isArray
	case "bool", "boolean":
		return atom == "bool" || atom == "true" || atom == "false"
	case "float", "double", "real":
		return atom == "float"
	case "int", "integer":
		return atom == "int"
	case "numeric":
		return atom == "int" || atom == "float"
	case "string":
		return atom == "string"
	case "scalar":
		switch atom {
		case "int", "float", "string", "bool", "true", "false":
			return true
		}
	case "null":
		return atom == "null" || atom == "void"
	case "object":
		return (isClass && !isArray) || atom == "object"
	// "resource": no declared return type is a resource (`: resource`
	// names a class), so it is never guaranteed.
	case "callable":
		return atom == "callable" || atom == `\closure`
	case "iterable":
		return atom == "iterable" || isArray
	}
	return false
}
