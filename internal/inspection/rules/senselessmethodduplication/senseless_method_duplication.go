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
	// A same-visibility duplicate is removed, so its whole signature must
	// match; a proxy keeps its own defaults and return type.
	identical := smdVisibility(m.Modifiers) == pm.Visibility
	if !smdSameSignature(ctx, m, pmDecl, identical) { // D7a
		return
	}
	roots, proots := smdRoots(m, identical), smdRoots(pmDecl, identical)
	mine, ok := smdSymbols(ctx, roots, ownFQN) // D8
	if !ok {
		return
	}
	theirs, ok := smdSymbols(ctx, proots, pm.Class)
	if !ok || len(mine) != len(theirs) {
		return
	}
	for k := range mine {
		if !theirs[k] {
			return
		}
	}
	if smdTouchesPrivate(ctx, pmDecl.Body, pm.Class) || smdTouchesPrivate(ctx, m.Body, ownFQN) { // D9
		return
	}
	if smdScopeBound(roots) || smdTraitProvides(ctx, cl, m.Name.Value) { // D11, D12
		return
	}
	f := ctx.File
	name := m.Name.Value
	if identical { // D10 I
		ctx.Report(m.Name.Span(), "Method '"+name+"' duplicates the inherited implementation; remove it.", diagnostic.Fix{
			Title: "Remove the method",
			Edits: func() []diagnostic.TextEdit { return astquery.MethodRemovalEdits(f, m) },
		})
		return
	}
	msg := "Method '" + name + "' duplicates the inherited implementation; delegate to parent::" + name + "() instead."
	if m.ByRef || smdReadsArguments(m.Body) { // F2: the delegating call would not be equivalent
		ctx.Report(m.Name.Span(), msg)
		return
	}
	ctx.Report(m.Name.Span(), msg, diagnostic.Fix{
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

// smdRoots lists the subtrees of m whose meaning must match the parent's:
// the body and, when the signature is compared in full, the defaults.
func smdRoots(m *syntax.Method, defaults bool) []syntax.Node {
	roots := []syntax.Node{m.Body}
	if defaults {
		for _, p := range m.Params {
			if p.Default != nil {
				roots = append(roots, p.Default)
			}
		}
	}
	return roots
}

// smdSameSignature compares the parameters of a and b (D7a): names, order,
// types, by-reference and variadic markers. full adds what only matters
// when the method is removed: defaults, return type and by-reference return.
func smdSameSignature(ctx *analysis.Context, a, b *syntax.Method, full bool) bool {
	if len(a.Params) != len(b.Params) {
		return false
	}
	if full && (a.ByRef != b.ByRef || !smdSameType(ctx, a.ReturnType, b.ReturnType)) {
		return false
	}
	for i, p := range a.Params {
		q := b.Params[i]
		if p.Var.Name != q.Var.Name || p.ByRef != q.ByRef || p.Variadic != q.Variadic || !smdSameType(ctx, p.Type, q.Type) {
			return false
		}
		if len(p.Modifiers) != 0 || len(q.Modifiers) != 0 { // promoted: the override declares a property
			return false
		}
		if !full || p.Default == nil && q.Default == nil {
			continue
		}
		if p.Default == nil || q.Default == nil || !astquery.EquivalentFoldNames(ctx.File, p.Default, q.Default) {
			return false
		}
	}
	return true
}

// smdSameType reports whether two type declarations are written alike and
// their class names resolve to the same class.
func smdSameType(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	compact := func(e syntax.Expr) string { return strings.Join(strings.Fields(ctx.Text(e)), "") }
	if !strings.EqualFold(compact(a), compact(b)) {
		return false
	}
	names := func(e syntax.Expr) (out []*syntax.Name) {
		syntax.Inspect(e, func(x syntax.Node) bool {
			if n, ok := x.(*syntax.Name); ok {
				out = append(out, n)
			}
			return true
		})
		return out
	}
	ix := ctx.Index()
	nb := names(b)
	for i, x := range names(a) {
		y := nb[i]
		fa, fb := ctx.Names().Class(x.Value, x.Span().Start), ctx.Names().Class(y.Value, y.Span().Start)
		// Builtin type names resolve to no class on either side.
		if !strings.EqualFold(fa, fb) && (ix.Class(fa, ctx.PHP) != nil || ix.Class(fb, ctx.PHP) != nil) {
			return false
		}
	}
	return true
}

// smdScopeBound reports code whose result depends on the declaring method
// or class, so an identical copy is not a duplicate (D11): class-, method-,
// namespace- and line-dependent magic constants, the method's own static
// variables, and functions answering for the calling scope.
func smdScopeBound(roots []syntax.Node) bool {
	bound := false
	for _, root := range roots {
		syntax.Inspect(root, func(x syntax.Node) bool {
			switch x := x.(type) {
			case *syntax.MagicConst:
				switch x.Token.Kind {
				case syntax.TClassC, syntax.TMethodC, syntax.TNsC, syntax.TLine:
					bound = true
				}
			case *syntax.StaticStmt:
				bound = bound || smdOwnStatement(x)
			case *syntax.FuncCall:
				switch strings.ToLower(astquery.CallLastName(x)) {
				case "get_class", "get_parent_class":
					bound = bound || x.Args == nil || len(x.Args.Args) == 0
				case "get_object_vars", "get_class_vars", "get_class_methods":
					bound = true // visibility is judged from the calling class
				}
			}
			return !bound
		})
	}
	return bound
}

// smdOwnStatement reports whether n belongs to the method itself and not
// to a closure or function nested in its body.
func smdOwnStatement(n syntax.Node) bool {
	for p := n.Parent(); ; p = p.Parent() {
		switch p.(type) {
		case *syntax.Closure, *syntax.Function:
			return false
		case *syntax.Method:
			return true
		}
	}
}

// smdTraitProvides reports whether a trait used by cl supplies a method
// called name, which would replace the inherited one once the class's own
// declaration is gone (D12). An unknown trait counts as supplying it.
func smdTraitProvides(ctx *analysis.Context, cl *syntax.ClassLike, name string) bool {
	ix := ctx.Index()
	for _, mem := range cl.Members {
		use, ok := mem.(*syntax.TraitUse)
		if !ok {
			continue
		}
		for _, a := range use.Adaptations {
			if a.Alias != nil && strings.EqualFold(a.Alias.Value, name) {
				return true
			}
		}
		for _, t := range use.Traits {
			fqn := ctx.Names().Class(t.Value, t.Span().Start)
			if ix.Class(fqn, ctx.PHP) == nil || ix.FindMethod(fqn, name, ctx.PHP) != nil {
				return true
			}
		}
	}
	return false
}

// smdReadsArguments reports calls reading the actual argument list, which a
// delegating call passing every declared parameter would change (F2).
func smdReadsArguments(body *syntax.Block) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if c, ok := x.(*syntax.FuncCall); ok {
			switch strings.ToLower(astquery.CallLastName(c)) {
			case "func_num_args", "func_get_args", "func_get_arg":
				found = true
			}
		}
		return !found
	})
	return found
}

// smdSymbols collects the resolved class, constant and function references
// of the given subtrees (D8); ok is false when one does not resolve.
func smdSymbols(ctx *analysis.Context, roots []syntax.Node, class string) (map[string]bool, bool) {
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
	visit := func(x syntax.Node) bool {
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
	}
	for _, root := range roots {
		syntax.Inspect(root, visit)
	}
	return out, ok
}

// smdTouchesPrivate reports $this->/$this:: accesses in body resolving to a
// private member, seen from class (D9): the parent's are out of the child's
// reach, the child's out of the inherited method's. Literal self:: accesses never get
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
		switch strings.ToLower(ctx.Text(m.ReturnType)) {
		case "void", "never":
			return false
		}
		return true
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
