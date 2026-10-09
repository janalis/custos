package unsafeissetoverarray

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// unSafeIsSetOverArray reports single-argument isset() calls that a more
// explicit construct expresses better: null comparison, array_key_exists(),
// or a precomputed concatenated key.
type unSafeIsSetOverArray struct{}

func (unSafeIsSetOverArray) ID() string               { return "UnSafeIsSetOverArray" }
func (unSafeIsSetOverArray) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIsset} }
func (unSafeIsSetOverArray) Semantic()                {}
func (unSafeIsSetOverArray) Check(ctx *analysis.Context, n syntax.Node) {
	is := n.(*syntax.Isset)
	if len(is.Vars) != 1 || is.Vars[0] == nil { // D1
		return
	}
	var subject syntax.Node = is // D2
	inverted := false
	if u, ok := is.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		subject, inverted = u, true
	}
	parent := subject.Parent()
	stored := false // D3
	switch parent.(type) {
	case *syntax.Assign, *syntax.Return:
		stored = true
	}
	if t, ok := parent.(*syntax.Ternary); ok && t.Cond == subject { // D4
		branch := t.Else
		if inverted {
			branch = t.Then
		}
		if syntax.IsNullConst(branch) {
			return
		}
	}
	arg := syntax.UnwrapParens(is.Vars[0]) // D5

	dim, isDim := arg.(*syntax.ArrayDimFetch)
	if !isDim {
		switch a := arg.(type) {
		case *syntax.Variable: // D6
			if !isetInFuncOrClass(a) {
				return
			}
		case *syntax.PropertyFetch, *syntax.StaticPropertyFetch: // D7
			if !issetDeclaredProperty(ctx, a) {
				return
			}
		}
		if !ctx.Bool("SUGGEST_TO_USE_NULL_COMPARISON") || issetInFinally(is) { // D8
			return
		}
		op := "!=="
		if inverted {
			op = "==="
		}
		a := ctx.Text(arg)
		r := a + " " + op + " null"
		if ctx.ComparisonStyle == analysis.StyleYoda {
			r = "null " + op + " " + a
		}
		ctx.ReportSeverity(subject.Span(), diagnostic.SeverityInfo, "Compare with null instead: '"+r+"'.", diagnostic.Fix{
			Title: "Compare with null",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: subject.Span(), NewText: r}}
			},
		})
		return
	}

	if ctx.Bool("REPORT_CONCATENATION_IN_INDEXES") && !stored { // D9
		// Every [...] level of the access chain (custos diverges: upstream
		// only looks at the last one).
		for d := dim; d != nil; {
			if b, ok := d.Dim.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
				ctx.ReportNode(arg, "Compute the concatenated key in a variable before using it.")
				return
			}
			d, _ = d.Var.(*syntax.ArrayDimFetch)
		}
	}
	if ctx.Bool("SUGGEST_TO_USE_ARRAY_KEY_EXISTS") && dim.Var != nil && !issetObjectContainer(ctx, dim.Var) { // D10
		ctx.ReportSeverity(arg.Span(), diagnostic.SeverityInfo, "Use array_key_exists() to check for the key itself.")
	}
}

func isetInFuncOrClass(n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.ClassLike); ok || syntax.IsFuncLike(p) {
			return true
		}
	}
	return false
}

func issetInFinally(n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Finally); ok {
			return true
		}
	}
	return false
}

// issetDeclaredProperty reports whether a property access resolves to a
// declared property (class body or constructor promotion; not magic). e is
// a *syntax.PropertyFetch or a *syntax.StaticPropertyFetch.
func issetDeclaredProperty(ctx *analysis.Context, e syntax.Expr) bool {
	var classes []string
	var name string
	if p, ok := e.(*syntax.PropertyFetch); ok {
		id, ok := p.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		name = id.Value
		classes = ctx.TypeOf(p.Var).Classes()
	} else {
		p := e.(*syntax.StaticPropertyFetch)
		v, ok := p.Name.(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return false
		}
		name = v.Name
		if n, ok := p.Class.(*syntax.Name); ok {
			switch strings.ToLower(n.Value) {
			case "self", "static", "parent":
				// parent:: is approximated by a hierarchy lookup from the
				// enclosing class.
				if c := ctx.Types().ClassFQN(syntax.EnclosingClass(p)); c != "" {
					classes = []string{c}
				}
			default:
				classes = []string{ctx.Names().Class(n.Value, n.Span().Start)}
			}
		} else {
			classes = ctx.TypeOf(p.Class).Classes()
		}
	}
	ix := ctx.Index()
	for _, c := range classes {
		if prop := ix.FindProperty(strings.TrimPrefix(c, `\`), name, ctx.PHP); prop != nil && !prop.Magic {
			return true
		}
	}
	return false
}

// issetObjectContainer reports whether the container's resolved types are
// only class-like (or mixed), ignoring null.
func issetObjectContainer(ctx *analysis.Context, container syntax.Expr) bool {
	t := ctx.TypeOf(container)
	objects := 0
	for _, a := range t.Atoms() {
		switch {
		case a == "null":
		case a == "mixed" || a == "object":
			objects++
		case strings.HasSuffix(a, "[]"):
			return false
		case strings.HasPrefix(a, `\`):
			objects++
		default:
			return false
		}
	}
	return objects > 0
}
