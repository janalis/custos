package senselessmethodduplication

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// senselessMethodDuplication reports methods whose body copies the
// inherited implementation.
type senselessMethodDuplication struct{}

func (senselessMethodDuplication) ID() string               { return "SenselessMethodDuplication" }
func (senselessMethodDuplication) Semantic()                {}
func (senselessMethodDuplication) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

func (senselessMethodDuplication) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	cl, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cl.ClassKind != syntax.KindClass || m.Name == nil || m.Body == nil || m.Name.Span().Len() == 0 { // D3
		return
	}
	if m.Modifiers.Has(syntax.TAbstract) || m.Modifiers.Has(syntax.TPrivate) || smdDeprecated(ctx, m) { // D1
		return
	}
	ownFQN := ctx.Names().DeclFQN(cl)
	if ctx.IsTestFile() || astquery.IsTestClassFQN(ownFQN) { // D2
		return
	}
	stmts := astquery.MethodStatements(m.Body) // D4
	if len(stmts) == 0 || len(stmts) > ctx.Int("MAX_METHOD_SIZE") {
		return
	}
	ix := ctx.Index()
	parentFQN := ctx.Names().ParentFQN(cl)
	if parentFQN == "" {
		return
	}
	pm := semanticquery.MethodInChain(ix, parentFQN, m.Name.Value, ctx.PHP) // D5
	if pm == nil || pm.Abstract || pm.Deprecated || pm.Visibility == index.Private {
		return
	}
	if pm.Static != m.Modifiers.Has(syntax.TStatic) { // D5: staticness differs
		return
	}
	pmDecl := semanticquery.MethodDecl(ctx.File, ix, pm, ctx.PHP)
	if pmDecl == nil || pmDecl.Body == nil || smdDeprecated(ctx, pmDecl) {
		return // the parent body is only comparable within this file
	}
	pstmts := astquery.MethodStatements(pmDecl.Body) // D6
	if len(pstmts) != len(stmts) {
		return
	}
	for i := range stmts { // D7
		if !astquery.EquivalentFoldNames(ctx.File, stmts[i], pstmts[i]) {
			return
		}
	}
	mine, ok := smdSymbols(ctx, m.Body, ownFQN) // D8
	if !ok {
		return
	}
	theirs, ok := smdSymbols(ctx, pmDecl.Body, pm.Class)
	if !ok || len(mine) != len(theirs) {
		return
	}
	for k := range mine {
		if !theirs[k] {
			return
		}
	}
	if smdTouchesPrivate(ctx, pmDecl.Body, pm.Class) { // D9
		return
	}
	f := ctx.File
	name := m.Name.Value
	if smdVisibility(m.Modifiers) == pm.Visibility { // D10 I
		ctx.Report(m.Name.Span(), "Method '"+name+"' duplicates the inherited implementation; remove it.", diagnostic.Fix{
			Title: "Remove the method",
			Edits: func() []diagnostic.TextEdit { return astquery.MethodRemovalEdits(f, m) },
		})
		return
	}
	ctx.Report(m.Name.Span(), "Method '"+name+"' duplicates the inherited implementation; delegate to parent::"+name+"() instead.", diagnostic.Fix{
		Title: "Delegate to the parent method",
		Edits: func() []diagnostic.TextEdit { return smdDelegateEdits(ctx, m) },
	})
}

func smdVisibility(mods syntax.Modifiers) index.Visibility {
	if mods.Has(syntax.TProtected) { // private methods are skipped by D1
		return index.Protected
	}
	return index.Public
}

// smdDeprecated reports a @deprecated doc tag or a #[\Deprecated] attribute.
func smdDeprecated(ctx *analysis.Context, m *syntax.Method) bool {
	for _, g := range m.Attrs {
		for _, a := range g.Attrs {
			if a.Name != nil && strings.EqualFold(ctx.Names().Class(a.Name.Value, a.Span().Start), "Deprecated") {
				return true
			}
		}
	}
	return semanticquery.DocHasTag(ctx.File, m, "deprecated")
}

