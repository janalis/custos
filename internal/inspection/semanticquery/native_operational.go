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
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
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
	if owner == nil {
		return syntax.EnclosingVariableScope(at) == nil
	}
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

// NativePriorCalls returns unconditionally preceding calls matching the same
// receiver allocation identity. Unknown identities establish no relation.
func NativePriorCalls(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, names ...string) []syntax.Node {
	target := ctx.Flow().Value(receiver)
	id := target.Identity
	if !target.Complete || id == 0 || target.Invalidated == ^uint32(0) {
		return nil
	}
	out := []syntax.Node{}
	for _, c := range ctx.Flow().Calls(syntax.EnclosingVariableScope(at)) {
		if !NativeDominates(c.Node, at) {
			continue
		}
		candidate := c.Receiver
		if candidate == nil {
			if call, ok := c.Node.(*syntax.FuncCall); ok {
				candidate = CallArgument(call.Args, 0, "")
			}
		}
		if candidate == nil {
			continue
		}
		value := ctx.Flow().Value(candidate)
		if !value.Complete || value.Identity != id || value.Invalidated != target.Invalidated {
			continue
		}
		for _, name := range names {
			if strings.EqualFold(c.Name, name) {
				out = append(out, c.Node)
				break
			}
		}
	}
	return out
}

func nativeFix(span syntax.Span, text, title string) diagnostic.Fix {
	return diagnostic.Fix{Title: title, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}

func nativeCleanEdit(ctx *analysis.Context, span syntax.Span) bool {
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

// nativeGlobalCallName avoids namespace declarations and imported call aliases.
func nativeGlobalCallName(ctx *analysis.Context, name string, at syntax.Node) string {
	fqn, _ := ctx.Names().Function(name, at.Span().Start)
	if !strings.EqualFold(fqn, name) {
		return `\` + name
	}
	return name
}
