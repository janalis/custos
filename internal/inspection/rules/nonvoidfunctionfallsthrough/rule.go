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
	if fallsThrough(ctx, body, nil, &budget) {
		ctx.ReportNode(typ, message)
	}
}

// alwaysThrows recognizes an expression that cannot complete:
// a throw, a plain assignment of one, or a match whose every arm throws (a
// subject matching no arm raises UnhandledMatchError).
func alwaysThrows(e syntax.Expr) bool {
	switch e := syntax.UnwrapParens(e).(type) {
	case *syntax.Throw:
		return true
	case *syntax.Assign:
		return e.Op.Kind == syntax.TEqual && alwaysThrows(e.Value)
	case *syntax.Match:
		for _, arm := range e.Arms {
			if !alwaysThrows(arm.Body) {
				return alwaysThrows(e.Cond)
			}
		}
		return true
	}
	return false
}

// boolParam reports whether v names a parameter declared
// exactly bool in its enclosing function.
func boolParam(ctx *analysis.Context, v *syntax.Variable) bool {
	for _, param := range syntax.VariableScopeParams(syntax.EnclosingVariableScope(v)) {
		if param.Var.Name == v.Name {
			return param.Type != nil && !param.Variadic && strings.EqualFold(ctx.Text(param.Type), "bool")
		}
	}
	return false
}

// canonicalCondition canonicalizes an if condition whose repeated
// evaluation yields the same result while its variables stay unwritten.
func canonicalCondition(ctx *analysis.Context, cond syntax.Expr) (predicate, bool) {
	var none predicate
	if v, ok := plainVariable(cond); ok {
		return predicate{op: syntax.TIsEqual, left: v}, true
	}
	binary, ok := syntax.UnwrapParens(cond).(*syntax.Binary)
	if !ok {
		return none, false
	}
	pred := predicate{op: binary.Op.Kind, left: syntax.UnwrapParens(binary.Left), right: syntax.UnwrapParens(binary.Right)}
	if _, negatable := negations[pred.op]; !negatable {
		return none, false
	}
	if semanticquery.IsLiteral(pred.left) {
		pred = pred.mirrored()
	}
	v, ok := plainVariable(pred.left)
	if !ok {
		return none, false
	}
	if _, variable := plainVariable(pred.right); !variable && !semanticquery.IsLiteral(pred.right) {
		return none, false
	}
	// A comparison with a boolean literal is a truthiness test: always when
	// loose, and when strict only for a value that is known to be a boolean.
	if b, isBool := astquery.BoolConst(pred.right); isBool {
		strict := pred.op == syntax.TIsIdentical || pred.op == syntax.TIsNotIdentical
		equality := strict || pred.op == syntax.TIsEqual || pred.op == syntax.TIsNotEqual
		if equality && (!strict || boolParam(ctx, v)) {
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

// fallsThrough proves at least one finite normal exit. Calls
// and unsupported control flow remove that candidate path from the proof, and
// so does a branch whose outcome an earlier branch on the same path decided.
func fallsThrough(ctx *analysis.Context, stmt syntax.Stmt, facts []predicate, budget *int) bool {
	*budget--
	if *budget < 0 || stmt == nil {
		return false
	}
	switch s := stmt.(type) {
	case *syntax.Block:
		for _, child := range s.Stmts {
			branch, isIf := child.(*syntax.If)
			if !isIf {
				if !fallsThrough(ctx, child, facts, budget) {
					return false
				}
				if writes(child) {
					facts = nil
				}
				continue
			}
			yes, no := ifExits(ctx, branch, facts, budget)
			if !yes && !no {
				return false
			}
			// Only the sides that reach the next statement shape what holds there.
			if writes(branch.Cond) || (yes && writes(branch.Body)) || (no && branch.Else != nil && writes(branch.Else)) {
				facts = nil
			} else if pred, comparable := canonicalCondition(ctx, branch.Cond); comparable && yes != no {
				if no {
					pred = pred.negated()
				}
				facts = append(slices.Clip(facts), pred)
			}
		}
		return true
	case *syntax.If:
		yes, no := ifExits(ctx, s, facts, budget)
		return yes || no
	case *syntax.ExprStmt:
		return !alwaysThrows(s.Expr) && !mayCall(s.Expr)
	case *syntax.Nop, *syntax.Echo:
		return !mayCall(s)
	default:
		return false
	}
}

// ifExits reports whether the then side and the else side of
// a branch can each reach the statement after it. facts are the predicates
// known to hold when the branch is evaluated.
func ifExits(ctx *analysis.Context, s *syntax.If, facts []predicate, budget *int) (yes, no bool) {
	if mayCall(s.Cond) || len(s.ElseIfs) > 0 {
		return false, false
	}
	pred, comparable := canonicalCondition(ctx, s.Cond)
	truth, known := semanticquery.NativeTruth(ctx, s.Cond)
	if !known && comparable {
		truth, known = implied(ctx, facts, pred)
	}
	if known {
		if truth {
			return fallsThrough(ctx, s.Body, facts, budget), false
		}
		return false, s.Else == nil || fallsThrough(ctx, s.Else.Body, facts, budget)
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
	yes = fallsThrough(ctx, s.Body, held, budget)
	no = s.Else == nil || fallsThrough(ctx, s.Else.Body, failed, budget)
	return yes, no
}

// implied decides a condition from the predicates that hold
// on the current path: an earlier branch on the same test or on its exact
// complement already fixed the outcome.
func implied(ctx *analysis.Context, facts []predicate, pred predicate) (truth, known bool) {
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

// predicate is a side-effect-free comparison in canonical
// form: a plain variable on the left, a plain variable or scalar literal on the
// right. A nil right operand stands for the boolean true, so `$v == true` and a
// bare `$v` share one form.
type predicate struct {
	op          syntax.TokenKind
	left, right syntax.Expr
}

// writes reports any assignment or increment, which may
// change a variable an earlier condition tested, directly or through an alias.
func writes(n syntax.Node) bool {
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

func (p predicate) mirrored() predicate {
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

func (p predicate) negated() predicate {
	p.op = negations[p.op]
	return p
}

func (p predicate) same(ctx *analysis.Context, q predicate) bool {
	if p.right == nil || q.right == nil {
		return p.op == q.op && p.right == nil && q.right == nil && astquery.Equivalent(ctx.File, p.left, q.left)
	}
	if p.op == q.op && astquery.Equivalent(ctx.File, p.left, q.left) && astquery.Equivalent(ctx.File, p.right, q.right) {
		return true
	}
	q = q.mirrored()
	return p.op == q.op && astquery.Equivalent(ctx.File, p.left, q.left) && astquery.Equivalent(ctx.File, p.right, q.right)
}

func mayCall(n syntax.Node) bool {
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

func plainVariable(e syntax.Expr) (*syntax.Variable, bool) {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return v, ok && v.NameExpr == nil
}

var negations = map[syntax.TokenKind]syntax.TokenKind{
	syntax.TIsEqual: syntax.TIsNotEqual, syntax.TIsNotEqual: syntax.TIsEqual,
	syntax.TIsIdentical: syntax.TIsNotIdentical, syntax.TIsNotIdentical: syntax.TIsIdentical,
	syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TLess,
	syntax.TGreater: syntax.TIsSmallerOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
}
