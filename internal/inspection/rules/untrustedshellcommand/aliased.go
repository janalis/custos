package untrustedshellcommand

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// aliased reports whether the scope of v can rebind it without an assignment
// its resolved value accounts for: a reference to it (alias, by-reference
// capture, loop or argument), a catch binding or static declaration of the
// same name, extract(), an assignment through a variable variable or the
// global symbol table. The
// structural proofs of this rule follow a variable to its reaching
// assignment, which such a rebinding leaves in place.
func aliased(ctx *analysis.Context, v *syntax.Variable) bool {
	found := false
	// root is the variable an expression reads or writes an element of.
	root := func(e syntax.Expr) *syntax.Variable {
		for {
			switch x := syntax.UnwrapParens(e).(type) {
			case *syntax.ArrayDimFetch:
				e = x.Var
				continue
			case *syntax.Variable:
				return x
			}
			return &syntax.Variable{}
		}
	}
	named := func(e syntax.Expr) bool { return root(e).Name == v.Name }
	scope := syntax.EnclosingVariableScope(v)
	local := func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			found = found || root(x.Var).NameExpr != nil || x.ByRef && (named(x.Var) || named(x.Value))
		case *syntax.Foreach:
			found = found || x.ByRef && named(x.Expr)
		case *syntax.Closure:
			for _, use := range x.Uses {
				found = found || use.ByRef && use.Var.Name == v.Name
			}
		case *syntax.Catch:
			found = found || x.Var != nil && x.Var.Name == v.Name
		case *syntax.StaticVar:
			found = found || x.Var.Name == v.Name
		case *syntax.FuncCall:
			found = found || semanticquery.NativeBuiltinName(ctx, x) == "extract"
			if callee, ok := semanticquery.ResolveCallee(ctx, x); ok && x.Args != nil {
				for i, node := range x.Args.Args {
					arg, isArg := node.(*syntax.Arg)
					found = found || isArg && i < len(callee.Params) && callee.Params[i].ByRef && named(arg.Value)
				}
			}
		}
		return !found && (n == scope || !syntax.IsVariableScope(n))
	}
	// The global symbol table is shared with every function of the file.
	shared := func(n syntax.Node) bool {
		if x, ok := n.(*syntax.Variable); ok {
			_, global := x.Parent().(*syntax.Global)
			entry, known := "", false
			if fetch, ok := x.Parent().(*syntax.ArrayDimFetch); ok && fetch.Var == syntax.Expr(x) && fetch.Dim != nil {
				entry, known = semanticquery.NativeString(ctx, fetch.Dim)
			}
			found = found || x.Name == "GLOBALS" && (!known || entry == v.Name) || global && x.Name == v.Name
		}
		return !found
	}
	if scope != nil {
		syntax.Inspect(scope, local)
		syntax.Inspect(scope, shared)
	} else {
		syntax.InspectFile(ctx.File, local)
		syntax.InspectFile(ctx.File, shared)
	}
	return found
}
