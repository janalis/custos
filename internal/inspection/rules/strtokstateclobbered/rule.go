// Package strtokstateclobbered implements the StrtokStateClobbered inspection.
package strtokstateclobbered

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StrtokStateClobbered" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Keep strtok sequences separate."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := strtokCall(ctx, n, 1)
	if call == nil {
		return
	}
	// Walk the straight-line path backwards, leaving a loop or `if` body
	// through its header, for the newest start and then the one before it.
	var second *syntax.FuncCall
	for at := syntax.Node(call); at != nil; at = entered(ctx, at, second == nil) {
		prior := flowquery.NativePriorStatements(ctx.File, at)
		for i := len(prior) - 1; i >= 0; i-- {
			start := startCall(ctx, prior[i])
			switch {
			case start == nil:
				if !transparent(ctx, prior[i], second == nil) {
					return
				}
			case second == nil:
				second = start
			default:
				if clobbered(ctx.File, start, second, call) {
					ctx.ReportNode(call, message)
				}
				return
			}
		}
	}
}
func (rule) Semantic() {}

// strtokCall returns n as a call to the builtin strtok with exactly args arguments.
func strtokCall(ctx *analysis.Context, n syntax.Node, args int) *syntax.FuncCall {
	call, _ := semanticquery.GlobalCall(ctx, n, "strtok")
	if call == nil || call.Args == nil || len(call.Args.Args) != args {
		return nil
	}
	return call
}

// startCall returns the two-argument strtok call a statement consists of,
// directly or as the assigned value.
func startCall(ctx *analysis.Context, s syntax.Stmt) *syntax.FuncCall {
	statement, ok := s.(*syntax.ExprStmt)
	if !ok {
		return nil
	}
	expression := statement.Expr
	if assign, ok := expression.(*syntax.Assign); ok {
		expression = assign.Value
	}
	return strtokCall(ctx, expression, 2)
}

// clobbered reports whether the continuation resumes the first sequence after
// the second one replaced the tokenizer state.
func clobbered(f *syntax.File, first, second, continuation *syntax.FuncCall) bool {
	if astquery.Equivalent(f, semanticquery.CallArgument(first.Args, 0, "string"), semanticquery.CallArgument(second.Args, 0, "string")) {
		return false
	}
	// A continuation using only the newer start's delimiter continues that
	// sequence: the older call just took a leading token.
	old, okOld := astquery.QuotedStringValue(semanticquery.CallArgument(first.Args, 1, "token"))
	recent, okRecent := astquery.QuotedStringValue(semanticquery.CallArgument(second.Args, 1, "token"))
	used, okUsed := astquery.QuotedStringValue(semanticquery.CallArgument(continuation.Args, 0, "token"))
	return !okOld || !okRecent || !okUsed || used != recent || used == old
}

// transparent reports whether running n cannot start a tokenizer sequence:
// it holds no call-like or suspending expression. One-argument strtok calls
// only advance the current sequence and are accepted when continuations is set.
func transparent(ctx *analysis.Context, n syntax.Node, continuations bool) bool {
	ok := true
	syntax.Inspect(n, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.FuncCall:
			if !continuations || strtokCall(ctx, node, 1) == nil {
				ok = false
			}
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Include, *syntax.Eval,
			*syntax.Yield, *syntax.YieldFrom, *syntax.Closure, *syntax.ArrowFunction:
			ok = false
		}
		return ok
	})
	return ok
}

// entered returns the loop or `if` whose body directly holds the statement
// around n, when its header is transparent: the statements before it then
// lead into that body on the first pass. Other constructs end the search.
func entered(ctx *analysis.Context, n syntax.Node, continuations bool) syntax.Node {
	var owner syntax.Node
	for ; n != nil; n = n.Parent() {
		if _, ok := n.(syntax.Stmt); ok {
			owner = bodyOwner(ctx, n, continuations)
			break
		}
	}
	return owner
}

func bodyOwner(ctx *analysis.Context, statement syntax.Node, continuations bool) syntax.Node {
	owner := statement.Parent()
	if block, ok := owner.(*syntax.Block); ok {
		owner = block.Parent()
	}
	var header []syntax.Expr
	switch o := owner.(type) {
	case *syntax.If: // `elseif` and `else` bodies have their own parent nodes
		header = []syntax.Expr{o.Cond}
	case *syntax.While:
		header = []syntax.Expr{o.Cond}
	case *syntax.Foreach:
		header = []syntax.Expr{o.Expr}
	case *syntax.For:
		header = append(append(header, o.Init...), o.Cond...)
	default:
		return nil
	}
	for _, e := range header {
		if !transparent(ctx, e, continuations) {
			return nil
		}
	}
	return owner
}
