package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// ExpansionCConstant matches a resolved builtin class constant, including aliases.
func ExpansionCConstant(ctx *analysis.Context, e syntax.Expr, class, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ClassConstFetch)
	if !ok {
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || id.Value != name || !strings.EqualFold(StaticCallClass(ctx, c.Class), class) {
		return false
	}
	decl := ctx.Index().Class(class, ctx.PHP)
	return decl != nil && strings.HasPrefix(decl.File, "stubs/")
}

// ExpansionCStatic recognizes a builtin static method without treating lookalikes as APIs.
func ExpansionCStatic(ctx *analysis.Context, e syntax.Expr, class, method string) bool {
	c, ok := e.(*syntax.StaticCall)
	if !ok || !strings.EqualFold(StaticCallClass(ctx, c.Class), class) {
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, method) {
		return false
	}
	m := ctx.Index().FindMethod(class, method, ctx.PHP)
	return m != nil && m.Builtin && m.Avail.In(ctx.PHP)
}

// ExpansionCTruthy identifies a direct boolean condition without assuming branch intent.
func ExpansionCTruthy(e syntax.Expr) bool {
	p, child := astquery.ParentSkipParens(e)
	if a, ok := p.(*syntax.Assign); ok && a.Value == child {
		p, child = astquery.ParentSkipParens(a)
	}
	if u, ok := p.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
		p, child = astquery.ParentSkipParens(u)
	}
	switch c := p.(type) {
	case *syntax.If:
		return c.Cond == child
	case *syntax.While:
		return c.Cond == child
	case *syntax.DoWhile:
		return c.Cond == child
	}
	return false
}

// ExpansionCOutputCall locates the most recent dominating out-parameter producer.
func ExpansionCOutputCall(ctx *analysis.Context, at syntax.Node, value syntax.Expr, functions []string, position int, name string) *syntax.FuncCall {
	v, ok := syntax.UnwrapParens(value).(*syntax.Variable)
	if !ok {
		return nil
	}
	scope := syntax.EnclosingVariableScope(at)
	calls := ctx.Flow().Calls(scope)
	if len(calls) > 256 {
		return nil
	}
	var found *syntax.FuncCall
	for _, fact := range calls {
		c, ok := fact.Node.(*syntax.FuncCall)
		if !ok || !NativeDominates(c, at) {
			continue
		}
		out, ok := CallArgument(c.Args, position, name).(*syntax.Variable)
		if !ok || out.Name != v.Name {
			continue
		}
		builtin := NativeBuiltinName(ctx, c)
		for _, name := range functions {
			if builtin == name && (found == nil || found.Span().Start < c.Span().Start) {
				found = c
			}
		}
	}
	if found == nil {
		return nil
	}
	valid := true
	visited := 0
	visit := func(n syntax.Node) bool {
		visited++
		if visited > 256 {
			valid = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		if n != scope && syntax.EnclosingVariableScope(n) != scope {
			return false
		}
		if n.Span().Start > found.Span().End {
			if a, ok := n.(*syntax.Assign); ok {
				if target, ok := a.Var.(*syntax.Variable); ok && target.Name == v.Name {
					valid = false
				}
			}
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		for _, stmt := range ctx.File.Stmts {
			if !valid || visited > 256 || stmt.Span().Start >= at.Span().Start {
				break
			}
			syntax.Inspect(stmt, visit)
		}
	}
	for _, fact := range calls {
		if fact.Node.Span().Start <= found.Span().End || fact.Node.Span().Start >= at.Span().Start {
			continue
		}
		var args *syntax.ArgList
		switch call := fact.Node.(type) {
		case *syntax.FuncCall:
			switch NativeBuiltinName(ctx, call) {
			case "pcntl_wifexited", "pcntl_wifsignaled", "pcntl_wifstopped", "pcntl_wexitstatus", "pcntl_wtermsig", "pcntl_wstopsig":
				continue
			}
			args = call.Args
		case *syntax.MethodCall:
			args = call.Args
		case *syntax.StaticCall:
			args = call.Args
		}
		if args != nil {
			for _, node := range args.Args {
				if arg, ok := node.(*syntax.Arg); ok && ExpansionCSame(ctx, arg.Value, value) {
					valid = false
				}
			}
		}
	}
	if !valid {
		return nil
	}
	return found
}

// ExpansionCSame compares stable source expressions; useful for output argument identities.
func ExpansionCSame(ctx *analysis.Context, a, b syntax.Expr) bool {
	return a != nil && b != nil && astquery.Equivalent(ctx.File, a, b)
}