// smdSymbols collects the resolved class, constant and function references
// of a body (D8); ok is false when one does not resolve.
func smdSymbols(ctx *analysis.Context, body *syntax.Block, class string) (map[string]bool, bool) {
	ix := ctx.Index()
	out := map[string]bool{}
	ok := true
	classRef := func(name *syntax.Name) {
		var fqn string
		switch strings.ToLower(name.Value) {
		case "self", "static":
			fqn = class
		case "parent":
			if c := ix.Class(class, ctx.PHP); c != nil {
				fqn = c.Parent
			}
		default:
			fqn = ctx.Names().Class(name.Value, name.Span().Start)
		}
		c := ix.Class(fqn, ctx.PHP)
		if c == nil {
			ok = false
			return
		}
		out["c:"+strings.ToLower(strings.TrimPrefix(c.FQN, `\`))] = true
	}
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x := x.(type) {
		case *syntax.Name:
			switch p := x.Parent().(type) {
			case *syntax.New:
				if p.Class == syntax.Expr(x) {
					classRef(x)
				}
			case *syntax.StaticCall:
				if p.Class == syntax.Expr(x) {
					classRef(x)
				}
			case *syntax.ClassConstFetch:
				if p.Class == syntax.Expr(x) {
					classRef(x)
				}
			case *syntax.StaticPropertyFetch:
				if p.Class == syntax.Expr(x) {
					classRef(x)
				}
			case *syntax.Instanceof:
				if p.Class == syntax.Expr(x) {
					classRef(x)
				}
			case *syntax.Catch:
				classRef(x)
			}
		case *syntax.ConstFetch:
			switch low := strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`)); low {
			case "true", "false", "null":
				out["k:"+low] = true
				return true
			}
			fqn, fb := ctx.Names().Const(x.Name.Value, x.Name.Span().Start)
			switch {
			case ix.Constant(fqn, ctx.PHP) != nil:
				out["k:"+fqn] = true
			case fb != "" && ix.Constant(fb, ctx.PHP) != nil:
				out["k:"+fb] = true
			default:
				ok = false
			}
		case *syntax.FuncCall:
			if _, named := x.Name.(*syntax.Name); !named {
				return true
			}
			fn := ctx.Types().ResolveFunction(x)
			if fn == nil {
				ok = false
				return true
			}
			out["f:"+strings.ToLower(strings.TrimPrefix(fn.FQN, `\`))] = true
		}
		return true
	})
	return out, ok
}

// smdTouchesPrivate reports $this->/$this:: accesses in body resolving to a
// private member, seen from class (D9). Literal self:: accesses never get
// here: D8 already resolves self to a different class in child and parent.
func smdTouchesPrivate(ctx *analysis.Context, body *syntax.Block, class string) bool {
	ix := ctx.Index()
	isThis := func(e syntax.Expr) bool {
		v, ok := e.(*syntax.Variable)
		return ok && v.Name == "this"
	}
	ident := func(e syntax.Expr) string {
		if id, ok := e.(*syntax.Identifier); ok {
			return id.Value
		}
		return ""
	}
	priv := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if priv {
			return false
		}
		switch x := x.(type) {
		case *syntax.MethodCall:
			if isThis(x.Var) && ident(x.Name) != "" {
				mm := ix.FindMethod(class, ident(x.Name), ctx.PHP)
				priv = mm != nil && mm.Visibility == index.Private
			}
		case *syntax.PropertyFetch:
			if isThis(x.Var) && ident(x.Name) != "" {
				p := ix.FindProperty(class, ident(x.Name), ctx.PHP)
				priv = p != nil && p.Visibility == index.Private
			}
		case *syntax.StaticCall:
			if isThis(x.Class) && ident(x.Name) != "" {
				mm := ix.FindMethod(class, ident(x.Name), ctx.PHP)
				priv = mm != nil && mm.Visibility == index.Private
			}
		case *syntax.ClassConstFetch:
			if isThis(x.Class) && ident(x.Name) != "" {
				k := ix.FindConst(class, ident(x.Name), ctx.PHP)
				priv = k != nil && k.Visibility == index.Private
			}
		case *syntax.StaticPropertyFetch:
			if v, ok := x.Name.(*syntax.Variable); ok && isThis(x.Class) && v.Name != "" {
				p := ix.FindProperty(class, v.Name, ctx.PHP)
				priv = p != nil && p.Visibility == index.Private
			}
		}
		return !priv
	})
	return priv
}

// smdDelegateEdits replaces the body with a parent:: call (F2).
func smdDelegateEdits(ctx *analysis.Context, m *syntax.Method) []diagnostic.TextEdit {
	args := make([]string, 0, len(m.Params))
	for _, p := range m.Params {
		a := "$" + p.Var.Name
		if p.Variadic {
			a = "..." + a
		}
		args = append(args, a)
	}
	ret := ""
	if smdReturns(ctx, m) {
		ret = "return "
	}
	indent := astquery.IndentBefore(ctx.Src, m.Span().Start)
	text := "{\n" + indent + "    " + ret + "parent::" + m.Name.Value + "(" + strings.Join(args, ", ") + ");\n" + indent + "}"
	return []diagnostic.TextEdit{{Span: m.Body.Span(), NewText: text}}
}

// smdReturns decides whether the delegating call is returned.
func smdReturns(ctx *analysis.Context, m *syntax.Method) bool {
	if m.ReturnType != nil {
		return !strings.EqualFold(ctx.Text(m.ReturnType), "void")
	}
	if c := index.DocComment(ctx.File, m); c != "" {
		if t := strings.TrimSpace(phpdoc.Parse(c).ReturnType()); t != "" && !strings.EqualFold(t, "void") {
			return true
		}
	}
	found := false
	syntax.Inspect(m.Body, func(x syntax.Node) bool {
		if found {
			return false
		}
		switch x := x.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Return:
			found = x.Expr != nil
		case *syntax.Yield, *syntax.YieldFrom:
			found = true
		}
		return !found
	})
	return found
}
