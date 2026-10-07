package probablebugs

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// issetArgumentExistence reports isset/empty/?? checks on a variable whose
// first mention in the function is that very check.
type issetArgumentExistence struct{}

func init() { register(issetArgumentExistence{}) }

func (issetArgumentExistence) ID() string { return "IssetArgumentExistence" }

func (issetArgumentExistence) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary, syntax.KIsset, syntax.KEmpty}
}

var issetSpecialNames = map[string]bool{
	"this": true, "_GET": true, "_POST": true, "_SESSION": true, "_REQUEST": true,
	"_FILES": true, "_COOKIE": true, "_ENV": true, "_SERVER": true, "GLOBALS": true,
	"HTTP_RAW_POST_DATA": true, "php_errormsg": true, "http_response_header": true,
}

func (r issetArgumentExistence) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Binary: // D1
		if x.Op.Kind == syntax.TCoalesce {
			r.candidate(ctx, x.Left)
		}
	case *syntax.Isset: // D2
		for _, v := range x.Vars {
			r.candidate(ctx, v)
		}
	case *syntax.Empty: // D3
		r.candidate(ctx, x.Expr)
	}
}

func (issetArgumentExistence) candidate(ctx *analysis.Context, e syntax.Expr) {
	v, ok := e.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name == "" || v.Span().Len() == 0 {
		return
	}
	name := v.Name
	if issetSpecialNames[name] { // D4
		return
	}
	fn := syntax.EnclosingFuncLike(v)
	body := syntax.FuncLikeBody(fn) // D6
	if fn == nil || body == nil {
		return
	}
	for _, p := range syntax.FuncLikeParams(fn) { // D5
		if p.Var != nil && p.Var.Name == name {
			return
		}
	}
	if c, ok := fn.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if u.Var != nil && u.Var.Name == name {
				return
			}
		}
	}

	// D7: first mention of name in the function's own scope.
	first := issetFirstMention(body, name)
	hasGoto, hasInclude, dynamic := false, false, false
	syntax.Inspect(body, func(c syntax.Node) bool {
		switch c := c.(type) {
		case *syntax.Goto:
			hasGoto = true
		case *syntax.Include:
			hasInclude = true
			if c.Span().Start < v.Span().Start { // D11: may define the variable
				dynamic = true
			}
		case *syntax.FuncCall: // D11: extract(), parse_str($s) define variables
			if ctx.IsGlobalFunctionCall(c, "extract") || (ctx.IsGlobalFunctionCall(c, "parse_str") && util.ArgCount(c) == 1) {
				dynamic = true
			}
		case *syntax.Assign: // D11: `$$name = …`
			if t, ok := c.Var.(*syntax.Variable); ok && t.NameExpr != nil {
				dynamic = true
			}
		}
		return true
	})
	if first == nil { // only in broken code (error recovery), e.g. `isset$x`
		return
	}
	if first != v {
		a, ok := first.Parent().(*syntax.Assign)
		if !ok || !isAncestor(a, v) {
			return
		}
	}
	if hasGoto || dynamic { // D9, D11
		return
	}
	if hasInclude && !ctx.Bool("IGNORE_INCLUDES") { // D10
		return
	}
	// D8: innermost loop enclosing the first mention.
	for p := first.Parent(); p != nil && p != fn; p = p.Parent() {
		switch p.(type) {
		case *syntax.For, *syntax.Foreach, *syntax.While, *syntax.DoWhile:
			if assignsPlainVar(p, name) {
				return
			}
			goto report
		}
	}
report:
	ctx.ReportSeverity(v.Span(), meta.SeverityError, "Variable '$"+name+"' is not defined in this scope.")
}

func isAncestor(anc, n syntax.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if p == anc {
			return true
		}
	}
	return false
}

// assignsPlainVar reports whether root contains an assignment whose target
// is the plain variable $name.
func assignsPlainVar(root syntax.Node, name string) bool {
	found := false
	syntax.Inspect(root, func(c syntax.Node) bool {
		if a, ok := c.(*syntax.Assign); ok {
			if t, ok := a.Var.(*syntax.Variable); ok && t.NameExpr == nil && t.Name == name {
				found = true
			}
		}
		return !found
	})
	return found
}

// issetFirstMention returns the first plain variable named name in body, in
// source order, staying in body's scope: closures contribute only their
// `use` list (which reads or binds the outer variable), arrow functions
// their body unless a parameter shadows name, and nested named functions
// and classes nothing.
func issetFirstMention(body syntax.Node, name string) *syntax.Variable {
	var first *syntax.Variable
	syntax.Inspect(body, func(c syntax.Node) bool {
		if first != nil {
			return false
		}
		switch c := c.(type) {
		case *syntax.Variable:
			if c.NameExpr == nil && c.Name == name && !util.IsStaticPropName(c) {
				first = c
				return false
			}
		case *syntax.Closure:
			for _, u := range c.Uses {
				if u.Var != nil && u.Var.Name == name {
					first = u.Var
					break
				}
			}
			return false
		case *syntax.ArrowFunction:
			for _, p := range c.Params {
				if p.Var != nil && p.Var.Name == name {
					return false
				}
			}
		case *syntax.Function, *syntax.ClassLike:
			return false
		}
		return true
	})
	return first
}
