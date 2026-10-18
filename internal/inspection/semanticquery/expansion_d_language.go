package semanticquery

import (
	"slices"
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
			if !expansionLanguageLiteral(value) {
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
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Exit, *syntax.Eval, *syntax.Clone:
			found = true
		case *syntax.InterpolatedString:
			found = found || n.Backtick
		case *syntax.Binary:
			found = found || n.Op.Kind == syntax.TDiv || n.Op.Kind == syntax.TMod
		}
		return !found
	})
	return found
}

// expansionLanguagePredicate is a side-effect-free comparison in canonical
// form: a plain variable on the left, a plain variable or scalar literal on the
// right. A nil right operand stands for the boolean true, so `$v == true` and a
// bare `$v` share one form.
type expansionLanguagePredicate struct {
	op          syntax.TokenKind
	left, right syntax.Expr
}

func expansionLanguagePlainVariable(e syntax.Expr) (*syntax.Variable, bool) {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return v, ok && v.NameExpr == nil
}

// expansionLanguageBoolParam reports whether v names a parameter declared
// exactly bool in its enclosing function.
func expansionLanguageBoolParam(ctx *analysis.Context, v *syntax.Variable) bool {
	for _, param := range syntax.VariableScopeParams(syntax.EnclosingVariableScope(v)) {
		if param.Var.Name == v.Name {
			return param.Type != nil && !param.Variadic && strings.EqualFold(ctx.Text(param.Type), "bool")
		}
	}
	return false
}

// expansionLanguageCondition canonicalizes an if condition whose repeated
// evaluation yields the same result while its variables stay unwritten.
func expansionLanguageCondition(ctx *analysis.Context, cond syntax.Expr) (expansionLanguagePredicate, bool) {
	var none expansionLanguagePredicate
	if v, ok := expansionLanguagePlainVariable(cond); ok {
		return expansionLanguagePredicate{op: syntax.TIsEqual, left: v}, true
	}
	binary, ok := syntax.UnwrapParens(cond).(*syntax.Binary)
	if !ok {
		return none, false
	}
	pred := expansionLanguagePredicate{op: binary.Op.Kind, left: syntax.UnwrapParens(binary.Left), right: syntax.UnwrapParens(binary.Right)}
	if _, negatable := expansionLanguageNegations[pred.op]; !negatable {
		return none, false
	}
	if expansionLanguageLiteral(pred.left) {
		pred = pred.mirrored()
	}
	v, ok := expansionLanguagePlainVariable(pred.left)
	if !ok {
		return none, false
	}
	if _, variable := expansionLanguagePlainVariable(pred.right); !variable && !expansionLanguageLiteral(pred.right) {
		return none, false
	}
	// A comparison with a boolean literal is a truthiness test: always when
	// loose, and when strict only for a value that is known to be a boolean.
	if b, isBool := astquery.BoolConst(pred.right); isBool {
		strict := pred.op == syntax.TIsIdentical || pred.op == syntax.TIsNotIdentical
		equality := strict || pred.op == syntax.TIsEqual || pred.op == syntax.TIsNotEqual
		if equality && (!strict || expansionLanguageBoolParam(ctx, v)) {
			pred.right = nil
			if (pred.op == syntax.TIsEqual || pred.op == syntax.TIsIdentical) == b {
				pred.op = syntax.TIsEqual
			} else {
				pred.op = syntax.TIsNotEqual
			}
		}
	}
	return pred, true
}

var expansionLanguageNegations = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TIsEqual: syntax.TIsNotEqual, syntax.TIsNotEqual: syntax.TIsEqual,
	syntax.TIsIdentical: syntax.TIsNotIdentical, syntax.TIsNotIdentical: syntax.TIsIdentical,
	syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TLess,
	syntax.TGreater: syntax.TIsSmallerOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
}

func (p expansionLanguagePredicate) negated() expansionLanguagePredicate {
	p.op = expansionLanguageNegations[p.op]
	return p
}

func (p expansionLanguagePredicate) mirrored() expansionLanguagePredicate {
	p.left, p.right = p.right, p.left
	switch p.op {
	case syntax.TLess:
		p.op = syntax.TGreater
	case syntax.TGreater:
		p.op = syntax.TLess
	case syntax.TIsSmallerOrEqual:
		p.op = syntax.TIsGreaterOrEqual
	case syntax.TIsGreaterOrEqual:
		p.op = syntax.TIsSmallerOrEqual
	}
	return p
}

