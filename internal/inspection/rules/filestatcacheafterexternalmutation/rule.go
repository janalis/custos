// Package filestatcacheafterexternalmutation implements the native FileStatCacheAfterExternalMutation inspection.
package filestatcacheafterexternalmutation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clear cached file metadata after external file changes."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileStatCacheAfterExternalMutation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "filesize" && name != "filemtime" {
		return
	}
	path, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "filename"))
	if !known {
		return
	}
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) < 2 {
		return
	}
	mutate, ok := statementCall(prior[len(prior)-1])
	if !ok || !semanticquery.NativeBuiltin(ctx, mutate, "exec") && !semanticquery.NativeBuiltin(ctx, mutate, "system") {
		return
	}
	command, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(mutate.Args, 0, "command"))
	if !known || !mutates(command, path) {
		return
	}
	before, ok := statementCall(prior[len(prior)-2])
	if !ok || !semanticquery.NativeBuiltin(ctx, before, name) {
		return
	}
	beforePath, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(before.Args, 0, "filename"))
	if known && beforePath == path {
		ctx.ReportNode(c, message)
	}
}

func statementCall(st syntax.Stmt) (*syntax.FuncCall, bool) {
	e, ok := st.(*syntax.ExprStmt)
	if !ok {
		return nil, false
	}
	value := e.Expr
	if a, ok := value.(*syntax.Assign); ok && !a.ByRef && a.Op.Kind == syntax.TEqual {
		value = a.Value
	}
	call, ok := value.(*syntax.FuncCall)
	return call, ok
}

func mutates(command, path string) bool {
	for _, ch := range command {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || strings.ContainsRune("/._- ", ch)) {
			return false
		}
	}
	fields := strings.Fields(command)
	if len(fields) == 2 && (fields[0] == "touch" || fields[0] == "rm") {
		return fields[1] == path && !strings.HasPrefix(path, "-")
	}
	return len(fields) == 4 && fields[0] == "truncate" && fields[1] == "-s" && fields[2] == "0" && fields[3] == path && !strings.HasPrefix(path, "-")
}
