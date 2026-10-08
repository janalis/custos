package typecompat

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// jsonEncodingAPIUsage asks json_decode() for an explicit result type and
// json_decode()/json_encode() for JSON_THROW_ON_ERROR.
type jsonEncodingAPIUsage struct{}

func init() { register(jsonEncodingAPIUsage{}) }

const (
	jsonTypeMsg  = "Pass the second argument to state whether JSON decodes to arrays or objects."
	jsonThrowMsg = "Pass JSON_THROW_ON_ERROR in the flags of this call."
)

func (jsonEncodingAPIUsage) ID() string { return "JsonEncodingApiUsage" }

func (jsonEncodingAPIUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (jsonEncodingAPIUsage) Semantic() {}

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

func (jsonEncodingAPIUsage) Check(ctx *analysis.Context, n syntax.Node) {
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
			b := "false"
			if ctx.Bool("DECODE_AS_ARRAY") {
				b = "true"
			}
			if jsonHasNamedArg(call) { // F1b: keep the named arguments
				at := call.Args.Args[len(call.Args.Args)-1].Span().End
				fixes = append(fixes, analysis.Fix{Title: "Pass the result type", Edits: func() []analysis.TextEdit {
					return []analysis.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: ", associative: " + b}}
				}})
			} else if first, ok := call.Args.Args[0].(*syntax.Arg); ok && first.Value != nil {
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
	throwConst := jsonThrowConst(ctx, span.Start)
	if hasFlags && flags.Name != nil { // F4: add the constant to the named flags
		val := flags.Value
		old := ctx.Text(val)
		if jsonNeedsParens(val) {
			old = "(" + old + ")"
		}
		ctx.Report(span, jsonThrowMsg, analysis.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: val.Span(), NewText: throwConst + " | " + old}}
		}})
		return
	}
	if jsonHasNamedArg(call) { // F5: append a named flags argument
		at := call.Args.Args[len(call.Args.Args)-1].Span().End
		ctx.Report(span, jsonThrowMsg, analysis.Fix{Title: "Add JSON_THROW_ON_ERROR", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: syntax.Span{Start: at, End: at}, NewText: ", flags: " + throwConst}}
		}})
		return
	}
	newFlags := throwConst
	if hasFlags {
		old := text(flags)
		if jsonNeedsParens(flags.Value) {
			old = "(" + old + ")"
		}
		newFlags += " | " + old
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

// jsonNeedsParens reports whether flags must be parenthesised as the right
// operand of `JSON_THROW_ON_ERROR | …`: operators binding more loosely
// than `|` (logical, `??`), ternaries, assignments and the keyword
// expressions. Bitwise, comparison and arithmetic operands keep their value.
func jsonNeedsParens(e syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.Binary:
		switch x.Op.Kind {
		case syntax.TOr, syntax.TXor, syntax.TAnd, syntax.TCoalesce, syntax.TBooleanOr, syntax.TBooleanAnd:
			return true
		}
		return false
	case *syntax.Ternary, *syntax.Assign, *syntax.Print, *syntax.Yield,
		*syntax.YieldFrom, *syntax.Include, *syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}

// jsonHasNamedArg reports whether the call passes any argument by name.
func jsonHasNamedArg(call *syntax.FuncCall) bool {
	for _, a := range call.Args.Args {
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil {
			return true
		}
	}
	return false
}

// jsonThrowConst spells JSON_THROW_ON_ERROR for a fix at offset at (F6): with
// a leading backslash when the file already writes a global constant that
// way (`\PHP_EOL`, `\JSON_PRETTY_PRINT`), or when a bare name there would
// not reach the global constant.
func jsonThrowConst(ctx *analysis.Context, at uint32) string {
	const name = "JSON_THROW_ON_ERROR"
	if q := util.QualifiedGlobalConst(ctx, name, at); q != name {
		return q
	}
	qualified := ctx.Memo("qualifiedConsts", func() any {
		found := false
		syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.ConstFetch); ok && c.Name != nil && strings.HasPrefix(c.Name.Value, `\`) && util.GlobalConstName(ctx, c) != "" {
				switch strings.ToLower(c.Name.Value) {
				case `\true`, `\false`, `\null`:
				default:
					found = true
				}
			}
			return !found
		})
		return found
	}).(bool)
	if qualified {
		return `\` + name
	}
	return name
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
