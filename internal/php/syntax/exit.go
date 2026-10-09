package syntax

// ExitInvokes reports whether evaluating exit/die calls the function rather
// than producing its first-class callable.
func ExitInvokes(n *Exit) bool {
	return exitArgsInvoke(n.Args)
}

// ExitInvocation reports a direct exit/die call, including an immediate
// invocation of its first-class callable. Creating a callable does not exit.
func ExitInvocation(e Expr) bool {
	switch n := UnwrapParens(e).(type) {
	case *Exit:
		return ExitInvokes(n)
	case *FuncCall:
		callee, ok := UnwrapParens(n.Name).(*Exit)
		return ok && !ExitInvokes(callee) && exitArgsInvoke(n.Args)
	}
	return false
}

func exitArgsInvoke(args *ArgList) bool {
	if args == nil || len(args.Args) != 1 {
		return true
	}
	_, callable := args.Args[0].(*VariadicPlaceholder)
	return !callable
}
