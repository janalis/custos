package confusing

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/syntax"
)

// referencingObjects reports objects taken by reference: typed by-ref
// parameters that are never re-assigned, and `= &new`.
type referencingObjects struct{}

func init() { register(referencingObjects{}) }

func (referencingObjects) ID() string { return "ReferencingObjects" }

func (referencingObjects) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KParam, syntax.KNew}
}

func (r referencingObjects) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Param:
		r.checkParam(ctx, x)
	case *syntax.New:
		r.checkNew(ctx, x)
	}
}

func (referencingObjects) checkParam(ctx *analysis.Context, p *syntax.Param) {
	var body *syntax.Block
	var meth *syntax.Method
	switch fn := p.Parent().(type) { // D1
	case *syntax.Function:
		body = fn.Body
	case *syntax.Method:
		body, meth = fn.Body, fn
	default:
		return
	}
	if !p.ByRef || p.Default != nil || p.Var.NameExpr != nil || p.Type == nil { // D2/D3
		return
	}
	if hasScalarMember(p.Type) { // D3, D3a
		return
	}
	if body == nil { // D6: abstract/interface methods fix the contract of implementations
		return
	}
	if usedAsReference(body, p.Var.Name) || forwardedByRef(ctx, body, p.Var.Name) { // D4, D4a
		return
	}
	if meth != nil && roInHierarchy(ctx, meth) { // D6
		return
	}
	f := ctx.File
	span := syntax.Span{Start: p.Span().Start, End: p.Var.Span().End}
	gap := syntax.Span{Start: p.Type.Span().End, End: p.Var.Span().Start}
	if p.Variadic {
		if t, ok := util.FindToken(f, gap, syntax.TEllipsis); ok {
			gap.End = t.Start
		}
	}
	ctx.Report(span, "Objects are handed over by handle already; drop the '&' before '$"+p.Var.Name+"'.", analysis.Fix{
		Title: "Remove the '&'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: gap, NewText: " "}} },
	})
}

var scalarTypeNames = map[string]bool{
	"string": true, "int": true, "float": true, "bool": true, "array": true,
	"mixed": true, "iterable": true, "null": true,
	"callable": true, // may be a string or an array (spec Divergences)
}

