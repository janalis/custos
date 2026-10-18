// Package csvescapedefaultdependency implements the native CsvEscapeDefaultDependency inspection.
package csvescapedefaultdependency

import (
	"slices"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Pass the CSV escape character explicitly."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CsvEscapeDefaultDependency" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	if ctx.PHP < phpversion.PHP84 {
		return
	}
	args, pos, _, ok := semanticquery.NativeCSVSignature(ctx, n)
	if !ok {
		return
	}
	escape := pos + 2
	if _, state := semanticquery.CallArgumentBinding(args, escape, "escape"); state != semanticquery.ArgumentMissing {
		return
	}
	fixable := true
	if c, ok := n.(*syntax.MethodCall); ok {
		if id := c.Name.(*syntax.Identifier); !strings.EqualFold(id.Value, "setCsvControl") {
			// Every possible control update must retain an explicit escape.
			history := semanticquery.NativeCallHistory(ctx, c, c.Var, "setCsvControl")
			explicit := false
			if history.Complete {
				for _, event := range history.Events {
					control := event.Node.(*syntax.MethodCall)
					_, state := semanticquery.CallArgumentBinding(control.Args, 2, "escape")
					provided := state == semanticquery.ArgumentBound
					if event.Must {
						explicit = provided
					} else if !provided {
						explicit = false
					}
				}
			}
			if explicit {
				return
			}
			fixable = defaultEscapeProven(ctx, c)
		}
	}
	if !fixable {
		ctx.ReportNode(n, message)
		return
	}
	point := syntax.Span{Start: args.Span().End - 1, End: args.Span().End - 1}
	text := `escape: "\\"`
	if len(args.Args) > 0 {
		trailing := false
		last := args.Args[len(args.Args)-1].Span().End
		for _, tok := range ctx.File.Tokens {
			if tok.Start >= last && tok.End <= point.Start && tok.Kind == syntax.TComma {
				trailing = true
			}
		}
		if !trailing {
			text = ", " + text
		}
	}
	fix := diagnostic.Fix{Title: "Pass the CSV escape explicitly", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: point, NewText: text}} }}
	ctx.ReportNode(n, message, fix)
}

// defaultEscapeProven reports whether the receiver's control escape is
// certainly still the backslash default: a variable assigned once from a new
// builtin file object and otherwise used only as the receiver of builtin
// methods that leave the escape alone. Anything else may have configured it.
func defaultEscapeProven(ctx *analysis.Context, call *syntax.MethodCall) bool {
	receiver, ok := call.Var.(*syntax.Variable)
	if !ok || receiver.Name == "" {
		return false
	}
	assigned, safe := false, true
	visit := func(n syntax.Node) bool {
		if !safe {
			return false
		}
		switch v := n.(type) {
		case *syntax.Include, *syntax.Eval:
			safe = false
		case *syntax.Variable:
			if v.Name != receiver.Name {
				// A variable variable may alias the receiver.
				safe = v.Name != ""
				break
			}
			switch parent := v.Parent().(type) {
			case *syntax.Assign:
				created, isNew := parent.Value.(*syntax.New)
				safe = parent.Var == v && parent.Op.Kind == syntax.TEqual && !parent.ByRef && isNew && !assigned && builtinFileObject(ctx, created) && semanticquery.NativeDominates(parent, call)
				assigned = true
			case *syntax.MethodCall:
				id, named := parent.Name.(*syntax.Identifier)
				safe = parent.Var == v && named && !(strings.EqualFold(id.Value, "setCsvControl") && setsEscape(parent.Args))
			default:
				safe = false
			}
		}
		return safe
	}
	if scope := syntax.EnclosingVariableScope(call); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return safe && assigned
}

// builtinFileObject accepts only the exact builtin classes: a subclass may
// configure the escape in its constructor or in an overridden method.
func builtinFileObject(ctx *analysis.Context, n *syntax.New) bool {
	classes := ctx.Types().Native().TypeOf(n).Classes()
	return len(classes) == 1 && slices.ContainsFunc([]string{"SplFileObject", "SplTempFileObject"}, func(name string) bool {
		return strings.EqualFold(strings.TrimPrefix(classes[0], `\`), name)
	})
}

// setsEscape reports whether a setCsvControl call may supply an escape.
func setsEscape(args *syntax.ArgList) bool {
	if semanticquery.CallArgument(args, 2, "escape") != nil {
		return true
	}
	for _, a := range args.Args {
		if arg, ok := a.(*syntax.Arg); !ok || arg.Unpack {
			return true
		}
	}
	return false
}
