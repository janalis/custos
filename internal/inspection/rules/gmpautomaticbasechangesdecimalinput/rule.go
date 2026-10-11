// Package gmpautomaticbasechangesdecimalinput implements the native GmpAutomaticBaseChangesDecimalInput inspection.
package gmpautomaticbasechangesdecimalinput

import (
	"math/big"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Specify decimal base for zero-prefixed GMP input."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GmpAutomaticBaseChangesDecimalInput" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "gmp_init") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "num")
	s, ok := semanticquery.NativeString(ctx, arg)
	if !ok || len(s) < 2 || len(s) > 1024 || s[0] != '0' || strings.Trim(s, "01234567") != "" {
		return
	}
	octal, _ := new(big.Int).SetString(s, 8)
	decimal, _ := new(big.Int).SetString(s, 10)
	if octal.Cmp(decimal) == 0 || stale(ctx, c, arg) {
		return
	}
	base := semanticquery.CallArgument(c.Args, 1, "base")
	if base != nil {
		v, known := semanticquery.NativeInt(ctx, base)
		if !known || v != 0 {
			return
		}
	}
	ctx.ReportNode(c, message, fixes(ctx, c, arg, base)...)
}

// fixes builds the decimal-base edit. It is withheld when the digits come from
// local flow or a late-bound constant: adding a decimal base to a value that
// turns out to be prefixed would throw.
func fixes(ctx *analysis.Context, c *syntax.FuncCall, arg, base syntax.Expr) []diagnostic.Fix {
	if !proven(arg) {
		return nil
	}
	if base != nil {
		if literal, ok := syntax.UnwrapParens(base).(*syntax.Literal); ok {
			return []diagnostic.Fix{astquery.ReplaceFix(literal.Span(), "10")}
		}
		return nil
	}
	if len(c.Args.Args) != 1 {
		return nil
	}
	a, ok := c.Args.Args[0].(*syntax.Arg)
	if !ok || a.Unpack || strings.ContainsAny(ctx.SpanText(syntax.Span{Start: a.Span().End, End: c.Args.Span().End}), "/#") {
		return nil
	}
	suffix := ", 10"
	if a.Name != nil {
		suffix = ", base: 10"
	}
	at := a.Span().End
	return []diagnostic.Fix{astquery.ReplaceFix(syntax.Span{Start: at, End: at}, suffix)}
}

// proven reports whether the digits are fixed by the argument expression
// itself: no variable and no `static::` constant, which a subclass may
// redefine.
func proven(arg syntax.Expr) bool {
	ok := true
	syntax.Inspect(arg, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable:
			ok = false
		case *syntax.ClassConstFetch:
			if class, isName := n.Class.(*syntax.Name); isName && strings.EqualFold(class.Value, "static") {
				ok = false
			}
		}
		return ok
	})
	return ok
}

// stale reports whether a string read from a variable may no longer be the
// assigned literal: the variable, or one whose value was copied into it, is
// changed by `++`/`--`, bound by a catch clause, bound by reference in a closure `use` or an array
// item, or the scope imports variables with extract() or one-argument
// parse_str(). File-level variables are also globals, so a `global`
// declaration or any `$GLOBALS` access in the file counts too.
func stale(ctx *analysis.Context, call *syntax.FuncCall, arg syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(arg).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(call)
	hidden := map[string]bool{}    // names written outside plain assignments
	feeds := map[string][]string{} // assigned name -> variables read by the values
	visit := func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.ClassLike, *syntax.ArrowFunction:
			return n == scope
		case *syntax.Closure:
			for _, use := range n.Uses {
				if use.ByRef && use.Var != nil {
					hidden[use.Var.Name] = true
				}
			}
			return n == scope
		case *syntax.IncDec:
			if target, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok {
				hidden[target.Name] = true
			}
		case *syntax.Catch:
			if n.Var != nil {
				hidden[n.Var.Name] = true
			}
		case *syntax.ArrayItem:
			if target, ok := n.Value.(*syntax.Variable); ok && n.ByRef {
				hidden[target.Name] = true
			}
		case *syntax.FuncCall:
			switch ctx.GlobalFunctionName(n) {
			case "extract":
				hidden[variable.Name] = true
			case "parse_str":
				if n.Args != nil && len(n.Args.Args) == 1 {
					hidden[variable.Name] = true
				}
			}
		case *syntax.Assign:
			if target, ok := n.Var.(*syntax.Variable); ok {
				syntax.Inspect(n.Value, func(x syntax.Node) bool {
					if read, ok := x.(*syntax.Variable); ok {
						feeds[target.Name] = append(feeds[target.Name], read.Name)
					}
					return true
				})
			}
		}
		return true
	}
	if scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.Global:
				for _, x := range n.Vars {
					syntax.Inspect(x, func(v syntax.Node) bool {
						if global, ok := v.(*syntax.Variable); ok {
							hidden[global.Name] = true
						}
						return true
					})
				}
			case *syntax.Variable:
				if n.Name == "GLOBALS" {
					hidden[variable.Name] = true
				}
			}
			return true
		})
	}
	seen := map[string]bool{variable.Name: true}
	for queue := []string{variable.Name}; len(queue) > 0; queue = queue[1:] {
		if hidden[queue[0]] {
			return true
		}
		for _, name := range feeds[queue[0]] {
			if !seen[name] {
				seen[name] = true
				queue = append(queue, name)
			}
		}
	}
	return false
}
