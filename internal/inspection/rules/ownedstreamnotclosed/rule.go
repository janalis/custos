// Package ownedstreamnotclosed implements the native OwnedStreamNotClosed inspection.
package ownedstreamnotclosed

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Close the owned stream before leaving its scope."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OwnedStreamNotClosed" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KReturn} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	ret := n.(*syntax.Return)

	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		open, ok := call.Node.(*syntax.FuncCall)
		if !ok || call.Name != "fopen" || open.Span().End >= ret.Span().Start {
			continue
		}
		assignment, ok := open.Parent().(*syntax.Assign)
		if !ok {
			continue
		}
		variable, ok := assignment.Var.(*syntax.Variable)
		if !ok {
			continue
		}
		var use syntax.Expr
		if ret.Expr != nil {
			syntax.Inspect(ret.Expr, func(child syntax.Node) bool {
				if v, ok := child.(*syntax.Variable); ok && v.Name == variable.Name {
					use = v
				}
				return true
			})
		}
		if use == nil {
			continue
		}
		if _, ok := ret.Expr.(*syntax.Variable); ok {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, use)
		if !known || state.Operation == "fclose" || semanticquery.CalledInFinally(ctx, ret, use, "fclose") {
			continue
		}
		if !ctx.Flow().Excludes(use, "false") {
			continue
		}
		if ownedStreamReleased(ctx, syntax.EnclosingVariableScope(ret), ret, variable.Name) {
			continue
		}
		ctx.ReportNode(ret, message)
		return
	}
}

// ownedStreamReleased reports whether the handle held by the named local may
// be closed by the return expression itself or may outlive the scope: any use
// other than a test, a rebinding or a borrowing builtin stream call hands the
// handle to code that can keep or close it.
func ownedStreamReleased(ctx *analysis.Context, scope syntax.Node, ret *syntax.Return, name string) bool {
	released := false
	syntax.Inspect(scope, func(n syntax.Node) bool {
		if released {
			return false
		}
		switch n := n.(type) {
		case *syntax.Closure:
			if n == scope {
				return true
			}
			for _, captured := range n.Uses {
				released = released || captured.Var.Name == name
			}
			return false
		case *syntax.ArrowFunction:
			// An arrow function captures every outer local its body reads.
			syntax.Inspect(n.Expr, func(inner syntax.Node) bool {
				v, ok := inner.(*syntax.Variable)
				released = released || (ok && v.Name == name)
				return true
			})
			return false
		case *syntax.Variable:
			if n.Name != name {
				return true
			}
			switch ownedStreamUse(ctx, n) {
			case ownedStreamEscapes:
				released = true
			case ownedStreamCloses:
				released = n.Span().Start >= ret.Span().Start && n.Span().End <= ret.Span().End
			}
		}
		return n == scope || !syntax.IsVariableScope(n)
	})
	return released
}

// ownedStreamUse classifies one read or write of the local holding a handle.
func ownedStreamUse(ctx *analysis.Context, v *syntax.Variable) int {
	var child syntax.Node = v
	parent := v.Parent()
	for {
		paren, ok := parent.(*syntax.Paren)
		if !ok {
			break
		}
		child, parent = paren, paren.Parent()
	}
	switch p := parent.(type) {
	case *syntax.Assign:
		// Rebinding the local is not a use of the handle; copying it is.
		if p.Var == child && !p.ByRef {
			return ownedStreamBorrowed
		}
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual,
			syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return ownedStreamBorrowed
		}
	case *syntax.Unary:
		if p.Op.Kind == syntax.TExclaim {
			return ownedStreamBorrowed
		}
	case *syntax.If, *syntax.ElseIf, *syntax.While, *syntax.DoWhile:
		return ownedStreamBorrowed
	case *syntax.Ternary:
		// A full ternary only tests its condition; `$h ?: x` yields the handle.
		if p.Cond == child && p.Then != nil {
			return ownedStreamBorrowed
		}
	case *syntax.Arg:
		call, ok := p.Parent().Parent().(*syntax.FuncCall)
		if !ok || p.ByRef || p.Unpack {
			return ownedStreamEscapes
		}
		switch semanticquery.NativeBuiltinName(ctx, call) {
		case "fclose":
			return ownedStreamCloses
		case "fread", "fwrite", "fputs", "fgets", "fgetc", "fgetcsv", "fputcsv", "fscanf", "fprintf", "vfprintf",
			"fpassthru", "feof", "fflush", "fsync", "fdatasync", "fseek", "ftell", "rewind", "ftruncate", "fstat", "flock",
			"stream_get_contents", "stream_get_line", "stream_get_meta_data", "stream_copy_to_stream",
			"stream_set_blocking", "stream_set_timeout", "stream_set_chunk_size", "stream_set_read_buffer",
			"stream_set_write_buffer", "stream_isatty", "stream_is_local", "stream_supports_lock",
			"is_resource", "get_resource_type", "get_resource_id":
			return ownedStreamBorrowed
		}
	}
	return ownedStreamEscapes
}

const (
	ownedStreamBorrowed = iota
	ownedStreamCloses
	ownedStreamEscapes
)
