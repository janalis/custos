// Package filetruncatedbeforelock implements the native FileTruncatedBeforeLock inspection.
package filetruncatedbeforelock

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Acquire the lock before truncating the file."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FileTruncatedBeforeLock" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "fopen" {
		return
	}
	mode, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "mode"))
	if !known || !strings.HasPrefix(mode, "w") {
		return
	}
	assignment, ok := c.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok {
		return
	}
	for _, later := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		lock, ok := later.Node.(*syntax.FuncCall)
		if !ok || later.Name != "flock" || lock.Span().Start <= c.Span().End {
			continue
		}
		arg, ok := semanticquery.CallArgument(lock.Args, 0, "stream").(*syntax.Variable)
		if !ok || arg.Name != variable.Name || ctx.Flow().Value(arg).Identity != ctx.Flow().Value(c).Identity {
			continue
		}
		// LOCK_UN (3) shares a bit with LOCK_EX (2), so compare the whole
		// operation once the optional LOCK_NB (4) modifier is masked out.
		if operation, known := semanticquery.FileLockOperation(ctx, semanticquery.CallArgument(lock.Args, 1, "operation")); known && operation&^4 == 2 {
			if holdsData(ctx, c, variable.Name) { // E4
				ctx.ReportNode(c, message)
			}
			return
		}
	}
}

// dataCalls read or write the content of a stream.
var dataCalls = map[string]bool{
	"fwrite": true, "fputs": true, "fputcsv": true, "fprintf": true, "vfprintf": true,
	"fread": true, "fgets": true, "fgetc": true, "fgetcsv": true, "fscanf": true,
	"fpassthru": true, "stream_get_contents": true, "stream_copy_to_stream": true,
}

// holdsData reports whether the scope of the fopen call reads or writes
// through the handle held in name (E4): a file that is only locked and
// closed is a lock file, with no content its truncation could destroy.
func holdsData(ctx *analysis.Context, open *syntax.FuncCall, name string) bool {
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(open)) {
		c, ok := call.Node.(*syntax.FuncCall)
		if !ok || !dataCalls[call.Name] {
			continue
		}
		for _, a := range c.Args.Args {
			if arg, ok := a.(*syntax.Arg); ok {
				if v, ok := arg.Value.(*syntax.Variable); ok && v.Name == name {
					return true
				}
			}
		}
	}
	return false
}
