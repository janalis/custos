// Package nonvoidfunctionfallsthrough implements NonVoidFunctionFallsThrough.
package nonvoidfunctionfallsthrough

import (
	"slices"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Return a value on every reachable path."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "NonVoidFunctionFallsThrough" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || ctx.PHP < phpversion.PHP70 || !semanticquery.NativeCallbackReachable(ctx, n) {
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
	if typ == nil || body == nil || semanticquery.NativeGeneratorBody(body) {
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
	if semanticquery.IsLiteral(pred.left) {
		pred = pred.mirrored()
	}
	v, ok := expansionLanguagePlainVariable(pred.left)
	if !ok {
		return none, false
	}
	if _, variable := expansionLanguagePlainVariable(pred.right); !variable && !semanticquery.IsLiteral(pred.right) {
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

// expansionLanguageIfExits reports whether the then side and the else side of
// a branch can each reach the statement after it. facts are the predicates
// known to hold when the branch is evaluated.
func expansionLanguageIfExits(ctx *analysis.Context, s *syntax.If, facts []expansionLanguagePredicate, budget *int) (yes, no bool) {
	if expansionLanguageMayCall(s.Cond) || len(s.ElseIfs) > 0 {
		return false, false
	}
	pred, comparable := expansionLanguageCondition(ctx, s.Cond)
	truth, known := semanticquery.NativeTruth(ctx, s.Cond)
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
		if astquery.Equivalent(ctx.File, cond.Left, cond.Right) || (semanticquery.NativeValue(ctx, cond.Left) != nil && semanticquery.NativeValue(ctx, cond.Right) != nil) {
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

// expansionLanguagePredicate is a side-effect-free comparison in canonical
// form: a plain variable on the left, a plain variable or scalar literal on the
// right. A nil right operand stands for the boolean true, so `$v == true` and a
// bare `$v` share one form.
type expansionLanguagePredicate struct {
	op          syntax.TokenKind
	left, right syntax.Expr
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

func (p expansionLanguagePredicate) negated() expansionLanguagePredicate {
	p.op = expansionLanguageNegations[p.op]
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

func expansionLanguagePlainVariable(e syntax.Expr) (*syntax.Variable, bool) {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return v, ok && v.NameExpr == nil
}

var expansionLanguageNegations = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TIsEqual: syntax.TIsNotEqual, syntax.TIsNotEqual: syntax.TIsEqual,
	syntax.TIsIdentical: syntax.TIsNotIdentical, syntax.TIsNotIdentical: syntax.TIsIdentical,
	syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TLess,
	syntax.TGreater: syntax.TIsSmallerOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
}
