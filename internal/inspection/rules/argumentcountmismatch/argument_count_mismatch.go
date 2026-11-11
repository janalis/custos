package argumentcountmismatch

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

type rule struct{}

func (rule) ID() string { return "ArgumentCountMismatch" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KNew}
}

func (rule) Semantic() {}

// written is what a call passes: its positional arguments and the names
// of its named ones.
type written struct {
	positional []*syntax.Arg
	named      map[string]bool
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	var list *syntax.ArgList
	switch call := n.(type) {
	case *syntax.FuncCall:
		list = call.Args
	case *syntax.MethodCall:
		list = call.Args
	case *syntax.StaticCall:
		list = call.Args
	case *syntax.New:
		list = call.Args
	}
	args, ok := arguments(list)
	if !ok {
		return
	}
	// The agreed contract settles the callee's name, union receivers and
	// repeated function declarations before the targets are listed.
	callee, ok := semanticquery.ResolveCallee(ctx, n)
	if !ok {
		return
	}
	name, targets, ok := semanticquery.CallTargets(ctx, n, callee, func(file, callee string) bool {
		return trustedBuiltin(file, callee, ctx.PHP)
	})
	if !ok {
		return
	}
	missing, _ := bind(targets[0].Params, args)
	limit := 0
	for _, t := range targets {
		lacking, bound := bind(t.Params, args)
		if !bound {
			return
		}
		if len(lacking) == 0 {
			missing = nil
		}
		accepted, capped := maximum(t.Params)
		if !t.Builtin || !capped {
			limit = -1
		} else if limit >= 0 {
			limit = max(limit, accepted)
		}
	}
	if len(missing) > 0 {
		ctx.ReportNode(n, missingMessage(missing))
		return
	}
	// Extra arguments are legal for functions of the project, which may
	// read them with func_get_args().
	if limit < 0 || len(args.positional) <= limit {
		return
	}
	extra := args.positional[limit:]
	ctx.Report(syntax.Span{Start: extra[0].Span().Start, End: extra[len(extra)-1].Span().End}, extraMessage(name, limit, len(extra)))
}

// arguments reads the written arguments; false for a first-class callable,
// an unpacked argument or a malformed list. A `new` without parentheses
// passes none.
func arguments(list *syntax.ArgList) (written, bool) {
	args := written{named: map[string]bool{}}
	if list == nil {
		return args, true
	}
	for _, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Value == nil {
			return written{}, false
		}
		if arg.Name != nil {
			args.named[arg.Name.Value] = true
			continue
		}
		args.positional = append(args.positional, arg)
	}
	return args, true
}

// bind returns the required parameters of params the arguments leave
// unbound; false when a named argument matches no parameter (which another
// rule reports) or is collected by a variadic one.
func bind(params []index.Param, args written) ([]string, bool) {
	known := map[string]bool{}
	for _, p := range params {
		if !p.Variadic {
			known[p.Name] = true
		}
	}
	for name := range args.named {
		if !known[name] {
			return nil, false
		}
	}
	var lacking []string
	for i, p := range params {
		if p.Optional || p.Variadic || i < len(args.positional) || args.named[p.Name] {
			continue
		}
		lacking = append(lacking, p.Name)
	}
	return lacking, true
}

// maximum is the number of positional arguments params accept; false when
// a variadic parameter takes any number.
func maximum(params []index.Param) (int, bool) {
	for _, p := range params {
		if p.Variadic {
			return 0, false
		}
	}
	return len(params), true
}

func missingMessage(missing []string) string {
	names := make([]string, len(missing))
	for i, m := range missing {
		names[i] = "$" + m
	}
	if len(names) == 1 {
		return "Pass the required argument " + names[0] + "."
	}
	return "Pass the required arguments " + strings.Join(names, ", ") + "."
}

func extraMessage(name string, limit, extra int) string {
	what := "argument"
	if extra > 1 {
		what = "arguments"
	}
	return "Remove the extra " + what + ": " + name + "() accepts at most " + strconv.Itoa(limit) + "."
}
