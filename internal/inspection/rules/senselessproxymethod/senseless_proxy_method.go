package senselessproxymethod

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// senselessProxyMethod reports overrides that only forward their
// parameters unchanged to the same parent method.
type senselessProxyMethod struct{}

func (senselessProxyMethod) ID() string               { return "SenselessProxyMethod" }
func (senselessProxyMethod) Semantic()                {}
func (senselessProxyMethod) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }
func (senselessProxyMethod) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	cl, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cl.ClassKind != syntax.KindClass || m.Name == nil || m.Body == nil || m.Name.Span().Len() == 0 {
		return
	}
	if m.Modifiers.Has(syntax.TAbstract) || m.Modifiers.Has(syntax.TPrivate) || len(m.Attrs) > 0 { // D1
		return
	}
	stmts := astquery.MethodStatements(m.Body) // D2
	if len(stmts) != 1 {
		return
	}
	var call syntax.Expr // D3
	returns := false
	switch s := stmts[0].(type) {
	case *syntax.Return:
		call, returns = s.Expr, true
	case *syntax.ExprStmt:
		call = s.Expr
	}
	sc, ok := call.(*syntax.StaticCall)
	if !ok {
		return
	}
	cls, ok := sc.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(cls.Value, "parent") {
		return
	}
	id, ok := sc.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, m.Name.Value) || sc.Args == nil {
		return
	}
	if len(sc.Args.Args) != len(m.Params) { // D4
		return
	}
	for i, a := range sc.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Name != nil || arg.Unpack || arg.ByRef {
			return
		}
		v, ok := arg.Value.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name != m.Params[i].Var.Name {
			return
		}
	}
	ix := ctx.Index()
	ownFQN := ctx.Names().DeclFQN(cl)
	parentFQN := ctx.Names().ParentFQN(cl)
	own := ix.Class(ownFQN, ctx.PHP)
	if own == nil || parentFQN == "" {
		return
	}
	mine := own.Methods[strings.ToLower(m.Name.Value)]
	pm := semanticquery.MethodInChain(ix, parentFQN, m.Name.Value, ctx.PHP) // D5
	if mine == nil || pm == nil {
		return
	}
	if len(pm.Params) != len(mine.Params) || pm.Abstract != mine.Abstract || pm.Static != mine.Static || // D6
		pm.Final != mine.Final || pm.Visibility != mine.Visibility {
		return
	}
	pmDecl := semanticquery.MethodDecl(ctx.File, ix, pm, ctx.PHP)
	if !returns && !spmParentReturnsNothing(pm, pmDecl) { // D10
		return
	}
	isCtor := strings.EqualFold(m.Name.Value, "__construct")
	if isCtor && spmLegacyCtor(ctx, cl, ownFQN) { // E7
		return
	}
	for i, p := range m.Params { // D7
		pp, mp := pm.Params[i], mine.Params[i]
		if (p.Default == nil) != (pp.Default == "") {
			return
		}
		if p.Default != nil {
			if _, ok := p.Default.(*syntax.MagicConst); ok && spmMagic[strings.ToUpper(ctx.Text(p.Default))] {
				return
			}
			if pmDecl != nil {
				if !astquery.EquivalentFoldNames(ctx.File, p.Default, pmDecl.Params[i].Default) {
					return
				}
			} else if astquery.SquashWhitespace(ctx.Text(p.Default)) != astquery.SquashWhitespace(pp.Default) {
				return
			}
		}
		if spmParamType(mp) != spmParamType(pp) || len(p.Attrs) > 0 {
			return
		}
		if isCtor && (mp.Promoted || pp.Promoted) {
			if !mp.Promoted || !pp.Promoted {
				return
			}
			op, oo := own.Props[mp.Name], ix.Class(pm.Class, ctx.PHP)
			if op == nil || oo == nil || oo.Props[pp.Name] == nil || op.Visibility != oo.Props[pp.Name].Visibility {
				return
			}
		}
	}
	if mine.Return != pm.Return { // D8
		return
	}
	if c := index.DocComment(ctx.File, m); c != "" { // D9
		if pmDecl == nil {
			return // the parent's doc comment is not available
		}
		mineTags := phpdoc.Parse(c).Tags
		var parentTags []phpdoc.Tag
		if pc := index.DocComment(ctx.File, pmDecl); pc != "" {
			parentTags = phpdoc.Parse(pc).Tags
		}
		if len(mineTags) != len(parentTags) {
			return
		}
		for _, t := range mineTags {
			found := false
			for _, o := range parentTags {
				if t.Name == o.Name && astquery.SquashWhitespace(t.Text) == astquery.SquashWhitespace(o.Text) {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}
	}
	f := ctx.File
	ctx.ReportSeverity(m.Name.Span(), diagnostic.SeverityInfo, "Method '"+m.Name.Value+"' only forwards to its parent; remove it.", diagnostic.Fix{
		Title: "Remove the method",
		Edits: func() []diagnostic.TextEdit { return astquery.MethodRemovalEdits(f, m) },
	})
}

// spmParentReturnsNothing implements D10: the parent method is known not to
// produce a value, so a child dropping the `return` is still equivalent.
func spmParentReturnsNothing(pm *index.Method, decl *syntax.Method) bool {
	switch strings.ToLower(pm.Return) {
	case "void", "never":
		return true
	case "":
	default:
		return false
	}
	if decl == nil || decl.Body == nil {
		return strings.ToLower(pm.DocReturn) == "void"
	}
	value := false
	syntax.Inspect(decl.Body, func(n syntax.Node) bool {
		if value {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Return:
			value = n.Expr != nil
		case *syntax.Yield, *syntax.YieldFrom:
			value = true // a generator returns a Generator object
		}
		return !value
	})
	return !value
}

// spmLegacyCtor implements E7: below PHP 8.0 a class outside any namespace
// with a method named like itself would make that method the constructor
// once `__construct` is removed.
func spmLegacyCtor(ctx *analysis.Context, cl *syntax.ClassLike, fqn string) bool {
	if !ctx.PHP.Below(phpversion.PHP80) || cl.Name == nil || strings.Contains(strings.TrimPrefix(fqn, `\`), `\`) {
		return false
	}
	for _, mem := range cl.Members {
		if mm, ok := mem.(*syntax.Method); ok && mm.Name != nil && strings.EqualFold(mm.Name.Value, cl.Name.Value) {
			return true
		}
	}
	return false
}

// spmMagic are magic constants evaluating differently in the child.
var spmMagic = map[string]bool{
	"__LINE__": true, "__FILE__": true, "__DIR__": true, "__FUNCTION__": true,
	"__CLASS__": true, "__TRAIT__": true, "__METHOD__": true, "__NAMESPACE__": true,
}

// spmParamType is the comparable native type of an indexed parameter.
func spmParamType(p index.Param) string {
	if p.Variadic {
		return "array"
	}
	return strings.ToLower(p.Type)
}