// hasScalarMember reports whether some member of the type other than null
// is a scalar/array pseudo-type (D3, D3a): for such values (`array|\ArrayAccess
// &$ctx` with `$ctx['k'] = 1`) the reference matters.
func hasScalarMember(t syntax.Expr) bool {
	switch x := t.(type) {
	case *syntax.Name:
		n := strings.ToLower(strings.TrimPrefix(x.Value, `\`))
		return n != "null" && scalarTypeNames[n]
	case *syntax.NullableType:
		return hasScalarMember(x.Type)
	case *syntax.UnionType:
		for _, m := range x.Types {
			if hasScalarMember(m) {
				return true
			}
		}
	}
	return false // intersection members are always classes
}

// forwardedByRef implements D4a: the parameter is passed directly to a call
// whose matching parameter is (or, for an unresolved callee, may be) by
// reference — the callee may re-assign it through the reference.
func forwardedByRef(ctx *analysis.Context, body *syntax.Block, name string) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		v, ok := n.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name != name {
			return !found
		}
		arg, ok := v.Parent().(*syntax.Arg)
		if !ok {
			return true
		}
		list := arg.Parent().(*syntax.ArgList)
		params, ok := roCallParams(ctx, list.Parent())
		found = !ok || roBindsByRef(list, arg, params)
		return !found
	})
	return found
}

// roBindsByRef reports whether arg binds to a by-reference parameter
// (positional, variadic tail or named).
func roBindsByRef(list *syntax.ArgList, arg *syntax.Arg, params []index.Param) bool {
	if arg.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return p.ByRef
			}
		}
		return false
	}
	pos := 0
	for _, a := range list.Args {
		if a == syntax.Expr(arg) {
			break
		}
		pos++
	}
	if pos < len(params) {
		return params[pos].ByRef
	}
	n := len(params)
	return n > 0 && params[n-1].Variadic && params[n-1].ByRef
}

// roCallParams resolves the parameters of the function, method or
// constructor called with an argument list.
func roCallParams(ctx *analysis.Context, call syntax.Node) ([]index.Param, bool) {
	var cls []string
	var name string
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := ctx.Types().ResolveFunction(c); f != nil {
			return f.Params, true
		}
		return nil, false
	case *syntax.MethodCall:
		cls = ctx.TypeOf(c.Var).Classes()
		name = roIdent(c.Name)
	case *syntax.StaticCall:
		if fqn := ctx.Types().ClassRef(c.Class); fqn != "" {
			cls = []string{fqn}
		}
		name = roIdent(c.Name)
	case *syntax.New:
		if fqn := ctx.Types().ClassRef(c.Class); fqn != "" {
			cls = []string{fqn}
		}
		name = "__construct"
	}
	if len(cls) != 1 || name == "" {
		return nil, false
	}
	m := ctx.Index().FindMethod(cls[0], name, ctx.PHP)
	if m == nil {
		return nil, false
	}
	return m.Params, true
}

func roIdent(e syntax.Expr) string {
	if id, ok := e.(*syntax.Identifier); ok {
		return id.Value
	}
	return ""
}

// roInHierarchy implements D6: the method overrides a method of an ancestor
// (or its class has an unresolvable ancestor that may declare it), or a
// known descendant overrides it — dropping the '&' would make the
// signatures incompatible.
func roInHierarchy(ctx *analysis.Context, m *syntax.Method) bool {
	cl := m.Parent().(*syntax.ClassLike) // methods only live in class-likes
	if cl.ClassKind == syntax.KindTrait || cl.Name == nil {
		return true
	}
	ix := ctx.Index()
	fqn := ctx.Types().ClassFQN(cl)
	lname := strings.ToLower(m.Name.Value)
	ancestors := ix.Ancestors(fqn, ctx.PHP)
	if !ix.AncestorsComplete(fqn, ctx.PHP) {
		return true // capped: the hierarchy is not fully known (MediaWiki's HookRunner)
	}
	for i, c := range ancestors {
		if i > 0 && c.Methods[lname] != nil {
			return true
		}
		for _, sup := range append(append([]string{c.Parent}, c.Interfaces...), c.Traits...) {
			if sup != "" && ix.Class(sup, ctx.PHP) == nil {
				return true
			}
		}
	}
	below := ctx.Memo("descendant-methods\x00"+strings.ToLower(fqn), func() any {
		return util.DescendantMethods(ix, fqn, ctx.PHP)
	}).(map[string]bool)
	return below[lname]
}

// usedAsReference reports whether the parameter is assigned to or used as a
// logical operand somewhere in body.
func usedAsReference(body *syntax.Block, name string) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if found {
			return false
		}
		v, ok := n.(*syntax.Variable)
		if !ok || v.NameExpr != nil || v.Name != name {
			return true
		}
		if a, ok := v.Parent().(*syntax.Assign); ok && a.Var == syntax.Expr(v) {
			found = true
			return false
		}
		found = util.IsLogicalOperand(v)
		return !found
	})
	return found
}

func (referencingObjects) checkNew(ctx *analysis.Context, nw *syntax.New) {
	a, ok := nw.Parent().(*syntax.Assign) // D5
	if !ok || !a.ByRef || a.Value != syntax.Expr(nw) {
		return
	}
	gap := syntax.Span{Start: a.Op.Span.Start, End: nw.Span().Start}
	ctx.ReportNode(nw, "Objects are handed over by handle already; assign the new instance without '&'.", analysis.Fix{
		Title: "Remove the '&'",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: gap, NewText: "= "}} },
	})
}
