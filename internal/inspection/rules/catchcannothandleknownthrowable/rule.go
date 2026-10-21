// Package catchcannothandleknownthrowable implements CatchCannotHandleKnownThrowable.
package catchcannothandleknownthrowable

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Catch a type that handles the thrown value."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CatchCannotHandleKnownThrowable" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KCatch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	caught := n.(*syntax.Catch)
	tr, ok := caught.Parent().(*syntax.Try)
	if !ok || tr.Finally != nil || len(tr.Body.Stmts) != 1 || len(caught.Types) == 0 {
		return
	}
	stmt, ok := tr.Body.Stmts[0].(*syntax.ExprStmt)
	if !ok {
		return
	}
	throw, ok := stmt.Expr.(*syntax.Throw)
	if !ok {
		return
	}
	construct, ok := syntax.UnwrapParens(throw.Expr).(*syntax.New)
	if !ok {
		return
	}
	decl := semanticquery.NativeNewClass(ctx, construct)
	if decl == nil || !ctx.Index().IsSubtype(decl.FQN, "Error", ctx.PHP) {
		return
	}
	if constructor := ctx.Index().FindMethod(decl.FQN, "__construct", ctx.PHP); constructor != nil && !constructor.Builtin {
		return
	}
	inert, bound := catchKnownThrowableArgs(construct.Args)
	if !inert {
		return
	}
	for _, clause := range tr.Catches {
		for _, typ := range clause.Types {
			fqn := ctx.Names().Class(typ.Value, typ.Span().Start)
			if ctx.Index().Class(fqn, ctx.PHP) == nil || ctx.Index().IsSubtype(decl.FQN, fqn, ctx.PHP) {
				return
			}
			// Arguments the constructor may reject raise TypeError or
			// ArgumentCountError before the object exists.
			if !bound && (ctx.Index().IsSubtype("TypeError", fqn, ctx.PHP) || ctx.Index().IsSubtype("ArgumentCountError", fqn, ctx.PHP)) {
				return
			}
		}
	}
	span := caught.Types[0].Span()
	span.End = caught.Types[len(caught.Types)-1].Span().End
	ctx.Report(span, message)
}

// catchKnownThrowableArgs classifies constructor arguments. inert means their
// evaluation cannot raise anything: only scalar literals, magic constants and
// plain variables qualify. bound additionally means the builtin constructor
// accepts them: no argument, or one positional string literal.
func catchKnownThrowableArgs(list *syntax.ArgList) (inert, bound bool) {
	if list == nil {
		return true, true
	}
	for _, item := range list.Args {
		arg, ok := item.(*syntax.Arg)
		if !ok || arg.Unpack || arg.ByRef {
			return false, false
		}
		switch value := syntax.UnwrapParens(arg.Value).(type) {
		case *syntax.Literal, *syntax.MagicConst:
		case *syntax.ConstFetch:
			if !semanticquery.IsLiteral(value) {
				return false, false
			}
		case *syntax.Variable:
			if value.NameExpr != nil {
				return false, false
			}
		default:
			return false, false
		}
	}
	if len(list.Args) == 0 {
		return true, true
	}
	first := list.Args[0].(*syntax.Arg)
	lit, ok := syntax.UnwrapParens(first.Value).(*syntax.Literal)
	return true, len(list.Args) == 1 && first.Name == nil && ok && lit.LitKind == syntax.LitString
}
