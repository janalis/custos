package typecompat

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// jsonEncodingApiUsage asks json_decode() for an explicit result type and
// json_decode()/json_encode() for JSON_THROW_ON_ERROR.
type jsonEncodingApiUsage struct{}

func init() { register(jsonEncodingApiUsage{}) }

const (
	jsonTypeMsg  = "Pass the second argument to state whether JSON decodes to arrays or objects."
	jsonThrowMsg = "Pass JSON_THROW_ON_ERROR in the flags of this call."
)

func (jsonEncodingApiUsage) ID() string { return "JsonEncodingApiUsage" }

func (jsonEncodingApiUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (jsonEncodingApiUsage) Semantic() {}

// jsonArg looks an argument up by name, else by unnamed position.
func jsonArg(call *syntax.FuncCall, name string, pos int) (*syntax.Arg, bool) {
	for _, a := range call.Args.Args {
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil && arg.Name.Value == name {
			return arg, true
		}
	}
	if pos < len(call.Args.Args) {
		if arg, ok := call.Args.Args[pos].(*syntax.Arg); ok && arg.Name == nil && !arg.Unpack && arg.Value != nil {
			return arg, true
		}
	}
	return nil, false
}

func jsonNamedArg(call *syntax.FuncCall, name string) bool {
	for _, a := range call.Args.Args {
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil && arg.Name.Value == name {
			return true
		}
	}
	return false
}

func (jsonEncodingApiUsage) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	qual, name, ok := util.CallName(call)
	lname := strings.ToLower(name)
	if !ok || (lname != "json_decode" && lname != "json_encode") || call.Args == nil || len(call.Args.Args) == 0 { // E1
		return
	}
	if !util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, name) { // D0, E2
		return
	}
	decode := lname == "json_decode"
	span := call.Span()
	text := func(a *syntax.Arg) string { return ctx.Text(a.Value) }

	if decode && ctx.Bool("HARDEN_DECODING_RESULT_TYPE") { // D1
		if _, has := jsonArg(call, "associative", 1); !has {
			var fixes []analysis.Fix
			if first, ok := call.Args.Args[0].(*syntax.Arg); ok && first.Value != nil {
				b := "false"
				if ctx.Bool("DECODE_AS_ARRAY") {
					b = "true"
				}
				newText := qual + name + "(" + text(first) + ", " + b + ")"
				fixes = append(fixes, analysis.Fix{Title: "Pass the result type", Edits: func() []analysis.TextEdit {
					return []analysis.TextEdit{{Span: span, NewText: newText}}
				}})
			}
			ctx.Report(span, jsonTypeMsg, fixes...)
		}
	}

	if !ctx.Bool("HARDEN_ERRORS_HANDLING") || ctx.PHP.Below(phpver.PHP73) { // D2
		return
	}
	subjectName, flagsPos := "value", 1
	if decode {
		subjectName, flagsPos = "json", 3
	}
	subject, ok := jsonArg(call, subjectName, 0) // D3
	if !ok {
		return
	}
	flags, hasFlags := jsonArg(call, "flags", flagsPos)
	if hasFlags && jsonStrictFlags(ctx, flags.Value) { // D4, D5
		return
	}
	if jsonNamedArg(call, "flags") { // F4
		ctx.Report(span, jsonThrowMsg)
		return
	}
	newFlags := "JSON_THROW_ON_ERROR"
	if hasFlags {
		newFlags += " | " + text(flags)
	}
	var newText string
	if decode { // F2
		assoc := "false"
		if a, ok := jsonArg(call, "associative", 1); ok {
			assoc = text(a)
		} else if ctx.Bool("HARDEN_DECODING_RESULT_TYPE") && ctx.Bool("DECODE_AS_ARRAY") {
			assoc = "true"
		}
		depth := "512"
		if d, ok := jsonArg(call, "depth", 2); ok {
			depth = text(d)
		}
		newText = qual + name + "(" + text(subject) + ", " + assoc + ", " + depth + ", " + newFlags + ")"
	} else { // F3
		newText = qual + name + "(" + text(subject) + ", " + newFlags
		if d, ok := jsonArg(call, "depth", 2); ok {
			newText += ", " + text(d)
		}
		newText += ")"
	}
	ctx.Report(span, jsonThrowMsg, analysis.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []analysis.TextEdit {
		return []analysis.TextEdit{{Span: span, NewText: newText}}
	}})
}

// jsonStrictFlags implements D5; an unknown discovery result counts as
// strict so that kind E is not reported (D5a).
func jsonStrictFlags(ctx *analysis.Context, f syntax.Expr) bool {
	values, known := util.DiscoverValuesKnown(ctx.Types(), f)
	if !known { // D5a
		return true
	}
	if len(values) != 1 {
		return false
	}
	v := values[0]
	num := v
	if u, ok := v.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		num = u.Expr
	}
	if lit, ok := num.(*syntax.Literal); ok && lit.LitKind != syntax.LitString {
		return lit.Raw == "4194304" || lit.Raw == "512"
	}
	strict := false
	syntax.Inspect(v, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ConstFetch); ok && c.Name != nil {
			switch util.LastNamePart(c.Name.Value) {
			case "JSON_THROW_ON_ERROR", "JSON_PARTIAL_OUTPUT_ON_ERROR":
				strict = true
			}
		}
		return !strict
	})
	return strict
}
