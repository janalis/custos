package unsafeissetoverarray

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
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
		// fixable: the comparison provably behaves like the isset (F1).
		fixable := false
		switch a := arg.(type) {
		case *syntax.Variable: // D6
			if !isetInFuncOrClass(a) {
				return
			}
			fixable = issetDefinedVariable(ctx, a, is)
		case *syntax.PropertyFetch, *syntax.StaticPropertyFetch: // D7
			var usable bool
			if usable, fixable = issetProperty(ctx, a, is); !usable {
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
		msg := "Compare with null instead: '" + r + "'."
		if !fixable {
			ctx.ReportSeverity(subject.Span(), diagnostic.SeverityInfo, msg)
			return
		}
		ctx.ReportSeverity(subject.Span(), diagnostic.SeverityInfo, msg, diagnostic.Fix{
			Title: "Compare with null",
			Edits: func() []diagnostic.TextEdit {
				text := r
				if issetNeedsParens(subject) {
					text = "(" + r + ")"
				}
				return []diagnostic.TextEdit{{Span: subject.Span(), NewText: text}}
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

// issetNeedsParens reports whether the comparison replacing subject must be
// parenthesised: its parent operator binds tighter than a comparison (unary
// operators and casts, instanceof, arithmetic, concatenation, ...) or is a
// non-associative comparison itself.
func issetNeedsParens(subject syntax.Node) bool {
	switch p := subject.Parent().(type) {
	case *syntax.Unary, *syntax.Instanceof:
		return true
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor, syntax.TCoalesce:
			return false
		}
		return true
	}
	return false
}

// issetDefinedVariable reports whether the plain local variable v is known
// to exist at the isset `at`, so that reading it raises no diagnostic: a
// parameter, a closure import, or a variable assigned / declared global or
// static by a statement executed on every path to at (no goto in the
// function) - and never unset() in its function.
func issetDefinedVariable(ctx *analysis.Context, v *syntax.Variable, at syntax.Node) bool {
	name, ok := astquery.PlainVariableName(v)
	scope := syntax.EnclosingFuncLike(at)
	if !ok || scope == nil {
		return false
	}
	defined := false
	for _, p := range syntax.FuncLikeParams(scope) {
		defined = defined || (p.Var != nil && p.Var.Name == name)
	}
	if c, ok := scope.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			defined = defined || (u.Var != nil && u.Var.Name == name)
		}
	}
	written := false
	for _, acc := range flowquery.VarAccesses(ctx.File, scope, name) {
		switch by := acc.By.(type) {
		case *syntax.Unset:
			return false
		case *syntax.Assign, *syntax.Global, *syntax.StaticVar:
			written = written || issetDominates(by, at)
		}
	}
	// A goto can jump over the write.
	return defined || written && !issetHasGoto(scope)
}

// issetHasGoto reports whether scope contains a goto statement.
func issetHasGoto(scope syntax.Node) bool {
	found := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		_, ok := n.(*syntax.Goto)
		found = found || ok
		return !found
	})
	return found
}

// issetDominates reports whether the write `by` (an assignment, a global
// statement or a static variable declaration) is a statement of a statement
// list enclosing at and placed before it, hence executed on every path to at.
func issetDominates(by, at syntax.Node) bool {
	st := by
	switch by.(type) {
	case *syntax.Assign:
		for st = by.Parent(); ; st = st.Parent() {
			if _, ok := st.(*syntax.Paren); !ok {
				break
			}
		}
		if _, ok := st.(*syntax.ExprStmt); !ok {
			return false
		}
	case *syntax.StaticVar:
		st = by.Parent()
	}
	switch st.Parent().(type) {
	case *syntax.Block, *syntax.Case:
		return st.Span().End <= at.Span().Start && flowquery.NodeContains(st.Parent(), at)
	}
	return false
}

// issetProperty classifies a property access (e is a *syntax.PropertyFetch
// or a *syntax.StaticPropertyFetch) used as the argument of the isset `at`.
// usable: it resolves to a declared property (class body or constructor
// promotion; not magic) that a null comparison can read from here - not a
// typed property without default (isset is the only probe of its
// uninitialized state) nor one hidden by its visibility. fixable: the
// comparison is moreover provably equivalent (D8a).
func issetProperty(ctx *analysis.Context, e syntax.Expr, at syntax.Node) (usable, fixable bool) {
	var classes []string
	var name string
	if p, ok := e.(*syntax.PropertyFetch); ok {
		id, ok := p.Name.(*syntax.Identifier)
		if !ok {
			return false, false
		}
		name = id.Value
		t := ctx.TypeOf(p.Var)
		classes = t.Classes()
		// The receiver must be an object for sure: `$this`, or a defined
		// variable whose every type is a class.
		if v, ok := p.Var.(*syntax.Variable); ok && len(classes) == len(t.Atoms()) {
			fixable = v.Name == "this" || issetDefinedVariable(ctx, v, at)
		}
	} else {
		p := e.(*syntax.StaticPropertyFetch)
		v, ok := p.Name.(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return false, false
		}
		name = v.Name
		if n, ok := p.Class.(*syntax.Name); ok {
			fixable = true
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
	unset := issetUnsetProperties(ctx)
	self := ctx.Types().ClassFQN(syntax.EnclosingClass(e))
	for _, c := range classes {
		c = strings.TrimPrefix(c, `\`)
		prop := ix.FindProperty(c, name, ctx.PHP)
		if prop == nil || prop.Magic {
			fixable = false
			continue
		}
		if prop.Type != "" && !prop.HasDefault && !prop.Promoted || !issetReadable(ctx, prop, self) {
			return false, false
		}
		usable = true
		// An unset() instance property is routed to __isset() / __get().
		if !prop.Static && (ix.FindMethod(c, "__isset", ctx.PHP) != nil || ix.FindMethod(c, "__get", ctx.PHP) != nil || unset[name] || unset[""]) {
			fixable = false
		}
	}
	return usable, usable && fixable
}

// issetReadable reports whether prop can be read from class self ("" outside
// a class).
func issetReadable(ctx *analysis.Context, prop *index.Property, self string) bool {
	switch prop.Visibility {
	case index.Public:
		return true
	case index.Private:
		// TypeClass: the class importing a trait's property.
		self = strings.TrimPrefix(self, `\`)
		return self != "" && (strings.EqualFold(strings.TrimPrefix(prop.Class, `\`), self) || strings.EqualFold(strings.TrimPrefix(prop.TypeClass, `\`), self))
	}
	return self != "" && ctx.Index().IsSubtype(self, prop.Class, ctx.PHP)
}

// issetUnsetProperties returns the names of the instance properties the file
// passes to unset(); "" stands for a dynamic name (any property).
func issetUnsetProperties(ctx *analysis.Context) map[string]bool {
	return ctx.Memo("unset", func() any {
		m := map[string]bool{}
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if u, ok := n.(*syntax.Unset); ok {
				for _, x := range u.Vars {
					if p, ok := syntax.UnwrapParens(x).(*syntax.PropertyFetch); ok {
						id, _ := p.Name.(*syntax.Identifier)
						if id == nil {
							m[""] = true
							continue
						}
						m[id.Value] = true
					}
				}
			}
			return true
		})
		return m
	}).(map[string]bool)
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
