package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// isEmptyFunctionUsage suggests type-specific checks instead of empty().
type isEmptyFunctionUsage struct{}

func init() { register(isEmptyFunctionUsage{}) }

func (isEmptyFunctionUsage) ID() string { return "IsEmptyFunctionUsage" }

func (isEmptyFunctionUsage) Semantic() {}

func (isEmptyFunctionUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KEmpty} }

const isEmptyGenericMsg = "Prefer a type-specific check over empty()."

// emptySubjectTypes returns the normalised type names of the subject (nil
// when unknown).
func emptySubjectTypes(ctx *analysis.Context, s syntax.Expr) []string {
	t := ctx.TypeOf(s) // calls without declared return are typed from their body
	if t.IsUnknown() {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, a := range t.Atoms() {
		switch {
		case strings.HasSuffix(a, "[]"):
			a = "array"
		case a == "true" || a == "false":
			a = "bool"
		}
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

func (r isEmptyFunctionUsage) Check(ctx *analysis.Context, n syntax.Node) {
	e := n.(*syntax.Empty)
	s := syntax.UnwrapParens(e.Expr)
	if _, ok := s.(*syntax.ArrayDimFetch); ok { // D0
		return
	}
	var inv *syntax.Unary
	if u, ok := e.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		inv = u
	}
	ts := emptySubjectTypes(ctx, s)
	if len(ts) > 0 && !isEmptySubjectAssigned(s) {
		// `count($v)` / `$v === null` warn on an unassigned variable;
		// empty() does not: only the generic report remains.
		ts = nil
	}

	// D1
	if ctx.Bool("SUGGEST_TO_USE_COUNT_CHECK") && len(ts) > 0 {
		ok := true
		for _, t := range ts {
			if t == "array" {
				continue
			}
			if !strings.HasPrefix(t, `\`) || ctx.Index().Class(t, ctx.PHP) == nil ||
				!ctx.Index().IsSubtype(t, "Countable", ctx.PHP) {
				ok = false
				break
			}
		}
		if ok {
			r.suggest(ctx, e, inv, s, util.QualifiedBuiltin(ctx, "count", e.Span().Start)+"(%s)", "0") // a namespaced count() would capture a bare call
			return
		}
	}

	// D2
	if ctx.Bool("SUGGEST_TO_USE_NULL_COMPARISON") && len(ts) > 0 {
		applies := false
		if ctx.Bool("SUGGEST_NULL_COMPARISON_FOR_SCALARS") && len(ts) == 2 {
			hasNull, scalar := false, false
			for _, t := range ts {
				switch t {
				case "null":
					hasNull = true
				case "resource": // int/float/bool: empty() is also true for 0, 0.0, false
					scalar = true
				}
			}
			applies = hasNull && scalar
		}
		if !applies {
			cnt, all := 0, true
			for _, t := range ts {
				if t == "null" {
					continue
				}
				cnt++
				// custos: SimpleXMLElement and GMP objects can be empty
				// (`empty(gmp_init(0))`), so `=== null` would differ.
				if !strings.HasPrefix(t, `\`) || isEmptyFalsyObject(ctx, t) {
					all = false
				}
			}
			applies = cnt > 0 && all
		}
		if applies {
			if isEmptyThroughProperty(s) {
				return
			}
			r.suggest(ctx, e, inv, s, "%s", "null")
			return
		}
	}

	// D3
	if ctx.Bool("REPORT_EMPTY_USAGE") {
		ctx.Report(e.Span(), isEmptyGenericMsg)
	}
}

// isEmptyThroughProperty descends from s through first child nodes (names
// and identifiers count as tokens) looking for a property access. The
// descent stops at argument lists: a property passed to a call does not
// make the call's result a property access.
func isEmptyThroughProperty(s syntax.Node) bool {
	for s != nil {
		switch s.(type) {
		case *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
			return true
		}
		var first syntax.Node
		syntax.Children(s, func(c syntax.Node) {
			if first != nil {
				return
			}
			switch c.(type) {
			case *syntax.Name, *syntax.Identifier:
				return
			}
			first = c
		})
		if _, isArgs := first.(*syntax.ArgList); isArgs {
			return false
		}
		s = first
	}
	return false
}

func (isEmptyFunctionUsage) suggest(ctx *analysis.Context, e *syntax.Empty, inv *syntax.Unary, s syntax.Expr, subjFmt, other string) {
	var target syntax.Node = e
	op := "==="
	if inv != nil {
		target, op = inv, "!=="
	}
	st := ctx.Text(s)
	if subjFmt == "%s" && util.NeedsParensAsEqualityOperand(s) {
		st = "(" + st + ")" // `$a ?? $b === null` would compare $b only
	}
	subj := strings.Replace(subjFmt, "%s", st, 1)
	repl := subj + " " + op + " " + other
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = other + " " + op + " " + subj
	}
	msg := "Replace with '" + repl + "'."
	if isEmptyNeedsParens(target) {
		repl = "(" + repl + ")"
	}
	span := target.Span()
	ctx.Report(span, msg, analysis.Fix{
		Title: "Use a type-specific check",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}

// isEmptyNeedsParens reports whether a comparison replacing n would bind
// looser than n's parent expression.
func isEmptyNeedsParens(n syntax.Node) bool {
	switch p := n.Parent().(type) {
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor, syntax.TCoalesce:
			return false
		}
		return true
	case *syntax.Unary, *syntax.Instanceof:
		return true
	}
	return false
}

// isEmptySubjectAssigned reports whether a variable subject is certainly
// assigned where empty() reads it: `$this`, a parameter or closure import,
// a variable at file scope or in an arrow function (not tracked), or one
// assigned (or declared global/static) by a statement of the function body
// itself before the statement holding empty(). Other subjects are not
// variables and always qualify.
func isEmptySubjectAssigned(s syntax.Expr) bool {
	v, ok := s.(*syntax.Variable)
	if !ok || v.Name == "" || v.Name == "this" {
		return true
	}
	fn := syntax.EnclosingFuncLike(v)
	body := syntax.FuncLikeBody(fn)
	if body == nil {
		return true
	}
	for _, p := range syntax.FuncLikeParams(fn) {
		if p.Var.Name == v.Name {
			return true
		}
	}
	if c, ok := fn.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if u.Var.Name == v.Name {
				return true
			}
		}
	}
	var holder syntax.Node = v
	for holder.Parent() != syntax.Node(body) {
		holder = holder.Parent()
		if holder == nil {
			return true // recovery tree: the subject is not under the body
		}
	}
	for _, st := range body.Stmts {
		if syntax.Node(st) == holder {
			break
		}
		switch st := st.(type) {
		case *syntax.ExprStmt:
			if a, ok := st.Expr.(*syntax.Assign); ok && isEmptyVarNamed(a.Var, v.Name) {
				return true
			}
		case *syntax.Global:
			for _, g := range st.Vars {
				if isEmptyVarNamed(g, v.Name) {
					return true
				}
			}
		case *syntax.StaticStmt:
			for _, sv := range st.Vars {
				if sv.Var.Name == v.Name {
					return true
				}
			}
		}
	}
	return false
}

func isEmptyVarNamed(e syntax.Node, name string) bool {
	v, ok := e.(*syntax.Variable)
	return ok && v.Name == name
}

// isEmptyFalsyObject reports whether objects of class t (or of an
// unresolvable class) may convert to false.
func isEmptyFalsyObject(ctx *analysis.Context, t string) bool {
	cls := strings.TrimPrefix(t, `\`)
	if ctx.Index().Class(cls, ctx.PHP) == nil {
		return false // unresolvable: as before (D2b)
	}
	return ctx.Index().IsSubtype(cls, "SimpleXMLElement", ctx.PHP) || ctx.Index().IsSubtype(cls, "GMP", ctx.PHP)
}
