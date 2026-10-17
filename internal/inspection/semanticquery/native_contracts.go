package semanticquery

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// stringOnly reports a parameter declared exactly `string` that null cannot
// be passed to: a null default makes the declared type nullable.
func stringOnly(p index.Param) bool {
	return p.Type == "string" && !strings.EqualFold(strings.TrimPrefix(strings.Trim(p.Default, "() \t\r\n"), `\`), "null")
}

// NativeStringInputs returns arguments bound to a resolved nonnullable string
// parameter. Unpack and unresolved contracts cannot establish a string use.
// The declaration read is the one the call resolves to: NativeStringConsumer
// tells whether it is the one that runs.
func NativeStringInputs(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList) []syntax.Expr {
	callee, ok := ResolveCallee(ctx, call)
	if !ok {
		return nil
	}
	var inputs []syntax.Expr
	for i, p := range callee.Params {
		if stringOnly(p) {
			if arg := CallArgument(list, i, p.Name); arg != nil {
				inputs = append(inputs, arg)
			}
		}
	}
	return inputs
}

// NativeStringConsumer confirms an argument NativeStringInputs returned:
// whichever declaration runs binds it to a parameter declared `string`. A
// function must be declared once. A method must be the one no other
// declaration of its class can stand for, and where a class of the project
// extending the receiver's replaces it, each replacement must declare the
// parameter the same (see flow.Env.Targets): at the same position, or
// wherever it stands when the argument is passed by name. Below PHP 7 a
// declaration of the project names a class with `string`.
func NativeStringConsumer(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList, input syntax.Expr) bool {
	callee, resolved := ResolveCallee(ctx, call)
	if !resolved || list == nil {
		return false
	}
	at := -1
	for i, p := range callee.Params {
		if stringOnly(p) && CallArgument(list, i, p.Name) == input {
			at = i
		}
	}
	if at < 0 {
		return false
	}
	name := callee.Params[at].Name
	named := slices.ContainsFunc(list.Args, func(n syntax.Expr) bool {
		arg, ok := n.(*syntax.Arg)
		return ok && arg.Value == input && arg.Name != nil
	})
	builtin := false
	if function, ok := call.(*syntax.FuncCall); ok {
		f := ctx.Types().ResolveFunction(function)
		if builtin = f.Builtin; !builtin && len(ctx.Index().FunctionDecls(f.FQN, ctx.PHP)) != 1 {
			return false
		}
	} else {
		targets := ctx.Flow().Targets(call)
		if len(targets) == 0 {
			return false
		}
		builtin = targets[0].Builtin
		for _, m := range targets {
			bound := at
			if named {
				bound = slices.IndexFunc(m.Params, func(p index.Param) bool { return p.Name == name && !p.Variadic })
			}
			if bound < 0 || bound >= len(m.Params) || !stringOnly(m.Params[bound]) || m.Params[bound].Name != name {
				return false
			}
		}
	}
	return builtin || ctx.PHP.AtLeast(phpversion.PHP70)
}

// NativeSuccessFollower finds an immediate unconditional true return after a
// discarded operation. It deliberately does not infer promises from names.
func NativeSuccessFollower(ctx *analysis.Context, call syntax.Node) bool {
	st, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	next, ok := astquery.NextStmt(ctx.File, st)
	if !ok {
		return false
	}
	ret, ok := next.(*syntax.Return)
	if !ok {
		return false
	}
	truth, known := astquery.BoolConst(ret.Expr)
	return known && truth
}
