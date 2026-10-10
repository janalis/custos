// Package worldwritablepermission implements the native WorldWritablePermission inspection.
package worldwritablepermission

import (
	"path/filepath"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Remove world-write permission from sensitive files."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WorldWritablePermission" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "chmod") && !semanticquery.NativeBuiltin(ctx, c, "mkdir") {
		return
	}
	pathName := "filename"
	if semanticquery.NativeBuiltin(ctx, c, "mkdir") {
		pathName = "directory"
	}
	path, pk := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, pathName))
	mode, mk := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "permissions"))
	if !pk || !mk || mode&0o002 == 0 {
		return
	}
	for _, pattern := range ctx.List("sensitivePaths") {
		if sensitiveMatch(pattern, path) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

func sensitiveMatch(pattern, path string) bool {
	p := strings.Split(strings.Trim(strings.ReplaceAll(pattern, `\`, "/"), "/"), "/")
	values := strings.Split(strings.Trim(strings.ReplaceAll(path, `\`, "/"), "/"), "/")
	previous := make([]bool, len(values)+1)
	previous[0] = true
	for _, component := range p {
		next := make([]bool, len(values)+1)
		if component == "**" {
			next[0] = previous[0]
			for i := 1; i <= len(values); i++ {
				next[i] = previous[i] || next[i-1]
			}
		} else {
			for i := 1; i <= len(values); i++ {
				match, err := filepath.Match(component, values[i-1])
				next[i] = previous[i-1] && err == nil && match
			}
		}
		previous = next
	}
	return previous[len(values)]
}
