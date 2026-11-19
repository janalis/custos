// Package packformatargumentmismatch implements the native PackFormatArgumentMismatch inspection.
package packformatargumentmismatch

import (
	"strconv"
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const message = "Match pack values to the format."

type rule struct{}

// New returns the PackFormatArgumentMismatch inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PackFormatArgumentMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pack") {
		return
	}
	format, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "format"))
	if !ok || len(format) > 4096 || c.Args == nil {
		return
	}
	if (ctx.PHP.Below(phpversion.PHP56) && strings.ContainsAny(format, "qQJP")) || (ctx.PHP.Below(phpversion.PHP70) && strings.ContainsAny(format, "gGeE")) || (ctx.PHP.Below(phpversion.PHP55) && strings.Contains(format, "Z")) {
		return
	}
	count := 0
	for i, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || (arg.Name != nil && (i != 0 || arg.Name.Value != "format")) {
			return
		}
		if i > 0 {
			count++
		}
	}
	required, known := requiredValues(format, count)
	if known && required > count {
		ctx.ReportNode(c, message)
	}
}

func requiredValues(format string, available int) (int, bool) {
	total := 0
	for i := 0; i < len(format); {
		code := format[i]
		i++
		start := i
		for i < len(format) && format[i] >= '0' && format[i] <= '9' {
			i++
		}
		count := 1
		if start != i {
			v, err := strconv.Atoi(format[start:i])
			if err != nil || v > 1000000 {
				return 0, false
			}
			count = v
		}
		star := i < len(format) && format[i] == '*'
		if star {
			i++
		}
		switch {
		case strings.ContainsRune("aAZhH", rune(code)):
			total++
		case strings.ContainsRune("xX@", rune(code)):
			if star {
				return 0, false
			}
		case strings.ContainsRune("cCsSnNvViIlLqQJPfFdDeEgG", rune(code)):
			if star {
				total = max(total, available)
			} else {
				total += count
			}
		default:
			return 0, false
		}
	}
	return total, true
}
