// Package jsonintegerprecisionloss implements the native JsonIntegerPrecisionLoss inspection.
package jsonintegerprecisionloss

import (
	"encoding/json"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Preserve large JSON integers as strings."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "JsonIntegerPrecisionLoss" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "json_decode") {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "json"))
	if !known || !json.Valid([]byte(text)) {
		return
	}
	flags := semanticquery.CallArgument(call.Args, 3, "flags")
	present, known := semanticquery.NativeFlag(ctx, flags, 2)
	if !known || present {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var data any
	// Validated bytes and UseNumber guarantee decoding succeeds.
	_ = decoder.Decode(&data)
	var large func(any) bool
	large = func(v any) bool {
		switch x := v.(type) {
		case json.Number:
			if strings.ContainsAny(string(x), ".eE") {
				return false
			}
			_, err := strconv.ParseInt(string(x), 10, 64)
			return err != nil
		case []any:
			for _, item := range x {
				if large(item) {
					return true
				}
			}
		case map[string]any:
			for _, item := range x {
				if large(item) {
					return true
				}
			}
		}
		return false
	}
	if !large(data) {
		return
	}
	if ctx.PHP < phpversion.PHP54 {
		ctx.ReportNode(call, message)
		return
	}
	var edits []diagnostic.TextEdit
	if flags != nil {
		edits = []diagnostic.TextEdit{{Span: flags.Span(), NewText: "(" + ctx.Text(flags) + ") | \\JSON_BIGINT_AS_STRING"}}
	} else {
		args := astquery.WrittenArguments(call.Args)
		end := args[len(args)-1].Span().End
		named := false
		for _, arg := range args {
			if arg.Name != nil {
				named = true
			}
		}
		text := ", \\JSON_BIGINT_AS_STRING"
		if named {
			text = ", flags: \\JSON_BIGINT_AS_STRING"
		} else {
			if len(args) < 2 {
				text = ", false" + text
			}
			if len(args) < 3 {
				text = strings.TrimSuffix(text, ", \\JSON_BIGINT_AS_STRING") + ", 512, \\JSON_BIGINT_AS_STRING"
			}
		}
		edits = []diagnostic.TextEdit{{Span: syntax.Span{Start: end, End: end}, NewText: text}}
	}
	ctx.ReportNode(call, message, diagnostic.Fix{Title: "Decode large integers exactly", Edits: func() []diagnostic.TextEdit { return edits }})
}
