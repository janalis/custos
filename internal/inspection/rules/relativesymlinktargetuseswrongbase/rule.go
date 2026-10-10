// Package relativesymlinktargetuseswrongbase implements the native RelativeSymlinkTargetUsesWrongBase inspection.
package relativesymlinktargetuseswrongbase

import (
	"path"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Resolve the link target against its directory."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RelativeSymlinkTargetUsesWrongBase" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.ExpansionDPathCall(ctx, c) {
		return
	}
	read, ok := syntax.UnwrapParens(semanticquery.CallArgument(c.Args, 0, "filename")).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, read, "readlink") {
		return
	}
	link, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(read.Args, 0, "path"))
	if !k || !path.IsAbs(link) {
		return
	}
	cwd, target := "", ""
	for _, event := range semanticquery.ExpansionDLocalEvents(ctx, c) {
		if event.Span().Start >= read.Span().Start {
			break
		}
		f, ok := event.(*syntax.FuncCall)
		if !ok {
			continue
		}
		name := semanticquery.NativeBuiltinName(ctx, f)
		if name == "chdir" {
			cwd = ""
			if semanticquery.NativeDominates(f, c) {
				cwd, _ = semanticquery.NativeString(ctx, semanticquery.CallArgument(f.Args, 0, "directory"))
			}
		}
		if name == "symlink" {
			p, pk := semanticquery.NativeString(ctx, semanticquery.CallArgument(f.Args, 1, "link"))
			if pk && p == link {
				target = ""
				if semanticquery.NativeDominates(f, c) {
					target, _ = semanticquery.NativeString(ctx, semanticquery.CallArgument(f.Args, 0, "target"))
				}
			}
		}
		if name == "unlink" || name == "rename" || name == "" {
			target = ""
			cwd = ""
		}
	}
	if path.IsAbs(cwd) && target != "" && !path.IsAbs(target) && path.Clean(cwd) != path.Dir(link) {
		ctx.ReportNode(c, message)
	}
}
