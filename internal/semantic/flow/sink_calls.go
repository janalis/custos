package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// nativeSinks selects resolved native sinks before binding their arguments.
func (e *Env) nativeSinks(c Call) []Sink {
	name, builtin := e.callName(c.Node)
	if !builtin {
		return nil
	}
	context := Context(0)
	position := 0
	switch c.Node.(type) {
	case *syntax.MethodCall:
		_, qualified, resolved := e.declaration(c.Node)
		if resolved && strings.HasPrefix(qualified, "pdo::") && (name == "query" || name == "exec" || name == "prepare") {
			context = SQL
		}
	case *syntax.FuncCall:
		switch name {
		case "system", "exec", "shell_exec", "passthru", "popen", "proc_open":
			context = Shell
		case "header":
			context = Header
		case "readfile", "unlink", "file_get_contents", "fopen", "file_put_contents", "rmdir":
			context = Path
		case "curl_init":
			context = URL
		case "curl_setopt":
			context = URL
			position = 2
		}
	}
	if context == 0 {
		return nil
	}
	args, known := e.bind(c.Node, c.Arguments)
	if !known {
		// A reference argument, which a command fills with its output,
		// leaves the first parameter what it was given.
		args, known = e.leading(c)
	}
	if !known || len(args) == 0 {
		return nil
	}
	if position == 2 {
		if len(args) <= 2 {
			return nil
		}
		constant, ok := args[1].Expr.(*syntax.ConstFetch)
		if !ok || !strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "CURLOPT_URL") {
			return nil
		}
	}
	if call, function := c.Node.(*syntax.FuncCall); function && name == "proc_open" && e.vector(call, args[0]) {
		// An array is an argument vector: no shell reads it.
		return nil
	}
	value := args[position]
	value.targets = nil
	value.returned = nil
	value.Expr = nil
	value.elements = nil
	value.interval = nil
	value.origins = nil
	value.Identity = 0
	sinks := []Sink{{Name: name, Context: context, Argument: value, Path: e.file.Path, Span: c.Node.Span()}}
	if name == "header" {
		if binary, ok := args[0].Expr.(*syntax.Binary); ok && binary.Op.Kind == syntax.TDot {
			if prefix, known := literalText(binary.Left); known && strings.HasPrefix(strings.ToLower(prefix), "location:") {
				sinks = append(sinks, Sink{Name: name, Context: URL, Argument: value, Path: e.file.Path, Span: c.Node.Span()})
			}
		}
	}
	return sinks
}

// vector reports a `proc_open` command that is an array: an array literal, a
// variable holding one on every path, or an expression typed as an array.
func (e *Env) vector(call *syntax.FuncCall, command Value) bool {
	if _, literal := syntax.UnwrapParens(command.Expr).(*syntax.Array); literal {
		return true
	}
	var few [8]int
	var written syntax.Expr
	params, _, _ := e.declaration(call)
	for i, at := range parameters(call.Args, params, few[:0]) {
		if arg, ok := call.Args.Args[i].(*syntax.Arg); ok && at == 0 {
			written = arg.Value
		}
	}
	return e.types.TypeOf(written).IsArrayLike()
}

// leading returns the value given to the first parameter of a call whose
// arguments do not all bind by value.
func (e *Env) leading(c Call) ([]Value, bool) {
	call, ok := c.Node.(*syntax.FuncCall)
	params, _, known := e.declaration(c.Node)
	if !ok || !known || call.Args == nil {
		return nil, false
	}
	var few [8]int
	for i, at := range parameters(call.Args, params, few[:0]) {
		if arg, written := call.Args.Args[i].(*syntax.Arg); written && at == 0 && !arg.Unpack && !arg.ByRef {
			return c.Arguments[i : i+1], true
		}
	}
	return nil, false
}

// commands lists the backtick strings a body runs, outside the functions
// declared in it: each is a command given to the shell.
func (e *Env) commands(body syntax.Node, r *scopeResult) []Sink {
	var sinks []Sink
	syntax.Inspect(body, func(n syntax.Node) bool {
		if n != body && syntax.IsVariableScope(n) {
			return false
		}
		if command, ok := n.(*syntax.InterpolatedString); ok && command.Backtick {
			if value, reached := r.values.get(command); reached {
				value.Expr, value.origins, value.Identity = nil, nil, 0
				sinks = append(sinks, Sink{Name: "shell_exec", Context: Shell, Argument: value, Path: e.file.Path, Span: command.Span()})
			}
		}
		return true
	})
	return sinks
}

// Sinks expands known project wrapper sink contracts at this call site.
// The argument sources and escaping contexts come from this call's flow.
func (e *Env) Sinks(at syntax.Node) []Sink {
	if sinks, known := e.scope(syntax.EnclosingVariableScope(at)).closureSinks[at]; known {
		return sinks
	}
	name := e.sinkSummaryName(at)
	summary, ok := e.contract(name)
	if !ok {
		return e.closedSinks(at)
	}
	args, known := e.BoundArguments(at)
	if !known {
		return nil
	}
	out := make([]Sink, 0, len(summary.Sinks))
	for _, s := range summary.Sinks {
		v := substitute(s.Argument, args, true)
		s.Argument = v
		out = append(out, s)
	}
	return out
}
