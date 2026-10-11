// Package exitstatusreadwithoutnormalexitcheck implements the native ExitStatusReadWithoutNormalExitCheck inspection.
package exitstatusreadwithoutnormalexitcheck

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check normal termination before decoding the exit code."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExitStatusReadWithoutNormalExitCheck" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pcntl_wexitstatus") {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "status")
	wait, flags := semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_waitpid"}, 1, "status"), 2
	if wait == nil {
		wait, flags = semanticquery.ExpansionCOutputCall(ctx, c, input, []string{"pcntl_wait"}, 0, "status"), 1
	}
	if wait == nil {
		return
	}
	p := proof{ctx: ctx, wait: wait, status: input, blocking: semanticquery.CallArgument(wait.Args, flags, "flags") == nil}
	for child, parent := syntax.Node(c), c.Parent(); parent != nil; child, parent = parent, parent.Parent() {
		switch x := parent.(type) {
		case *syntax.If:
			if child == x.Body {
				p.assume(x.Cond, true)
			} else if child != x.Cond {
				// An elseif or else branch runs only after every earlier condition failed.
				p.assume(x.Cond, false)
				for _, earlier := range x.ElseIfs {
					if earlier.Span().Start < child.Span().Start {
						p.assume(earlier.Cond, false)
					}
				}
			}
		case *syntax.ElseIf:
			if child == x.Body {
				p.assume(x.Cond, true)
			}
		case *syntax.Ternary:
			switch child {
			case x.Then:
				p.assume(x.Cond, true)
			case x.Else:
				p.assume(x.Cond, false)
			}
		case *syntax.Binary:
			if child == x.Right {
				switch x.Op.Kind {
				case syntax.TBooleanAnd, syntax.TAnd:
					p.assume(x.Left, true)
				case syntax.TBooleanOr, syntax.TOr:
					p.assume(x.Left, false)
				}
			}
		}
		if _, ok := parent.(syntax.Stmt); !ok {
			continue
		}
		// Falling past a guard whose body always leaves proves its condition failed.
		for _, prior := range flowquery.NativePriorStatements(ctx.File, parent) {
			if guard, ok := prior.(*syntax.If); ok && syntax.Terminates(guard.Body) {
				p.assume(guard.Cond, false)
			}
		}
	}
	if p.success && p.exited {
		return
	}

	ctx.ReportNode(c, message)
}

// proof accumulates what the conditions dominating the decoded status establish.
type proof struct {
	ctx             *analysis.Context
	wait            *syntax.FuncCall
	status          syntax.Expr
	blocking        bool // flags omitted: the result is a child pid or -1
	success, exited bool
}

var (
	mirrored = map[syntax.TokenKind]syntax.TokenKind{
		syntax.TLess: syntax.TGreater, syntax.TGreater: syntax.TLess,
		syntax.TIsSmallerOrEqual: syntax.TIsGreaterOrEqual, syntax.TIsGreaterOrEqual: syntax.TIsSmallerOrEqual,
	}
	// Only operators whose negation can prove a successful wait are listed.
	negated = map[syntax.TokenKind]syntax.TokenKind{
		syntax.TLess: syntax.TIsGreaterOrEqual, syntax.TIsSmallerOrEqual: syntax.TGreater,
		syntax.TIsIdentical: syntax.TIsNotIdentical, syntax.TIsEqual: syntax.TIsNotEqual,
	}
)

// assume records the facts that follow from e evaluating to truth.
func (p *proof) assume(e syntax.Expr, truth bool) {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Unary:
		if x.Op.Kind == syntax.TExclaim {
			p.assume(x.Expr, !truth)
		}
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd:
			if truth {
				p.assume(x.Left, true)
				p.assume(x.Right, true)
			}
		case syntax.TBooleanOr, syntax.TOr:
			if !truth {
				p.assume(x.Left, false)
				p.assume(x.Right, false)
			}
		default:
			p.compare(x, truth)
		}
	case *syntax.FuncCall:
		// A check placed before the wait describes an older status value.
		if truth && x.Span().Start > p.wait.Span().End && semanticquery.NativeBuiltin(p.ctx, x, "pcntl_wifexited") && semanticquery.ExpansionCSame(p.ctx, p.status, semanticquery.CallArgument(x.Args, 0, "status")) {
			p.exited = true
		}
	}
}

// compare recognises a comparison of the wait result with an integer, in
// either operand order, that rules out a failed or empty wait.
func (p *proof) compare(b *syntax.Binary, truth bool) {
	op, subject := b.Op.Kind, b.Left
	bound, known := semanticquery.NativeInt(p.ctx, b.Right)
	if !known {
		if bound, known = semanticquery.NativeInt(p.ctx, b.Left); !known {
			return
		}
		subject = b.Right
		if flipped, ok := mirrored[op]; ok {
			op = flipped
		}
	}
	if semanticquery.NativeValue(p.ctx, subject) != p.wait {
		return
	}
	if !truth {
		op = negated[op]
	}
	switch op {
	case syntax.TGreater:
		p.success = p.success || bound >= 0
	case syntax.TIsGreaterOrEqual:
		p.success = p.success || bound >= 1
	case syntax.TIsNotIdentical, syntax.TIsNotEqual:
		// Without flags the call blocks, so the only non-child result is -1.
		p.success = p.success || (bound == -1 && p.blocking)
	}
}
