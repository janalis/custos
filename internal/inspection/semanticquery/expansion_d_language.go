package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// CheckInterfaceInstantiation applies the independent interface contract.
func CheckInterfaceInstantiation(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	construct := n.(*syntax.New)
	decl := NativeNewClass(ctx, construct)
	if decl != nil && decl.Kind == syntax.KindInterface {
		ctx.ReportNode(construct, message)
	}
}

// CheckCatchCannotHandleKnownThrowable checks a directly known thrown type.
func CheckCatchCannotHandleKnownThrowable(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
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
	decl := NativeNewClass(ctx, construct)
	if decl == nil || !ctx.Index().IsSubtype(decl.FQN, "Error", ctx.PHP) {
		return
	}
	if constructor := ctx.Index().FindMethod(decl.FQN, "__construct", ctx.PHP); constructor != nil && !constructor.Builtin {
		return
	}
	if expansionLanguageMayCall(construct.Args) {
		return
	}
	for _, clause := range tr.Catches {
		for _, typ := range clause.Types {
			fqn := ctx.Names().Class(typ.Value, typ.Span().Start)
			if ctx.Index().Class(fqn, ctx.PHP) == nil || ctx.Index().IsSubtype(decl.FQN, fqn, ctx.PHP) {
				return
			}
		}
	}
	span := caught.Types[0].Span()
	span.End = caught.Types[len(caught.Types)-1].Span().End
	ctx.Report(span, message)
}

func expansionLanguageMayCall(n syntax.Node) bool {
	found := false
	if n == nil {
		return false
	}
	syntax.Inspect(n, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch n := n.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Exit:
			found = true
		case *syntax.Binary:
			found = n.Op.Kind == syntax.TDiv || n.Op.Kind == syntax.TMod
		}
		return !found
	})
	return found
}

// expansionLanguageFallsThrough proves at least one finite normal exit. Calls
// and unsupported control flow remove that candidate path from the proof.
func expansionLanguageFallsThrough(ctx *analysis.Context, stmt syntax.Stmt, budget *int) bool {
	*budget--
	if *budget < 0 || stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *syntax.Block:
		for _, child := range s.Stmts {
			if !expansionLanguageFallsThrough(ctx, child, budget) {
				return false
			}
		}
		return true
	case *syntax.If:
		if expansionLanguageMayCall(s.Cond) {
			return false
		}
		if len(s.ElseIfs) > 0 {
			return false
		}
		truth, known := NativeTruth(ctx, s.Cond)
		if known {
			if truth {
				return expansionLanguageFallsThrough(ctx, s.Body, budget)
			}
			if s.Else != nil {
				return expansionLanguageFallsThrough(ctx, s.Else.Body, budget)
			}
			return true
		}
		if !known {
			switch cond := syntax.UnwrapParens(s.Cond).(type) {
			case *syntax.Variable:
			case *syntax.Binary:
				switch cond.Op.Kind {
				case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
				default:
					return false
				}
				if astquery.Equivalent(ctx.File, cond.Left, cond.Right) || (NativeValue(ctx, cond.Left) != nil && NativeValue(ctx, cond.Right) != nil) {
					return false
				}
			default:
				return false
			}
		}
		yes := expansionLanguageFallsThrough(ctx, s.Body, budget)
		no := true
		if s.Else != nil {
			no = expansionLanguageFallsThrough(ctx, s.Else.Body, budget)
		}
		return yes || no
	case *syntax.ExprStmt:
		if _, ok := syntax.UnwrapParens(s.Expr).(*syntax.Throw); ok {
			return false
		}
		return !expansionLanguageMayCall(s.Expr)
	case *syntax.Nop, *syntax.Echo:
		return !expansionLanguageMayCall(s)
	default:
		return false
	}
}

// CheckNonVoidFunctionFallsThrough proves a typed function's normal exit.
func CheckNonVoidFunctionFallsThrough(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || ctx.PHP < phpversion.PHP70 || !NativeCallbackReachable(n) {
		return
	}
	var typ syntax.Expr
	var body *syntax.Block
	switch f := n.(type) {
	case *syntax.Function:
		typ, body = f.ReturnType, f.Body
	case *syntax.Method:
		typ, body = f.ReturnType, f.Body
	case *syntax.Closure:
		typ, body = f.ReturnType, f.Body
	}
	if typ == nil || body == nil || NativeGeneratorBody(body) {
		return
	}
	text := strings.ToLower(ctx.Text(typ))
	if text == "void" || text == "mixed" || text == "never" {
		return
	}
	budget := 512
	if expansionLanguageFallsThrough(ctx, body, &budget) {
		ctx.ReportNode(typ, message)
	}
}