func (p expansionLanguagePredicate) same(ctx *analysis.Context, q expansionLanguagePredicate) bool {
	if p.right == nil || q.right == nil {
		return p.op == q.op && p.right == nil && q.right == nil && astquery.Equivalent(ctx.File, p.left, q.left)
	}
	if p.op == q.op && astquery.Equivalent(ctx.File, p.left, q.left) && astquery.Equivalent(ctx.File, p.right, q.right) {
		return true
	}
	q = q.mirrored()
	return p.op == q.op && astquery.Equivalent(ctx.File, p.left, q.left) && astquery.Equivalent(ctx.File, p.right, q.right)
}

// expansionLanguageImplied decides a condition from the predicates that hold
// on the current path: an earlier branch on the same test or on its exact
// complement already fixed the outcome.
func expansionLanguageImplied(ctx *analysis.Context, facts []expansionLanguagePredicate, pred expansionLanguagePredicate) (truth, known bool) {
	for _, fact := range facts {
		if fact.same(ctx, pred) {
			return true, true
		}
		if fact.same(ctx, pred.negated()) {
			return false, true
		}
	}
	return false, false
}

// expansionLanguageWrites reports any assignment or increment, which may
// change a variable an earlier condition tested, directly or through an alias.
func expansionLanguageWrites(n syntax.Node) bool {
	found := false
	syntax.Inspect(n, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Assign, *syntax.IncDec:
			found = true
		}
		return !found
	})
	return found
}

// expansionLanguageAlwaysThrows recognizes an expression that cannot complete:
// a throw, a plain assignment of one, or a match whose every arm throws (a
// subject matching no arm raises UnhandledMatchError).
func expansionLanguageAlwaysThrows(e syntax.Expr) bool {
	switch e := syntax.UnwrapParens(e).(type) {
	case *syntax.Throw:
		return true
	case *syntax.Assign:
		return e.Op.Kind == syntax.TEqual && expansionLanguageAlwaysThrows(e.Value)
	case *syntax.Match:
		for _, arm := range e.Arms {
			if !expansionLanguageAlwaysThrows(arm.Body) {
				return expansionLanguageAlwaysThrows(e.Cond)
			}
		}
		return true
	}
	return false
}

// expansionLanguageIfExits reports whether the then side and the else side of
// a branch can each reach the statement after it. facts are the predicates
// known to hold when the branch is evaluated.
func expansionLanguageIfExits(ctx *analysis.Context, s *syntax.If, facts []expansionLanguagePredicate, budget *int) (yes, no bool) {
	if expansionLanguageMayCall(s.Cond) || len(s.ElseIfs) > 0 {
		return false, false
	}
	pred, comparable := expansionLanguageCondition(ctx, s.Cond)
	truth, known := NativeTruth(ctx, s.Cond)
	if !known && comparable {
		truth, known = expansionLanguageImplied(ctx, facts, pred)
	}
	if known {
		if truth {
			return expansionLanguageFallsThrough(ctx, s.Body, facts, budget), false
		}
		return false, s.Else == nil || expansionLanguageFallsThrough(ctx, s.Else.Body, facts, budget)
	}
	switch cond := syntax.UnwrapParens(s.Cond).(type) {
	case *syntax.Variable:
	case *syntax.Binary:
		switch cond.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
		default:
			return false, false
		}
		if astquery.Equivalent(ctx.File, cond.Left, cond.Right) || (NativeValue(ctx, cond.Left) != nil && NativeValue(ctx, cond.Right) != nil) {
			return false, false
		}
	default:
		return false, false
	}
	held, failed := facts, facts
	if comparable {
		held = append(slices.Clip(facts), pred)
		failed = append(slices.Clip(facts), pred.negated())
	}
	yes = expansionLanguageFallsThrough(ctx, s.Body, held, budget)
	no = s.Else == nil || expansionLanguageFallsThrough(ctx, s.Else.Body, failed, budget)
	return yes, no
}

