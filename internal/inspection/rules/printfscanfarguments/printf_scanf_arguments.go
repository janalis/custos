package printfscanfarguments

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// printfScanfArguments checks printf/scanf format strings against the number
// of arguments passed.
type printfScanfArguments struct{}

func (printfScanfArguments) ID() string { return "PrintfScanfArguments" }
func (printfScanfArguments) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

var printfFormatPos = map[string]int{
	"printf": 0, "sprintf": 0, "fprintf": 1, "sscanf": 1, "fscanf": 1,
}

// Conversion characters of each family; printf gained `h`/`H` in PHP 8.0.
const (
	printfConversions = "sducoxXbgGeEfF"
	scanfConversions  = "ndDiouxXfeEgsc"
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
	vals, complete := flowquery.PossibleValuesComplete(ctx.File, a)
	if !complete || len(vals) == 0 {
		return
	}
	scanf := fname == "sscanf" || fname == "fscanf"
	// D3–D6 per candidate; all candidates must agree (custos diverges).
	malformed, expected := false, -1
	for i, v := range vals {
		lit := printfLiteral(v)
		if lit == nil {
			return
		}
		bad, exp, ok := printfVerdict(lit, pos, scanf, ctx.PHP >= phpversion.PHP80)
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
	if scanf && len(args) == 2 && printfUsedAsValue(call) { // D7a
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
// the format is unknown (interpolation) or empty. scanf selects the scanf
// grammar; php8 enables the printf additions of PHP 8.0.
func printfVerdict(lit syntax.Expr, pos int, scanf, php8 bool) (malformed bool, expected int, ok bool) {
	// D3: a format with interpolation is only known at run time.
	l, plainLit := lit.(*syntax.Literal)
	if !plainLit {
		return false, 0, false
	}
	content, _ := astquery.StringLiteralValue(l.Raw)
	content = strings.TrimSpace(content)
	if content == "" {
		return false, 0, false
	}
	c := formatCursor{s: content}
	if scanf {
		ok = c.scanf()
	} else {
		c.s = strings.ReplaceAll(content, "%%", "") // D4
		ok = c.printf(php8)
	}
	if !ok { // D6
		return true, 0, true
	}
	return false, pos + 1 + max(c.plain, c.maxArg), true
}

// formatCursor walks a format string and counts the arguments its
// conversions consume: plain sequential ones and the highest `N$` number.
type formatCursor struct {
	s             string
	i             int
	plain, maxArg int
}

// peek returns the byte under the cursor, 0 at the end of the format.
func (c *formatCursor) peek() byte {
	if c.i < len(c.s) {
		return c.s[c.i]
	}
	return 0
}

func (c *formatCursor) digits() {
	for b := c.peek(); b >= '0' && b <= '9'; b = c.peek() {
		c.i++
	}
}

// argnum consumes an `N$` argument number and records it; without one the
// cursor does not move.
func (c *formatCursor) argnum() bool {
	start := c.i
	c.digits()
	if c.i == start || c.peek() != '$' {
		c.i = start
		return false
	}
	if k, err := strconv.Atoi(c.s[start:c.i]); err == nil && k > c.maxArg {
		c.maxArg = k
	}
	c.i++
	return true
}

// amount consumes a printf width or precision: digits or, when star is
// set, `*` / `*N$`, which reads the amount from an argument.
func (c *formatCursor) amount(star bool) {
	if !star || c.peek() != '*' {
		c.digits()
		return
	}
	c.i++
	if !c.argnum() {
		c.plain++
	}
}

// printf applies the printf-family grammar of D5 to a format whose `%%`
// are already removed; false means a `%` starts no valid specification.
func (c *formatCursor) printf(php8 bool) bool {
	conversions := printfConversions
	if php8 {
		conversions += "hH"
	}
	for ; c.i < len(c.s); c.i++ {
		if c.s[c.i] != '%' {
			continue
		}
		c.i++
		positional := c.argnum()
		for flags := true; flags; {
			switch c.peek() {
			case ' ', '0', '-', '+':
				c.i++
			case '\'':
				c.i += 2 // the quote and its padding character
			default:
				flags = false
			}
		}
		c.amount(php8)
		if c.peek() == '.' {
			c.i++
			c.amount(php8)
		}
		if c.peek() == 'l' {
			c.i++
		}
		if strings.IndexByte(conversions, c.peek()) < 0 {
			return false
		}
		if !positional {
			c.plain++
		}
	}
	return true
}

// scanf applies the scanf-family grammar of D5; false means a `%` starts
// no valid specification.
func (c *formatCursor) scanf() bool {
	for ; c.i < len(c.s); c.i++ {
		if c.s[c.i] != '%' {
			continue
		}
		c.i++
		if c.peek() == '%' {
			continue
		}
		assigned, positional := true, false
		if c.peek() == '*' {
			c.i++
			assigned = false
		} else {
			positional = c.argnum()
		}
		c.digits()
		if b := c.peek(); b == 'h' || b == 'l' || b == 'L' {
			c.i++
		}
		if c.peek() == '[' {
			// A set runs to its closing bracket; a `]` right after the
			// opening (or after `^`) is a member, as is any `%`.
			c.i++
			if c.peek() == '^' {
				c.i++
			}
			if c.peek() == ']' {
				c.i++
			}
			end := strings.IndexByte(c.s[c.i:], ']')
			if end < 0 {
				return false
			}
			c.i += end
		} else if strings.IndexByte(scanfConversions, c.peek()) < 0 {
			return false
		}
		if assigned && !positional {
			c.plain++
		}
	}
	return true
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
	p, _ := astquery.ParentSkipParens(call)
	if _, discarded := p.(*syntax.ExprStmt); discarded {
		return false
	}
	return !astquery.IsLogicalOperand(call)
}
