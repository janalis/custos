// Package validatedbooleanfalserejectedasinvalid implements the native ValidatedBooleanFalseRejectedAsInvalid inspection.
package validatedbooleanfalserejectedasinvalid

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish valid false from invalid boolean input."

const (
	filterValidateBool = 258
	// Flags that change the meaning or shape of the result: the array flags
	// reject a scalar input or wrap the result, so false is not "valid false".
	filterRequireArray  = 16777216
	filterForceArray    = 67108864
	filterNullOnFailure = 134217728
	// The boolean filter strips these ASCII bytes only; any other leading or
	// trailing whitespace (form feed, NBSP, ...) makes the input invalid.
	filterTrimSet = " \t\n\r\v"
)

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ValidatedBooleanFalseRejectedAsInvalid" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsIdentical {
		return
	}
	c, ok := syntax.UnwrapParens(b.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "filter_var") {
		return
	}
	sentinel, k := syntax.UnwrapParens(b.Right).(*syntax.ConstFetch)
	if !k || semanticquery.GlobalConstName(ctx, sentinel) != "false" {
		return
	}
	filter, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(c.Args, 1, "filter"))
	if !known || filter != filterValidateBool {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 2, "options")
	if flags != nil {
		bits, fk := semanticquery.NativeContractInt(ctx, flags)
		if !fk || bits&(filterNullOnFailure|filterRequireArray|filterForceArray) != 0 {
			return
		}
	}
	value, vk := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "value"))
	if !vk {
		return
	}
	switch strings.ToLower(strings.Trim(value, filterTrimSet)) {
	case "", "0", "false", "off", "no":
	default:
		return
	}
	branch, ok := b.Parent().(*syntax.If)
	if ok && semanticquery.ExpansionDFinalOutcome(ctx, branch.Body, false) {
		ctx.ReportNode(b, message)
	}
}
