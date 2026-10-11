// Package nonexhaustiveenummatch implements the NonExhaustiveEnumMatch inspection.
package nonexhaustiveenummatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonExhaustiveEnumMatch" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMatch} }

const message = "Handle every reachable enum case."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP81 {
		return
	}
	match := n.(*syntax.Match)
	for _, arm := range match.Arms {
		if arm.Conds == nil {
			return
		}
	}
	classes := ctx.TypeOf(match.Cond).Classes()
	if len(classes) != 1 {
		return
	}
	class := ctx.Index().Class(strings.TrimPrefix(classes[0], `\`), ctx.PHP)
	if class == nil || class.Kind != syntax.KindEnum {
		return
	}
	// missing starts as every case and loses the ones an arm handles.
	missing := map[string]bool{}
	for name, constant := range class.Consts {
		if constant.Case {
			missing[name] = true
		}
	}
	for _, arm := range match.Arms {
		for _, cond := range arm.Conds {
			name := caseName(ctx, class, cond)
			if name == "" {
				return
			}
			delete(missing, name)
		}
	}
	if len(missing) == 0 {
		return
	}
	if name := caseName(ctx, class, semanticquery.NativeValue(ctx, match.Cond)); name != "" {
		keepOnly(missing, name)
	}
	narrow(ctx, match, class, missing)
	if len(missing) != 0 {
		ctx.ReportNode(match, message)
	}
}

func (rule) Semantic() {}

// caseName returns the enum case an `Enum::NAME` expression denotes, following
// constants of the enum that alias one of its cases; "" when e is anything
// else or the constant cannot be traced to a case.
func caseName(ctx *analysis.Context, class *index.Class, e syntax.Expr) string {
	fetch, ok := e.(*syntax.ClassConstFetch)
	if !ok {
		return ""
	}
	id, ok := fetch.Name.(*syntax.Identifier)
	if !ok || semanticquery.StaticCallClass(ctx, fetch.Class) != class.FQN {
		return ""
	}
	name := id.Value
	// Bounded: a constant chain is short, and a cyclic one never ends.
	for range 8 {
		constant := class.Consts[name]
		if constant == nil {
			return ""
		}
		if constant.Case {
			return name
		}
		owner, target, found := strings.Cut(constant.Value, "::")
		owner = strings.TrimPrefix(strings.TrimSpace(owner), `\`)
		short := class.FQN[strings.LastIndexByte(class.FQN, '\\')+1:]
		if !found || !(strings.EqualFold(owner, "self") || strings.EqualFold(owner, short) || strings.EqualFold(owner, class.FQN)) {
			return ""
		}
		name = strings.TrimSpace(target)
	}
	return ""
}

func keepOnly(cases map[string]bool, name string) {
	for other := range cases {
		if other != name {
			delete(cases, other)
		}
	}
}

// narrow removes from cases the ones strict comparisons of the subject
// against enum cases exclude on every path reaching match. Only a subject
// variable that keeps its value for the whole function qualifies: the
// comparisons then hold regardless of what runs between them and the match.
func narrow(ctx *analysis.Context, match *syntax.Match, class *index.Class, cases map[string]bool) {
	subject, ok := syntax.UnwrapParens(match.Cond).(*syntax.Variable)
	if !ok || subject.NameExpr != nil {
		return
	}
	scope := syntax.EnclosingVariableScope(match)
	if !syntax.IsFuncLike(scope) || !stable(ctx, scope, subject.Name) {
		return
	}
	g := guards{ctx: ctx, class: class, name: subject.Name, cases: cases}
	for child, parent := syntax.Node(match), match.Parent(); parent != scope; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.If:
			if child == syntax.Node(p.Body) {
				g.refine(p.Cond, true)
			}
		case *syntax.ElseIf:
			g.failedBefore(p.Parent().(*syntax.If), p)
			if child == syntax.Node(p.Body) {
				g.refine(p.Cond, true)
			}
		case *syntax.Else:
			g.failedBefore(p.Parent().(*syntax.If), nil)
		case *syntax.Ternary:
			if child == syntax.Node(p.Then) {
				g.refine(p.Cond, true)
			}
			if child == syntax.Node(p.Else) {
				g.refine(p.Cond, false)
			}
		case *syntax.Binary:
			if child == syntax.Node(p.Right) {
				switch p.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					g.refine(p.Left, true)
				case syntax.TBooleanOr, syntax.TOr:
					g.refine(p.Left, false)
				}
			}
		}
		// Earlier statements of the same list ran first: a guard whose every
		// branch leaves the list was not taken.
		list, _ := syntax.StmtListOf(parent)
		for i := range max(syntax.StmtIndex(list, child), 0) {
			if exit, ok := list[i].(*syntax.If); ok && exit.Else == nil && leaves(exit) {
				g.failedBefore(exit, nil)
			}
		}
	}
}

// leaves reports whether every branch of an else-less if chain exits abruptly.
func leaves(chain *syntax.If) bool {
	for _, branch := range chain.ElseIfs {
		if !syntax.Terminates(branch.Body) {
			return false
		}
	}
	return syntax.Terminates(chain.Body)
}

// stable reports whether the local variable name provably holds one value
// throughout scope: it is never rebound, never bound to a reference, and the
// scope has no construct able to write a variable chosen at run time or to
// jump past a guard.
func stable(ctx *analysis.Context, scope syntax.Node, name string) bool {
	for _, param := range syntax.FuncLikeParams(scope) {
		if param.ByRef && param.Var != nil && param.Var.Name == name {
			return false
		}
	}
	for _, access := range flowquery.VarAccesses(ctx.File, scope, name) {
		if access.Write {
			return false
		}
		switch p := access.Var.Parent().(type) {
		case *syntax.ClosureUse:
			if p.ByRef {
				return false
			}
		case *syntax.Assign:
			if p.ByRef {
				return false
			}
		case *syntax.ArrayItem:
			if p.ByRef {
				return false
			}
		case *syntax.Yield:
			return false // a by-reference generator hands out the variable itself
		case *syntax.Arg:
			if !passedByValue(ctx, p) {
				return false
			}
		}
	}
	if closure, ok := scope.(*syntax.Closure); ok {
		for _, use := range closure.Uses {
			if use.ByRef && use.Var != nil && use.Var.Name == name {
				return false
			}
		}
	}
	dynamic := false
	syntax.Inspect(syntax.VariableScopeBody(scope), func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Goto, *syntax.Include, *syntax.Eval:
			dynamic = true
		case *syntax.Variable:
			dynamic = dynamic || x.NameExpr != nil
		case *syntax.FuncCall:
			dynamic = dynamic || ctx.IsGlobalFunctionCall(x, "extract")
		}
		return !dynamic
	})
	return !dynamic
}

// passedByValue reports whether arg is proven to bind to a by-value
// parameter: only a resolved function declaration establishes that.
func passedByValue(ctx *analysis.Context, arg *syntax.Arg) bool {
	list := arg.Parent().(*syntax.ArgList)
	call, ok := list.Parent().(*syntax.FuncCall)
	if !ok || arg.ByRef || arg.Unpack {
		return false
	}
	function := ctx.Types().ResolveFunction(call)
	if function == nil {
		return false
	}
	position := 0
	for _, other := range list.Args {
		if other == syntax.Expr(arg) {
			break
		}
		position++
	}
	if arg.Name != nil {
		position = len(function.Params)
		for i, param := range function.Params {
			if strings.TrimPrefix(param.Name, "$") == arg.Name.Value {
				position = i
			}
		}
		if position == len(function.Params) {
			return false
		}
	}
	if last := len(function.Params) - 1; position > last {
		if last < 0 || !function.Params[last].Variadic {
			return true // surplus arguments are discarded
		}
		position = last
	}
	return !function.Params[position].ByRef
}

// guards applies comparison outcomes to the set of still possible cases.
type guards struct {
	ctx   *analysis.Context
	class *index.Class
	name  string
	cases map[string]bool
}

// failedBefore records that the conditions of chain evaluated before branch
// (all of them when branch is nil) were false.
func (g guards) failedBefore(chain *syntax.If, branch *syntax.ElseIf) {
	g.refine(chain.Cond, false)
	for _, earlier := range chain.ElseIfs {
		if earlier == branch {
			return
		}
		g.refine(earlier.Cond, false)
	}
}

// refine narrows the cases with the knowledge that cond evaluated to truth.
func (g guards) refine(cond syntax.Expr, truth bool) {
	switch x := syntax.UnwrapParens(cond).(type) {
	case *syntax.Unary:
		if x.Op.Kind == syntax.TExclaim {
			g.refine(x.Expr, !truth)
		}
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			// A true conjunction and a false disjunction fix both operands.
			if truth == (x.Op.Kind == syntax.TBooleanAnd || x.Op.Kind == syntax.TAnd) {
				g.refine(x.Left, truth)
				g.refine(x.Right, truth)
			}
		case syntax.TIsIdentical, syntax.TIsNotIdentical:
			name := g.compared(x.Left, x.Right)
			if name == "" {
				name = g.compared(x.Right, x.Left)
			}
			switch {
			case name == "":
			case truth == (x.Op.Kind == syntax.TIsIdentical):
				keepOnly(g.cases, name)
			default:
				delete(g.cases, name)
			}
		}
	}
}

// compared returns the case subject is compared with when one operand is the
// subject variable and the other a case of its enum.
func (g guards) compared(subject, other syntax.Expr) string {
	variable, ok := syntax.UnwrapParens(subject).(*syntax.Variable)
	if !ok || variable.NameExpr != nil || variable.Name != g.name {
		return ""
	}
	return caseName(g.ctx, g.class, syntax.UnwrapParens(other))
}
