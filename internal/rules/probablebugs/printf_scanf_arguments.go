package probablebugs

import (
	"regexp"
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// printfScanfArguments checks printf/scanf format strings against the number
// of arguments passed.
type printfScanfArguments struct{}

func init() { register(printfScanfArguments{}) }

func (printfScanfArguments) ID() string { return "PrintfScanfArguments" }

func (printfScanfArguments) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var printfFormatPos = map[string]int{
	"printf": 0, "sprintf": 0, "fprintf": 1, "sscanf": 1, "fscanf": 1,
}

var (
	printfSpec = regexp.MustCompile(`%(?:(\d+)\$)?[+-]?(?: |0|\\?'.)?-?\d*(?:\.\d*)?l?[\[sducoxXbgGeEfF]`)
)

func (printfScanfArguments) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	fname := ctx.GlobalFunctionName(call) // D1: case-insensitive, global function only
	pos, ok := printfFormatPos[fname]
	if !ok || len(call.Args.Args) < pos+1 {
		return
	}
	arg, ok := call.Args.Args[pos].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return
	}
	a := arg.Value

	// D2: every possible format must be a known literal (custos diverges).
	vals, complete := util.PossibleValuesComplete(ctx.File, a)
	if !complete || len(vals) == 0 {
		return
	}
	// D3–D6 per candidate; all candidates must agree (custos diverges).
	malformed, expected := false, -1
	for i, v := range vals {
		lit := printfLiteral(v)
		if lit == nil {
			return
		}
		bad, exp, ok := printfVerdict(lit, pos)
		if !ok || (i > 0 && (bad != malformed || exp != expected)) {
			return
		}
		malformed, expected = bad, exp
	}
	if malformed {
		ctx.ReportNode(a, "Malformed format string.")
		return
	}
	// D7
	args := call.Args.Args
	if len(args) == expected {
		return
	}
	if (fname == "sscanf" || fname == "fscanf") && len(args) == 2 && printfUsedAsValue(call) { // D7a
		return
	}
	if last, ok := args[len(args)-1].(*syntax.Arg); ok && last.Unpack { // D7b
		return
	}
	// D7c (compound-assigned format variable) is subsumed by D2: value
	// discovery yields no complete value set for such a variable.
	ctx.ReportNode(call.Name, "This call needs "+strconv.Itoa(expected)+" argument(s) in total.")
}

// printfVerdict applies D3–D6 to one candidate literal: whether the format
// is malformed and, if not, the expected argument count. ok is false when
// the format is unknown (interpolation) or empty.
func printfVerdict(lit syntax.Expr, pos int) (malformed bool, expected int, ok bool) {
	// D3: a format with interpolation is only known at run time.
	if _, interp := lit.(*syntax.InterpolatedString); interp {
		return false, 0, false
	}
	content, ok := printfContent(lit)
	if !ok {
		return false, 0, false
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return false, 0, false
	}
	adapted := strings.ReplaceAll(content, "%%", "") // D4
	// D5
	matches, plain, maxArg := 0, 0, 0
	for _, m := range printfSpec.FindAllStringSubmatch(adapted, -1) {
		matches++
		if m[1] == "" {
			plain++
		} else if k, err := strconv.Atoi(m[1]); err == nil && k > maxArg {
			maxArg = k
		}
	}
	// D6
	if matches != strings.Count(strings.ReplaceAll(adapted, "%*", ""), "%") {
		return true, 0, true
	}
	return false, pos + 1 + max(plain, maxArg), true
}

// printfContent returns the format text of a quoted literal: the raw
// contents of a single-quoted string, the decoded value of a double-quoted
// one (so `"%1\$s"` reads as `%1$s`).
func printfContent(lit syntax.Expr) (string, bool) {
	raw, quote, ok := util.QuotedStringRaw(lit)
	if !ok {
		return "", false
	}
	if quote == '"' {
		return util.StringLiteralValue(lit.(*syntax.Literal).Raw)
	}
	return raw, true
}

// printfLiteral returns e when it is a quoted string literal.
func printfLiteral(e syntax.Expr) syntax.Expr {
	switch l := e.(type) {
	case *syntax.Literal:
		if l.LitKind == syntax.LitString && len(l.Raw) >= 2 && (l.Raw[0] == '\'' || l.Raw[0] == '"') {
			return l
		}
	case *syntax.InterpolatedString:
		if !l.Heredoc && !l.Backtick {
			return l
		}
	}
	return nil
}

// printfUsedAsValue implements D7a: custos accepts any use of the result
// as a value (`sscanf($t, $f) ?? []`, `return sscanf(…)`, an element of an
// array), not only assignments and arguments; a discarded call or one used
// as a truth value (`if (sscanf($s, '%d'))`, always a non-empty array) is
// still reported.
func printfUsedAsValue(call *syntax.FuncCall) bool {
	p, _ := util.ParentSkipParens(call)
	if _, discarded := p.(*syntax.ExprStmt); discarded {
		return false
	}
	return !util.IsLogicalOperand(call)
}