// expansionLanguageFallsThrough proves at least one finite normal exit. Calls
// and unsupported control flow remove that candidate path from the proof, and
// so does a branch whose outcome an earlier branch on the same path decided.
func expansionLanguageFallsThrough(ctx *analysis.Context, stmt syntax.Stmt, facts []expansionLanguagePredicate, budget *int) bool {
	*budget--
	if *budget < 0 || stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *syntax.Block:
		for _, child := range s.Stmts {
			branch, isIf := child.(*syntax.If)
			if !isIf {
				if !expansionLanguageFallsThrough(ctx, child, facts, budget) {
					return false
				}
				if expansionLanguageWrites(child) {
					facts = nil
				}
				continue
			}
			yes, no := expansionLanguageIfExits(ctx, branch, facts, budget)
			if !yes && !no {
				return false
			}
			// Only the sides that reach the next statement shape what holds there.
			if expansionLanguageWrites(branch.Cond) || (yes && expansionLanguageWrites(branch.Body)) || (no && branch.Else != nil && expansionLanguageWrites(branch.Else)) {
				facts = nil
			} else if pred, comparable := expansionLanguageCondition(ctx, branch.Cond); comparable && yes != no {
				if no {
					pred = pred.negated()
				}
				facts = append(slices.Clip(facts), pred)
			}
		}
		return true
	case *syntax.If:
		yes, no := expansionLanguageIfExits(ctx, s, facts, budget)
		return yes || no
	case *syntax.ExprStmt:
		return !expansionLanguageAlwaysThrows(s.Expr) && !expansionLanguageMayCall(s.Expr)
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
	if expansionLanguageFallsThrough(ctx, body, nil, &budget) {
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
			if !later && node.Span().Start > c.Span().End && assertionRebindsLocal(ctx, node, v.Name) {
				earlier = true
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
						case *syntax.FuncCall:
							// A read in another assertion is compiled out with this one.
							if NativeBuiltin(ctx, x, "assert") {
								required = false
							}
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

// assertionOutParameters lists zero-based by-reference parameters that builtins
// overwrite without reading. In-out parameters (sort, array_push, the output
// list of exec, ...) are deliberately absent.
var assertionOutParameters = map[string][]int{
	"preg_match":             {2},
	"preg_match_all":         {2},
	"preg_replace":           {4},
	"preg_replace_callback":  {4},
	"str_replace":            {3},
	"str_ireplace":           {3},
	"similar_text":           {2},
	"parse_str":              {1},
	"mb_parse_str":           {1},
	"exec":                   {2},
	"system":                 {1},
	"passthru":               {1},
	"getimagesize":           {1},
	"getimagesizefromstring": {1},
	"headers_sent":           {0, 1},
	"getmxrr":                {1, 2},
	"is_callable":            {2},
	"fsockopen":              {2, 3},
}

// assertionRebindsLocal reports whether node gives the named local a new value
// without reading it: a foreach target, a catch variable or a positional
// write-only argument of a resolved builtin.
func assertionRebindsLocal(ctx *analysis.Context, node syntax.Node, name string) bool {
	switch x := node.(type) {
	case *syntax.Foreach:
		return !astquery.MentionsVariable(x.Expr, name) && (astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name))
	case *syntax.Catch:
		return x.Var != nil && x.Var.Name == name
	case *syntax.FuncCall:
		positions := assertionOutParameters[NativeBuiltinName(ctx, x)]
		if positions == nil {
			return false
		}
		written := false
		for i, expr := range x.Args.Args {
			arg, ok := expr.(*syntax.Arg)
			if !ok || arg.Name != nil || arg.Unpack {
				return false
			}
			target, ok := arg.Value.(*syntax.Variable)
			if ok && target.NameExpr == nil && target.Name == name && slices.Contains(positions, i) {
				written = true
			} else if astquery.MentionsVariable(arg.Value, name) {
				return false
			}
		}
		return written
	}
	return false
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
	if fiberStartCallable(c) || expansionLanguageFiber(ctx, c, "start") == nil {
		return
	}
	for _, first := range expansionDMethodCalls(ctx, c, c.Var, "Fiber", "start") {
		if !fiberStartCallable(first) && !fiberStartRebound(ctx, first, c) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// fiberStartCallable recognizes `$fiber->start(...)`, which only creates a
// closure and starts nothing.
func fiberStartCallable(c *syntax.MethodCall) bool {
	if c.Args == nil || len(c.Args.Args) != 1 {
		return false
	}
	_, callable := c.Args.Args[0].(*syntax.VariadicPlaceholder)
	return callable
}

// fiberStartRebound reports whether the receiver variable of second may hold
// another fiber than at first: an assignment target (destructuring included),
// a foreach key/value or a catch variable binds it in between.
func fiberStartRebound(ctx *analysis.Context, first, second *syntax.MethodCall) bool {
	name := ""
	if v, ok := syntax.UnwrapParens(second.Var).(*syntax.Variable); ok {
		name = v.Name
	}
	scope := syntax.EnclosingVariableScope(second)
	rebound := false
	visit := func(n syntax.Node) bool {
		if rebound || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		if n.Span().Start < first.Span().End || n.Span().Start >= second.Span().Start {
			return true
		}
		switch x := n.(type) {
		case *syntax.Assign:
			rebound = astquery.MentionsVariable(x.Var, name)
		case *syntax.Foreach:
			rebound = astquery.MentionsVariable(x.Key, name) || astquery.MentionsVariable(x.Value, name)
		case *syntax.Catch:
			rebound = x.Var != nil && x.Var.Name == name
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rebound
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

// fiberResumeCallableOnly recognizes `$fiber->method(...)`, which creates a
// closure without invoking the method.
func fiberResumeCallableOnly(c *syntax.MethodCall) bool {
	for _, arg := range c.Args.Args {
		if _, ok := arg.(*syntax.VariadicPlaceholder); ok {
			return true
		}
	}
	return false
}

// fiberResumeStarted requires an executed start() before the resume.
func fiberResumeStarted(ctx *analysis.Context, resume *syntax.MethodCall) bool {
	for _, start := range expansionDMethodCalls(ctx, resume, resume.Var, "Fiber", "start") {
		if !fiberResumeCallableOnly(start) {
			return true
		}
	}
	return false
}

// fiberResumeRebound reports whether the receiver may hold another fiber than
// construct. Between the allocation and the resume, a foreach target, a
// destructuring target, a catch or static variable or unset() gives the
// variable a new value; a reference to it (reference assignment, by-reference
// closure capture, `global`) anywhere before the resume lets other code do so.
func fiberResumeRebound(ctx *analysis.Context, resume *syntax.MethodCall, name string, construct *syntax.New) bool {
	from, to := construct.Span().End, resume.Span().Start
	mentions := func(target syntax.Node, anywhere bool) bool {
		found := false
		syntax.Inspect(target, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok && v.Name == name && v.Span().Start < to && (anywhere || v.Span().Start >= from) {
				found = true
			}
			return !found
		})
		return found
	}
	scope := syntax.EnclosingVariableScope(resume)
	rebound := false
	visit := func(n syntax.Node) bool {
		if rebound || n.Span().Start >= to {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			if x.ByRef {
				rebound = mentions(x, true)
				break
			}
			switch x.Var.(type) {
			case *syntax.Array, *syntax.List:
				rebound = mentions(x.Var, false)
			}
		case *syntax.Foreach:
			rebound = mentions(x.Value, false) || (x.Key != nil && mentions(x.Key, false))
		case *syntax.Catch:
			rebound = x.Var != nil && mentions(x.Var, false)
		case *syntax.StaticVar:
			rebound = mentions(x, false)
		case *syntax.Unset:
			rebound = mentions(x, false)
		case *syntax.Global:
			rebound = mentions(x, true)
		case *syntax.Closure:
			for _, use := range x.Uses {
				rebound = rebound || (use.ByRef && mentions(use, true))
			}
		}
		return !rebound && (n == scope || !syntax.IsVariableScope(n))
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return rebound
}

// CheckFiberResumedAfterTermination recognizes a trivially completing callback.
func CheckFiberResumedAfterTermination(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	construct := expansionLanguageFiber(ctx, c, "resume")
	receiver, local := syntax.UnwrapParens(c.Var).(*syntax.Variable)
	if construct == nil || !local || fiberResumeCallableOnly(c) || !fiberResumeStarted(ctx, c) || fiberResumeRebound(ctx, c, receiver.Name, construct) {
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

// CheckFiberReturnBeforeTermination checks a definitely suspended callback
// whose fiber is only allocated and started before the read.
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
	if expansionLanguageFiberUntouched(ctx, c, construct) {
		ctx.ReportNode(c, message)
	}
}

// expansionLanguageFiberUntouched proves that nothing but the allocation and
// dominating start calls names the receiver before the read at c executes.
// Any other mention (a resume in a loop or branch, a state guard, an alias, an
// argument, a capture) may complete the fiber, so it abandons the proof.
func expansionLanguageFiberUntouched(ctx *analysis.Context, c *syntax.MethodCall, construct *syntax.New) bool {
	// The allocation only resolves for a plain variable receiver.
	read := syntax.UnwrapParens(c.Var)
	receiver, _ := astquery.PlainVariableName(read)
	allowed := []syntax.Node{read}
	for _, start := range expansionDMethodCalls(ctx, c, c.Var, "Fiber", "start") {
		allowed = append(allowed, syntax.UnwrapParens(start.Var))
	}
	scope := syntax.EnclosingVariableScope(c)
	// Code after the read cannot run first, unless a loop repeats it.
	limit := c.Span().End
	for p := c.Parent(); p != nil && p != scope; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			limit = p.Span().End
		}
	}
	untouched := true
	visit := func(n syntax.Node) bool {
		if !untouched || n.Span().Start >= limit {
			return false
		}
		switch x := n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.PropertyHook, *syntax.Closure:
			// Own variables; a captured fiber already fails the alias proof.
			return false
		case *syntax.FuncCall:
			// These expose every local by name without mentioning it.
			if name, ok := x.Name.(*syntax.Name); ok && slices.Contains([]string{"compact", "extract", "get_defined_vars"}, strings.ToLower(strings.TrimPrefix(name.Value, `\`))) {
				untouched = false
			}
		case *syntax.Variable:
			if x.NameExpr != nil {
				untouched = false
			} else if x.Name == receiver && !slices.Contains(allowed, syntax.Node(x)) {
				assign, ok := x.Parent().(*syntax.Assign)
				if !ok || assign.Var != syntax.Expr(x) || syntax.UnwrapParens(assign.Value) != syntax.Expr(construct) {
					untouched = false
				}
			}
		}
		return untouched
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Children(scope, func(child syntax.Node) { syntax.Inspect(child, visit) })
	}
	return untouched
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
	receiver, ok := syntax.UnwrapParens(c.Var).(*syntax.Variable)
	if !ok || receiver.NameExpr != nil {
		return
	}
	construct := NativeConstruction(ctx, c.Var, "ArrayIterator")
	if construct == nil {
		return
	}
	// Counting an empty iterator leaves the same invalid position it started at.
	if items := NativeArray(ctx, CallArgument(construct.Args, 0, "array")); items == nil || len(items.Items) == 0 {
		return
	}
	var counted *syntax.FuncCall
	for _, event := range expansionDCalls(ctx, c) {
		p, ok := event.(*syntax.FuncCall)
		if ok && NativeDominates(p, c) && NativeBuiltin(ctx, p, "iterator_count") && expansionDSame(ctx, CallArgument(p.Args, 0, "iterator"), c.Var) {
			counted = p
		}
	}
	if counted != nil && !iteratorCountPositionDisturbed(ctx, receiver.Name, construct, counted, c) {
		ctx.ReportNode(c, message)
	}
}

// iteratorCountPositionDisturbed reports whether the iterator variable is used
// between its construction and the read in a way that may move the cursor
// after the last count or hand the object to other code. Only iterator_count
// arguments and ArrayIterator calls that keep an exhausted cursor exhausted
// are transparent; rewind() and seek() are transparent before the last count
// only. A foreach subject, a loop header, a wrapper constructor, an alias, a
// closure capture or any other mention ends the proof.
func iteratorCountPositionDisturbed(ctx *analysis.Context, name string, construct *syntax.New, counted *syntax.FuncCall, read *syntax.MethodCall) bool {
	scope := syntax.EnclosingVariableScope(read)
	transparent := map[syntax.Node]bool{}
	disturbed := false
	visit := func(n syntax.Node) bool {
		if disturbed {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			// Captured objects stay shared with the enclosing scope.
		case *syntax.FuncCall:
			if NativeBuiltin(ctx, x, "iterator_count") {
				if arg := CallArgument(x.Args, 0, "iterator"); arg != nil {
					transparent[syntax.UnwrapParens(arg)] = true
				}
			}
		case *syntax.MethodCall:
			if iteratorCountKeepsPosition(ctx, x, x.Span().Start < counted.Span().Start) {
				transparent[syntax.UnwrapParens(x.Var)] = true
			}
		case *syntax.Variable:
			if x.NameExpr == nil && x.Name == name && x.Span().Start >= construct.Span().End && x.Span().Start < read.Span().Start && !transparent[n] {
				disturbed = true
			}
		default:
			if n != scope && syntax.IsVariableScope(n) {
				return false
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return disturbed
}

// iteratorCountKeepsPosition recognizes ArrayIterator calls that cannot make an
// exhausted cursor valid again. Repositioning calls qualify only when a later
// count consumes the iterator anyway.
func iteratorCountKeepsPosition(ctx *analysis.Context, c *syntax.MethodCall, beforeCount bool) bool {
	for _, method := range []string{"current", "key", "valid", "next", "count"} {
		if NativeMethod(ctx, c, "ArrayIterator", method) {
			return true
		}
	}
	return beforeCount && (NativeMethod(ctx, c, "ArrayIterator", "rewind") || NativeMethod(ctx, c, "ArrayIterator", "seek"))
}
