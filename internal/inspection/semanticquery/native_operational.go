package semanticquery

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// NativeMethod proves the extension-owned method contract, including overrides.
func NativeMethod(ctx *analysis.Context, c *syntax.MethodCall, class, name string) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, name) {
		return false
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], class, ctx.PHP) {
		return false
	}
	m := ctx.Index().FindMethod(classes[0], id.Value, ctx.PHP)
	return m != nil && m.Builtin && m.Avail.In(ctx.PHP)
}

// NativeDominates proves that an earlier expression statement executes on every
// ordinary path to at in the same lexical block. Conditional bodies do not.
func NativeDominates(earlier, at syntax.Node) bool {
	if earlier.Span().End > at.Span().Start {
		return false
	}
	// A nested conditional expression does not prove its guarded calls execute.
	for node := earlier; node != nil; node = node.Parent() {
		if _, ok := node.(*syntax.ExprStmt); ok {
			break
		}
		switch parent := node.Parent().(type) {
		case *syntax.Ternary:
			if node != parent.Cond {
				return false
			}
		case *syntax.Binary:
			if node == parent.Right && (parent.Op.Kind == syntax.TBooleanAnd || parent.Op.Kind == syntax.TBooleanOr || parent.Op.Kind == syntax.TAnd || parent.Op.Kind == syntax.TOr || parent.Op.Kind == syntax.TCoalesce) {
				return false
			}
		case *syntax.Assign:
			if node == parent.Value && parent.Op.Kind == syntax.TCoalesceEqual {
				return false
			}
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.MatchArm:
			// Another arm, or code after the match, runs without this one.
			if !parent.Span().Contains(at.Span()) {
				return false
			}
		case *syntax.MethodCall:
			// A null receiver skips the arguments of the whole chain.
			if node != parent.Var && nativeNullsafeChain(parent) && !parent.Span().Contains(at.Span()) {
				return false
			}
		case *syntax.PropertyFetch:
			if node != parent.Var && nativeNullsafeChain(parent) && !parent.Span().Contains(at.Span()) {
				return false
			}
		case *syntax.ArrayDimFetch:
			if node != parent.Var && nativeNullsafeChain(parent.Var) && !parent.Span().Contains(at.Span()) {
				return false
			}
		}
	}
	statement := earlier
	for statement != nil {
		if _, ok := statement.(*syntax.ExprStmt); ok {
			break
		}
		statement = statement.Parent()
	}
	if statement == nil {
		return false
	}
	owner := statement.Parent()
	if _, ok := syntax.StmtListOf(owner); !ok {
		return false
	}
	for p := at; p != nil; p = p.Parent() {
		if p == owner {
			return true
		}
	}
	return false
}

// nativeNullsafeChain reports a member chain that a null-safe access can cut.
func nativeNullsafeChain(e syntax.Expr) bool {
	for {
		switch x := syntax.UnwrapParens(e).(type) {
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

// ReplacementFix is the quick-fix titled title that replaces span with text.
func ReplacementFix(span syntax.Span, text, title string) diagnostic.Fix {
	return diagnostic.Fix{Title: title, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}

// CommentFreeSpan reports whether the source of span holds no comment
// marker, even inside a string, so replacing it loses no comment.
func CommentFreeSpan(ctx *analysis.Context, span syntax.Span) bool {
	text := ctx.SpanText(span)
	return !strings.Contains(text, "/*") && !strings.Contains(text, "//") && !strings.Contains(text, "#")
}

func nativeFinally(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, names ...string) bool {
	for p := at.Parent(); p != nil; p = p.Parent() {
		tr, ok := p.(*syntax.Try)
		if !ok || tr.Finally == nil {
			continue
		}
		for _, s := range tr.Finally.Body.Stmts {
			st, ok := s.(*syntax.ExprStmt)
			if !ok {
				continue
			}
			var candidate syntax.Expr
			name := ""
			switch c := st.Expr.(type) {
			case *syntax.FuncCall:
				name = NativeBuiltinName(ctx, c)
				candidate = CallArgument(c.Args, 0, "")
			case *syntax.MethodCall:
				candidate = c.Var
				if id, ok := c.Name.(*syntax.Identifier); ok {
					name = strings.ToLower(id.Value)
				}
			}
			if receiver == nil || candidate == nil || ctx.Flow().Value(receiver).Identity == 0 || ctx.Flow().Value(receiver).Identity != ctx.Flow().Value(candidate).Identity {
				continue
			}
			for _, wanted := range names {
				if wanted == name {
					return true
				}
			}
		}
	}
	return false
}

// GlobalCallName is how to call the global function name at at: with a
// leading backslash when a namespace or an imported alias would make the
// bare name resolve elsewhere.
func GlobalCallName(ctx *analysis.Context, name string, at syntax.Node) string {
	fqn, _ := ctx.Names().Function(name, at.Span().Start)
	if !strings.EqualFold(fqn, name) {
		return `\` + name
	}
	return name
}