// CheckAssertionContainsRequiredSideEffect finds an assertion-only definition.
func CheckAssertionContainsRequiredSideEffect(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !NativeBuiltin(ctx, c, "assert") {
		return
	}
	scope := syntax.EnclosingVariableScope(c)
	params := syntax.VariableScopeParams(scope)
	assignments := []*syntax.Assign{}
	syntax.Inspect(c.Args, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if a, ok := n.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
			assignments = append(assignments, a)
		}
		return true
	})
	for _, a := range assignments {
		v, ok := a.Var.(*syntax.Variable)
		if !ok {
			continue
		}
		parameter := false
		for _, p := range params {
			if p.Var.Name == v.Name {
				parameter = true
			}
		}
		if parameter {
			continue
		}
		earlier := false
		later := false
		nodes := 0
		visit := func(node syntax.Node) bool {
			if node != scope && syntax.IsVariableScope(node) {
				return false
			}
			nodes++
			if nodes > 512 {
				return false
			}
			if node.Span().Start >= c.Span().Start && node.Span().End <= c.Span().End {
				return false
			}
			if variable, ok := node.(*syntax.Variable); ok && variable.Name == v.Name {
				if variable.Span().Start < c.Span().Start {
					earlier = true
					return false
				}
				if variable.Span().Start > c.Span().End && NativeCallbackReachable(variable) {
					required := true
					for parent := variable.Parent(); parent != nil && parent != scope; parent = parent.Parent() {
						switch x := parent.(type) {
						case *syntax.Isset, *syntax.Empty, *syntax.Unset:
							required = false
						case *syntax.Binary:
							if x.Op.Kind == syntax.TCoalesce && x.Left.Span().Contains(variable.Span()) {
								required = false
							}
						}
						if _, ok := parent.(syntax.Stmt); ok {
							if parent.Parent() != c.Parent().Parent() {
								required = false
							}
							break
						}
					}
					if required && NativeDominates(c, variable) {
						later = true
					}
				}
			}
			if assign, ok := node.(*syntax.Assign); ok && assign.Span().Start > c.Span().End && astquery.MentionsVariable(assign.Var, v.Name) {
				if !later {
					earlier = true
				}
				return false
			}
			if node.Span().Start > c.Span().End {
				if u, ok := node.(*syntax.Unset); ok && astquery.MentionsVariable(u, v.Name) && !later {
					earlier = true
				}
				if a, ok := node.(*syntax.Assign); ok && a.ByRef && astquery.MentionsVariable(a.Value, v.Name) && !later {
					earlier = true
				}
			}
			if node.Span().Start < c.Span().Start {
				switch node.(type) {
				case *syntax.Global, *syntax.StaticStmt:
					if astquery.MentionsVariable(node, v.Name) {
						earlier = true
					}
				}
			}
			return !earlier
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		if !earlier && later && nodes <= 512 {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func expansionLanguageFiber(ctx *analysis.Context, c *syntax.MethodCall, name string) *syntax.New {
	if ctx.PHP < phpversion.PHP81 || !NativeMethod(ctx, c, "Fiber", name) {
		return nil
	}
	return NativeConstruction(ctx, c.Var, "Fiber")
}

func expansionLanguageFiberStarted(ctx *analysis.Context, c *syntax.MethodCall) bool {
	return len(expansionDMethodCalls(ctx, c, c.Var, "Fiber", "start")) > 0
}

// CheckFiberStartedTwice enforces one start per allocation.
func CheckFiberStartedTwice(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if expansionLanguageFiber(ctx, c, "start") != nil && expansionLanguageFiberStarted(ctx, c) {
		ctx.ReportNode(c, message)
	}
}

func expansionLanguageLiteral(e syntax.Expr) bool {
	switch value := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return value.Name != nil && (strings.EqualFold(value.Name.Value, "true") || strings.EqualFold(value.Name.Value, "false") || strings.EqualFold(value.Name.Value, "null"))
	}
	return false
}

// CheckFiberResumedAfterTermination recognizes a trivially completing callback.
func CheckFiberResumedAfterTermination(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	construct := expansionLanguageFiber(ctx, c, "resume")
	if construct == nil || !expansionLanguageFiberStarted(ctx, c) {
		return
	}
	params, body := NativeCallback(ctx, CallArgument(construct.Args, 0, "callback"))
	if len(params) > 0 {
		return
	}
	var value syntax.Expr
	if e, ok := body.(syntax.Expr); ok {
		value = e
	} else if b, ok := body.(*syntax.Block); ok && len(b.Stmts) == 1 {
		if r, ok := b.Stmts[0].(*syntax.Return); ok {
			value = r.Expr
		}
	}
	if expansionLanguageLiteral(value) {
		ctx.ReportNode(c, message)
	}
}

func expansionLanguageSuspends(ctx *analysis.Context, body syntax.Node) bool {
	if e, ok := body.(syntax.Expr); ok {
		return ExpansionCStatic(ctx, syntax.UnwrapParens(e), "Fiber", "suspend")
	}
	block, ok := body.(*syntax.Block)
	if !ok || len(block.Stmts) == 0 {
		return false
	}
	stmt, ok := block.Stmts[0].(*syntax.ExprStmt)
	return ok && ExpansionCStatic(ctx, syntax.UnwrapParens(stmt.Expr), "Fiber", "suspend")
}

// CheckFiberReturnBeforeTermination checks a definitely suspended callback.
func CheckFiberReturnBeforeTermination(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	construct := expansionLanguageFiber(ctx, c, "getReturn")
	if construct == nil || !expansionLanguageFiberStarted(ctx, c) {
		return
	}
	params, body := NativeCallback(ctx, CallArgument(construct.Args, 0, "callback"))
	if len(params) > 0 {
		return
	}
	if !expansionLanguageSuspends(ctx, body) {
		return
	}
	if len(expansionDMethodCalls(ctx, c, c.Var, "Fiber", "resume", "throw")) == 0 {
		ctx.ReportNode(c, message)
	}
}

// CheckFiberSuspendOutsideFiber uses an explicit current-fiber guard.
func CheckFiberSuspendOutsideFiber(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || ctx.PHP < phpversion.PHP81 || !NativeCallbackReachable(n) || !ExpansionCStatic(ctx, n.(syntax.Expr), "Fiber", "suspend") {
		return
	}
	for p := n.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(n.Span()) {
			continue
		}
		b, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
		if !ok || b.Op.Kind != syntax.TIsIdentical {
			continue
		}
		if (syntax.IsNullConst(syntax.UnwrapParens(b.Right)) && ExpansionCStatic(ctx, syntax.UnwrapParens(b.Left), "Fiber", "getCurrent")) || (syntax.IsNullConst(syntax.UnwrapParens(b.Left)) && ExpansionCStatic(ctx, syntax.UnwrapParens(b.Right), "Fiber", "getCurrent")) {
			ctx.ReportNode(n, message)
			return
		}
	}
}

// CheckGeneratorRewindAfterAdvance recognizes a finite local generator body.
func CheckGeneratorRewindAfterAdvance(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || ctx.PHP < phpversion.PHP55 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "rewind") {
		return
	}
	origin, ok := NativeLocalValue(ctx, c.Var).(*syntax.FuncCall)
	if !ok {
		return
	}
	_, body := NativeCallback(ctx, origin.Name)
	if body == nil {
		body = NativeFunctionBody(ctx, origin)
	}
	block, ok := body.(*syntax.Block)
	if !ok {
		return
	}
	yields := 0
	for _, stmt := range block.Stmts {
		expr, ok := stmt.(*syntax.ExprStmt)
		if !ok {
			return
		}
		if _, ok := expr.Expr.(*syntax.Yield); !ok {
			return
		}
		yields++
	}
	if yields < 2 {
		return
	}
	for _, event := range expansionDCalls(ctx, c) {
		prior, ok := event.(*syntax.MethodCall)
		if !ok || !NativeDominates(prior, c) || !expansionDSame(ctx, prior.Var, c.Var) {
			continue
		}
		name, ok := prior.Name.(*syntax.Identifier)
		if ok && (strings.EqualFold(name.Value, "next") || strings.EqualFold(name.Value, "send")) {
			method := ctx.Index().FindMethod("Generator", name.Value, ctx.PHP)
			if method != nil && method.Builtin {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}

// CheckIteratorCountChangesRequiredPosition tracks finite ArrayIterator state.
func CheckIteratorCountChangesRequiredPosition(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !NativeMethod(ctx, c, "ArrayIterator", "current") && !NativeMethod(ctx, c, "ArrayIterator", "key") {
		return
	}
	construct := NativeConstruction(ctx, c.Var, "ArrayIterator")
	if construct == nil || NativeArray(ctx, CallArgument(construct.Args, 0, "array")) == nil {
		return
	}
	counted := false
	for _, event := range expansionDCalls(ctx, c) {
		if !NativeDominates(event, c) {
			continue
		}
		switch p := event.(type) {
		case *syntax.FuncCall:
			if NativeBuiltin(ctx, p, "iterator_count") && expansionDSame(ctx, CallArgument(p.Args, 0, "iterator"), c.Var) {
				counted = true
			}
		case *syntax.MethodCall:
			if expansionDSame(ctx, p.Var, c.Var) && (NativeMethod(ctx, p, "ArrayIterator", "rewind") || NativeMethod(ctx, p, "ArrayIterator", "seek")) {
				counted = false
			}
		}
	}
	if counted {
		ctx.ReportNode(c, message)
	}
}
