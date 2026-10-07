package controlflow

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// strlenInEmptyStringCheckContext reports strlen()/mb_strlen() used only to
// test whether a string is empty.
type strlenInEmptyStringCheckContext struct{}

func init() { register(strlenInEmptyStringCheckContext{}) }

func (strlenInEmptyStringCheckContext) ID() string { return "StrlenInEmptyStringCheckContext" }

func (strlenInEmptyStringCheckContext) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strlenInEmptyStringCheckContext) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if name := ctx.GlobalFunctionName(call); name != "strlen" && name != "mb_strlen" { // D1
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) == 0 || args[0] == nil {
		return
	}
	arg := args[0]
	// D2: reported everywhere, top-level code included (custos diverges).
	target, empty, ok := strlenTarget(ctx, call)
	if !ok {
		return
	}
	// F1: parenthesise A when it would not bind as a single operand of the
	// cast or of the comparison (custos diverges).
	v := ctx.Text(arg)
	if t := ctx.TypeOf(arg); !(len(t.Atoms()) == 1 && t.Atoms()[0] == "string") {
		if util.NeedsParensAsUnaryOperand(arg) {
			v = "(" + v + ")"
		}
		v = "(string)" + v
	} else if util.NeedsParensAsEqualityOperand(arg) {
		v = "(" + v + ")"
	}
	op := "!=="
	if empty {
		op = "==="
	}
	repl := v + " " + op + " ''"
	if ctx.ComparisonStyle == analysis.StyleYoda {
		repl = "'' " + op + " " + v
	}
	ctx.ReportNode(target, "Compare with an empty string instead: '"+repl+"'.", analysis.Fix{
		Title: "Compare with an empty string",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: target.Span(), NewText: repl}}
		},
	})
}

// strlenTarget applies D3–D7 and returns the node to replace and whether
// the replacement tests for emptiness.
func strlenTarget(ctx *analysis.Context, call *syntax.FuncCall) (syntax.Node, bool, bool) {
	if b, ok := call.Parent().(*syntax.Binary); ok {
		left := b.Left == syntax.Expr(call)
		other := b.Right
		if !left {
			other = b.Left
		}
		if util.IsNumberLiteral(other) {
			num := ctx.Text(other)
			switch {
			case b.Op.Kind == syntax.TGreater && left && num == "0": // D3
				return b, false, true
			case (b.Op.Kind == syntax.TLess || b.Op.Kind == syntax.TIsGreaterOrEqual) && left && num == "1": // D4
				return b, b.Op.Kind == syntax.TLess, true
			case num == "0": // D5
				switch b.Op.Kind {
				case syntax.TIsEqual, syntax.TIsIdentical:
					return b, true, true
				case syntax.TIsNotEqual, syntax.TIsNotIdentical:
					return b, false, true
				}
			}
		}
	}
	parent, child := util.ParentSkipParens(call) // D6
	logical := false
	switch p := parent.(type) {
	case *syntax.If:
		logical = p.Cond == child
	case *syntax.ElseIf:
		logical = p.Cond == child
	case *syntax.While:
		logical = p.Cond == child
	case *syntax.DoWhile:
		logical = p.Cond == child
	case *syntax.Unary:
		logical = p.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			logical = true
		}
	case *syntax.Ternary:
		logical = p.Then != nil && p.Cond == child
	}
	if !logical {
		return nil, false, false
	}
	if u, ok := call.Parent().(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim { // D7
		return u, true, true
	}
	return call, false, true
}
