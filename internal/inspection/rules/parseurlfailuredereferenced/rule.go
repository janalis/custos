// Package parseurlfailuredereferenced implements the native ParseUrlFailureDereferenced inspection.
package parseurlfailuredereferenced

import (
	"net/url"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject URL parsing failure before indexing."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ParseUrlFailureDereferenced" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, a.Var, n) && !semanticquery.NativeFailureFlowOnly(ctx, a.Var) {
		return
	}
	// Among several reaching assignments only a parse_url result counts:
	// an unguarded result of another call does not stand in for it.
	var c *syntax.FuncCall
	for _, candidate := range semanticquery.NativeLocalBuiltinResults(ctx, a.Var, "false", "parse_url") {
		if call, ok := candidate.(*syntax.FuncCall); ok && relevantParse(ctx, call) {
			c = call
			break
		}
	}
	ok := c != nil
	if !ok {
		// The lexical scan only reads assignments that are statements; the
		// flow also knows a result assigned inside a condition.
		c, ok = semanticquery.NativeValue(ctx, a.Var).(*syntax.FuncCall)
	}
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "parse_url") {
		return
	}
	if !relevantParse(ctx, c) {
		return
	}

	if !semanticquery.NativeFailureFlowOnly(ctx, a.Var) && semanticquery.NativeLocalSentinelGuard(ctx, a.Var, "false") || ctx.Flow().Excludes(a.Var, "false") {
		return
	}
	if semanticquery.NativeFailureFlowOnly(ctx, a.Var) || !(guard{ctx: ctx, target: a.Var, after: c.Span().End}).covers() {
		ctx.ReportNode(a, message)
	}
}

// relevantParse excludes other result contracts before choosing a producer,
// so a safe literal or component result cannot hide another failing parse.
func relevantParse(ctx *analysis.Context, c *syntax.FuncCall) bool {
	if component, state := semanticquery.CallArgumentBinding(c.Args, 1, "component"); state != semanticquery.ArgumentMissing {
		v, known := semanticquery.NativeInt(ctx, component)
		if !known || v != -1 {
			return false
		}
	}
	raw, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "url"))
	return !known || literalURLFails(raw)
}

// guard proves lexically that the failure value cannot reach a read of target.
// Only conditions written after the producing call are considered.
type guard struct {
	ctx    *analysis.Context
	target syntax.Expr
	after  uint32
}

// covers walks the enclosing conditions and the earlier terminating branches
// of every enclosing statement list, up to the variable scope.
func (g guard) covers() bool {
	switch syntax.UnwrapParens(g.target).(type) {
	case *syntax.Variable, *syntax.PropertyFetch:
	default:
		return false
	}
	var child syntax.Node = g.target
	for p := g.target.Parent(); p != nil && !syntax.IsVariableScope(p); child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Ternary:
			if (child == syntax.Node(x.Then) && g.proves(x.Cond, true)) || (child == syntax.Node(x.Else) && g.proves(x.Cond, false)) {
				return true
			}
		case *syntax.Binary:
			if child == syntax.Node(x.Right) && g.shortCircuits(x) {
				return true
			}
		case *syntax.If:
			if g.branch(x, g.target.Span()) {
				return true
			}
		}
		st, ok := p.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, listed := astquery.StmtList(g.ctx.File, st)
		if !listed {
			continue
		}
		for _, prev := range list[:i] {
			if exit, ok := prev.(*syntax.If); ok && syntax.Terminates(exit.Body) && g.proves(exit.Cond, false) {
				return true
			}
		}
	}
	return false
}

// shortCircuits reports whether the right operand only runs once the left
// operand has excluded the failure value.
func (g guard) shortCircuits(b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		return g.proves(b.Left, true)
	case syntax.TBooleanOr, syntax.TOr:
		return g.proves(b.Left, false)
	}
	return false
}

// branch resolves which outcome of each condition leads to the branch at at.
func (g guard) branch(x *syntax.If, at syntax.Span) bool {
	if x.Body.Span().Contains(at) {
		return g.proves(x.Cond, true)
	}
	if x.Cond.Span().Contains(at) {
		return false
	}
	if g.proves(x.Cond, false) {
		return true
	}
	for _, e := range x.ElseIfs {
		if e.Span().Contains(at) {
			return e.Body.Span().Contains(at) && g.proves(e.Cond, true)
		}
		if g.proves(e.Cond, false) {
			return true
		}
	}
	return false
}

// proves reports whether cond evaluating to truth excludes the failure value.
func (g guard) proves(cond syntax.Expr, truth bool) bool {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Variable, *syntax.PropertyFetch:
		return truth && g.same(x)
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && g.proves(x.Expr, !truth)
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truth {
				return g.proves(x.Left, true) || g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) && g.proves(x.Right, false)
		case syntax.TBooleanOr, syntax.TOr:
			if truth {
				return g.proves(x.Left, true) && g.proves(x.Right, true)
			}
			return g.proves(x.Left, false) || g.proves(x.Right, false)
		case syntax.TIsIdentical, syntax.TIsEqual:
			return !truth && g.comparesFalse(x)
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			return truth && g.comparesFalse(x)
		}
	case *syntax.Isset:
		// isset($parts) accepts false; only a key probe implies an array.
		for _, v := range x.Vars {
			if _, keyed := syntax.UnwrapParens(v).(*syntax.ArrayDimFetch); keyed && truth && g.same(root(v)) {
				return true
			}
		}
	case *syntax.Empty:
		return !truth && g.same(root(x.Expr))
	case *syntax.FuncCall:
		if !truth {
			return false
		}
		if semanticquery.NativeBuiltin(g.ctx, x, "is_array") {
			return g.same(semanticquery.CallArgument(x.Args, 0, "value"))
		}
		if semanticquery.NativeBuiltin(g.ctx, x, "array_key_exists") {
			return g.same(semanticquery.CallArgument(x.Args, 1, "array"))
		}
	}
	return false
}

func (g guard) comparesFalse(b *syntax.Binary) bool {
	for _, pair := range [][2]syntax.Expr{{b.Left, b.Right}, {b.Right, b.Left}} {
		c, ok := syntax.UnwrapParens(pair[1]).(*syntax.ConstFetch)
		if ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") && g.same(pair[0]) {
			return true
		}
	}
	return false
}

// same ties a tested expression to the value that reaches the read. A test
// before the producing call, or of a reassigned variable, proves nothing.
func (g guard) same(e syntax.Expr) bool {
	if e == nil || e.Span().Start < g.after || !astquery.Equivalent(g.ctx.File, e, g.target) {
		return false
	}
	if _, property := syntax.UnwrapParens(e).(*syntax.PropertyFetch); property {
		// The flow does not follow properties: nothing between the two
		// reads may change this one.
		return semanticquery.NativePropertyUnchanged(g.ctx, e, g.target)
	}
	tested, read := g.ctx.Flow().Value(e), g.ctx.Flow().Value(g.target)
	return tested.Expr == read.Expr && semanticquery.NativeUnreplaced(g.ctx, e, g.target)
}

// root strips array accesses down to the indexed expression.
func root(e syntax.Expr) syntax.Expr {
	for {
		e = syntax.UnwrapParens(e)
		fetch, ok := e.(*syntax.ArrayDimFetch)
		if !ok {
			return e
		}
		e = fetch.Var
	}
}

// Literal URL diagnostics require a supported failure proof. Unknown literal
// shapes cannot justify a missing runtime guard. Dynamic inputs may still fail.
func literalURLFails(raw string) bool {
	if lower := strings.ToLower(raw); lower == "http://" || lower == "https://" {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		return err != nil || n > 65535
	}
	return false
}
