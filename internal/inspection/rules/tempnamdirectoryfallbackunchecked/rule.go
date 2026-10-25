// Package tempnamdirectoryfallbackunchecked implements the native TempnamDirectoryFallbackUnchecked inspection.
package tempnamdirectoryfallbackunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Verify the temporary file remains in the required directory."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TempnamDirectoryFallbackUnchecked" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "file_put_contents" && name != "fopen" {
		return
	}
	path := semanticquery.CallArgument(c.Args, 0, "filename")
	open, ok := semanticquery.NativeValue(ctx, path).(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, open) != "tempnam" {
		return
	}
	// Asking for the system temporary directory leaves nothing to fall back to.
	directory := semanticquery.CallArgument(open.Args, 0, "directory")
	if value := semanticquery.NativeValue(ctx, directory); value != nil {
		directory = value
	}
	if system, ok := syntax.UnwrapParens(directory).(*syntax.FuncCall); ok && semanticquery.NativeBuiltinName(ctx, system) == "sys_get_temp_dir" {
		return
	}
	// E6: test fixtures are throwaway files, not sensitive content.
	if semanticquery.InTestContext(ctx, c) {
		return
	}
	if !directoryVerified(ctx, c, path, open) {
		ctx.ReportNode(c, message)
	}
}

// canonicalParent returns the variable read by realpath(dirname($name)),
// nil for anything else.
func canonicalParent(ctx *analysis.Context, e syntax.Expr, name string) *syntax.Variable {
	dir, ok := syntax.UnwrapParens(canonicalPath(ctx, e)).(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, dir) != "dirname" {
		return nil
	}
	v, ok := syntax.UnwrapParens(semanticquery.CallArgument(dir.Args, 0, "path")).(*syntax.Variable)
	if !ok || v.Name != name {
		return nil
	}
	return v
}

// canonicalPath returns the argument of a builtin realpath() call.
func canonicalPath(ctx *analysis.Context, e syntax.Expr) syntax.Expr {
	call, ok := syntax.UnwrapParens(e).(*syntax.FuncCall)
	if !ok || semanticquery.NativeBuiltinName(ctx, call) != "realpath" {
		return nil
	}
	return semanticquery.CallArgument(call.Args, 0, "path")
}

// directoryVerified proves that the canonical parent of the temporary file
// held by path was compared with a canonical directory after source created it
// and that the comparison is known to have matched where at runs: an equality
// that the conditions guarding at prove true, or an inequality they prove
// false, while path still holds the same file.
func directoryVerified(ctx *analysis.Context, at, path syntax.Expr, source syntax.Node) bool {
	variable, ok := syntax.UnwrapParens(path).(*syntax.Variable)
	if !ok || semanticquery.NativeLabelled(ctx, at) {
		return false
	}
	env := ctx.Types().Native()
	return semanticquery.NativeEarlierCallProves(ctx, at, func(n syntax.Expr) (types.Type, func(types.Type) bool) {
		cmp, ok := n.(*syntax.Binary)
		if !ok || cmp.Span().Start < source.Span().End {
			return types.Unknown, nil
		}
		// An equality must have held, an inequality failed. Canonical paths
		// are absolute strings or false: the loose operators decide alike.
		want := ""
		switch cmp.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsEqual:
			want = "true"
		case syntax.TIsNotIdentical, syntax.TIsNotEqual:
			want = "false"
		default:
			return types.Unknown, nil
		}
		parent := canonicalParent(ctx, cmp.Left, variable.Name)
		if parent == nil || canonicalPath(ctx, cmp.Right) == nil {
			if parent = canonicalParent(ctx, cmp.Right, variable.Name); parent == nil || canonicalPath(ctx, cmp.Left) == nil {
				return types.Unknown, nil
			}
		}
		if !env.ReachesUnchanged(parent, at) {
			return types.Unknown, nil // another file since
		}
		return types.Bool, func(t types.Type) bool { return t.OnlyOf(want) }
	})
}
