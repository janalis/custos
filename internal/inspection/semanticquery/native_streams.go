package semanticquery

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// SameOperand reports whether a and b are known to hold the same value:
// the same expression, the same literal string, or the same object not
// invalidated in between.
func SameOperand(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	av, bv := ctx.Flow().Value(a), ctx.Flow().Value(b)
	if !av.Complete || !bv.Complete {
		return false
	}
	if av.Expr != nil && av.Expr == bv.Expr {
		return true
	}
	if left, ok := NativeString(ctx, a); ok {
		right, known := NativeString(ctx, b)
		return known && left == right
	}
	return av.Identity != 0 && av.Identity == bv.Identity && av.Invalidated == bv.Invalidated && av.Invalidated != ^uint32(0)
}

// FileLockOperation folds a flock operation made of known integers and
// constants joined by bitwise or. Anything else stays unknown.
func FileLockOperation(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if e == nil {
		return 0, false
	}
	if n, ok := NativeInt(ctx, e); ok {
		return n, true
	}
	if binary, ok := NativeValue(ctx, e).(*syntax.Binary); ok && binary.Op.Kind == syntax.TBar {
		left, lk := FileLockOperation(ctx, binary.Left)
		right, rk := FileLockOperation(ctx, binary.Right)
		return left | right, lk && rk
	}
	return 0, false
}

// PathArgumentCall identifies a builtin consuming its first argument as a path.
func PathArgumentCall(ctx *analysis.Context, c *syntax.FuncCall) bool {
	switch NativeBuiltinName(ctx, c) {
	case "file_get_contents", "file_put_contents", "fopen", "file", "is_file", "is_dir", "file_exists", "unlink", "mkdir", "rmdir", "scandir", "opendir":
		return true
	}
	return false
}
