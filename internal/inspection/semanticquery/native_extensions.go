package semanticquery

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// AppendedFrames proves a lower bound for successfully appended local frames.
func AppendedFrames(ctx *analysis.Context, c *syntax.MethodCall) int {
	return appendedFrames(ctx, c.Var, c, 0)
}

func appendedFrames(ctx *analysis.Context, obj, at syntax.Expr, depth int) int {
	if depth > 8 {
		return 0
	}
	if !UnaliasedLocal(ctx, obj, at) {
		return 0
	}
	creation := NativeConstruction(ctx, obj, "Imagick")
	if creation == nil {
		return 0
	}
	if _, state := CallArgumentBinding(creation.Args, 0, "files"); state != ArgumentMissing {
		return 0
	}
	count := 0
	for _, prior := range ObjectMethodHistory(ctx, at, obj) {
		switch LiteralMethodName(prior) {
		case "newimage":
			if !NativeCallSucceeded(ctx, prior, at) {
				return 0
			}
			count++
		case "addimage":
			if !NativeCallSucceeded(ctx, prior, at) {
				return 0
			}
			frames := appendedFrames(ctx, CallArgument(prior.Args, 0, "source"), prior, depth+1)
			if frames == 0 {
				return 0
			}
			count += frames
		case "getnumberimages", "getimageblob", "getimagesblob", "setiteratorindex":
		default:
			return 0
		}
	}
	return count
}

// MatchingQueueSend identifies one preceding successful local send selected by
// an exact positive type, with no intervening queue receive or send ambiguity.
func MatchingQueueSend(ctx *analysis.Context, c *syntax.FuncCall) *syntax.FuncCall {
	queue := CallArgument(c.Args, 0, "queue")
	if !UnaliasedLocal(ctx, queue, c) {
		return nil
	}
	wanted, known := NativeContractInt(ctx, CallArgument(c.Args, 1, "desired_message_type"))
	if !known || wanted <= 0 {
		return nil
	}
	flags := CallArgument(c.Args, 6, "flags")
	if flags != nil {
		v, k := NativeContractInt(ctx, flags)
		if !k || v&4 != 0 {
			return nil
		}
	}
	calls := NativeStreamCalls(ctx, c, queue, "msg_send", "msg_receive", "msg_remove_queue")
	if len(calls) != 1 {
		return nil
	}
	send := calls[0]
	if !NativeBuiltin(ctx, send, "msg_send") || !NativeCallSucceeded(ctx, send, c) {
		return nil
	}
	sent, k := NativeContractInt(ctx, CallArgument(send.Args, 1, "message_type"))
	if !k || sent != wanted {
		return nil
	}
	return send
}

// FiberConstruction returns the `new Fiber` that the receiver of c, a call
// of the Fiber method name, holds; nil before PHP 8.1 or when unknown.
func FiberConstruction(ctx *analysis.Context, c *syntax.MethodCall, name string) *syntax.New {
	if ctx.PHP.Below(phpversion.PHP81) || !NativeMethod(ctx, c, "Fiber", name) {
		return nil
	}
	return NativeConstruction(ctx, c.Var, "Fiber")
}

// BCMathOperands returns the numeric operand slots of a resolved BCMath
// call. Scale and exponent parameters are not decimal-value operands.
func BCMathOperands(ctx *analysis.Context, c *syntax.FuncCall) []syntax.Expr {
	switch NativeBuiltinName(ctx, c) {
	case "bcadd", "bcsub", "bcmul", "bcdiv", "bcmod", "bccomp":
		return []syntax.Expr{CallArgument(c.Args, 0, "num1"), CallArgument(c.Args, 1, "num2")}
	case "bcpowmod":
		return []syntax.Expr{CallArgument(c.Args, 0, "num"), CallArgument(c.Args, 2, "modulus")}
	case "bcpow", "bcsqrt":
		return []syntax.Expr{CallArgument(c.Args, 0, "num")}
	}
	return nil
}
