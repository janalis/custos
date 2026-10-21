// Package seeksuccesscheckedastruthy implements the native SeekSuccessCheckedAsTruthy inspection.
package seeksuccesscheckedastruthy

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare the seek result with zero."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SeekSuccessCheckedAsTruthy" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "fseek") && conditionUse(ctx, c, true) != nil {
		ctx.ReportNode(c, message)
	}
}

// conditionUse returns the truthiness expression containing a call.
// Integer sentinel comparisons are deliberately excluded.
// A first-class callable creates a closure and has no result to test; only a
// plain assignment of the result keeps its value as the tested operand.
func conditionUse(ctx *analysis.Context, call *syntax.FuncCall, booleanComparisons bool) syntax.Expr {
	if call.Args != nil && len(call.Args.Args) == 1 {
		if _, ok := call.Args.Args[0].(*syntax.VariadicPlaceholder); ok {
			return nil
		}
	}
	var e syntax.Expr = call
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
			e = p
		case *syntax.Assign:
			if p.Op.Kind != syntax.TEqual || p.Value != e {
				return nil
			}
			e = p
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil
			}
			e = p
		case *syntax.Binary:
			if !booleanComparisons {
				return nil
			}
			if p.Op.Kind != syntax.TIsEqual && p.Op.Kind != syntax.TIsNotEqual && p.Op.Kind != syntax.TIsIdentical && p.Op.Kind != syntax.TIsNotIdentical {
				return nil
			}
			other := p.Right
			if other == e {
				other = p.Left
			}
			c, ok := other.(*syntax.ConstFetch)
			if !ok || c.Name == nil || (!strings.EqualFold(c.Name.Value, "true") && !strings.EqualFold(c.Name.Value, "false")) {
				return nil
			}
			e = p
		case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile:
			return e
		case *syntax.For:
			// Only the last condition expression decides the loop.
			if n := len(p.Cond); n == 0 || p.Cond[n-1] != e {
				return nil
			}
			return e
		default:
			return nil
		}
	}
}
