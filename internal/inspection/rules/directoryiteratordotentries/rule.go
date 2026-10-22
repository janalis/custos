// Package directoryiteratordotentries implements the native DirectoryIteratorDotEntries inspection.
package directoryiteratordotentries

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Skip dot entries before modifying directory entries."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DirectoryIteratorDotEntries" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, c)
	if name != "unlink" && name != "rmdir" {
		return
	}
	path, ok := semanticquery.CallArgument(c.Args, 0, "filename").(*syntax.MethodCall)
	if !ok {
		return
	}
	variable, ok := path.Var.(*syntax.Variable)
	if !ok {
		return
	}
	for p := c.Parent(); p != nil; p = p.Parent() {
		loop, ok := p.(*syntax.Foreach)
		if !ok {
			continue
		}
		iterator, ok := loop.Expr.(*syntax.New)
		if !ok {
			continue
		}
		class, ok := iterator.Class.(*syntax.Name)
		if !ok || ctx.Names().Class(class.Value, class.Span().Start) != "DirectoryIterator" {
			continue
		}
		value, ok := loop.Value.(*syntax.Variable)
		if !ok || value.Name != variable.Name {
			continue
		}
		if iteratorReassigned(loop, c, variable.Name) {
			return
		}
		if dotEntriesExcludedAt(ctx, loop, c, variable.Name) != dotEntryNames {
			ctx.ReportNode(c, message)
		}
		return
	}
}

// dotEntriesExcludedAt returns which dot entries the loop variable can
// no longer name when c runs: what the conditions guarding c prove of the
// entry's earlier isDot()/isDir()/isFile()/isLink() results and of its name
// (getFilename(), argument-less getBasename()), joined over those calls.
func dotEntriesExcludedAt(ctx *analysis.Context, loop *syntax.Foreach, c *syntax.FuncCall, variable string) uint8 {
	if semanticquery.NativeLabelled(ctx, c) {
		return 0 // a jump may land between a test and the call
	}
	env := ctx.Types().Native()
	var mask uint8
	syntax.Inspect(loop.Body, func(n syntax.Node) bool {
		if n.Span().Start >= c.Span().Start || syntax.IsVariableScope(n) {
			return false
		}
		call, ok := n.(*syntax.MethodCall)
		if !ok || call.Span().End > c.Span().Start {
			return true
		}
		switch method := entryMethod(call, variable); {
		case method == "isdot" || method == "isdir":
			if t := env.CallResultAsAt(call, c, types.Bool); t.IsUnknown() || t.OnlyOf("false") {
				mask = dotEntryNames
			}
		case method == "isfile" || method == "islink":
			if t := env.CallResultAsAt(call, c, types.Bool); t.IsUnknown() || t.OnlyOf("true") {
				mask = dotEntryNames
			}
		case entryName(call, variable):
			t := env.CallResultAsAt(call, c, types.TrackedStrings())
			if t.StringExcludes(".") {
				mask |= 1
			}
			if t.StringExcludes("..") {
				mask |= 2
			}
		}
		return true
	})
	return mask
}

// dotEntryNames is the mask of both dot entries: bit 1 for ".", bit 2
// for "..".
const dotEntryNames uint8 = 3

// entryMethod returns the lower-cased name of a method called directly
// on the loop variable, or "".
func entryMethod(call *syntax.MethodCall, variable string) string {
	id, ok := call.Name.(*syntax.Identifier)
	receiver, rok := call.Var.(*syntax.Variable)
	if !ok || !rok || receiver.Name != variable {
		return ""
	}
	return strings.ToLower(id.Value)
}

// entryName reports whether call reads the entry's own name, which is
// "." or ".." exactly for the dot entries.
func entryName(call *syntax.MethodCall, variable string) bool {
	if call.Args != nil && len(call.Args.Args) != 0 {
		return false
	}
	name := entryMethod(call, variable)
	return name == "getfilename" || name == "getbasename"
}

// iteratorReassigned suppresses a lexical foreach contract after writes
// that can change the element binding before its destructive use.
func iteratorReassigned(loop *syntax.Foreach, at syntax.Node, name string) bool {
	changed := false
	syntax.Inspect(loop.Body, func(node syntax.Node) bool {
		if syntax.IsVariableScope(node) {
			return false
		}
		if node.Span().Start >= at.Span().Start {
			return false
		}
		var targets []syntax.Expr
		switch n := node.(type) {
		case *syntax.Assign:
			targets = []syntax.Expr{n.Var}
		case *syntax.IncDec:
			targets = []syntax.Expr{n.Var}
		case *syntax.Unset:
			targets = n.Vars
		case *syntax.Foreach:
			targets = []syntax.Expr{n.Value, n.Key}
		}
		for _, target := range targets {
			if variable, ok := target.(*syntax.Variable); ok && variable.Name == name {
				changed = true
			}
		}
		return true
	})
	return changed
}
