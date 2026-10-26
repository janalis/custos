package phpunitdeprecations

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/phpunit"
	"custos/internal/php/syntax"
)

// phpUnitDeprecations reports PHPUnit assertion APIs deprecated in the
// configured PHPUnit version.
type phpUnitDeprecations struct{}

func (phpUnitDeprecations) ID() string { return "PhpUnitDeprecations" }
func (phpUnitDeprecations) Semantic()  {}
func (phpUnitDeprecations) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

// phpUnitVersion parses PHP_UNIT_VERSION (`PHPUNIT91` -> 91); when unset it
// is inferred from the indexed PHPUnit, else (or when malformed) it is 80.
func phpUnitVersion(ctx *analysis.Context) int {
	if !ctx.OptionSet("PHP_UNIT_VERSION") {
		if v, ok := phpunit.DetectedPHPUnitVersion(ctx); ok {
			return v
		}
	}
	v := strings.ToUpper(ctx.String("PHP_UNIT_VERSION"))
	if n, err := strconv.Atoi(strings.TrimPrefix(v, "PHPUNIT")); err == nil && strings.HasPrefix(v, "PHPUNIT") {
		return n
	}
	return 80
}

var renamedAssertions = map[string]string{
	"assertFileNotExists":      "assertFileDoesNotExist",
	"assertDirectoryNotExists": "assertDirectoryDoesNotExist",
}

func (phpUnitDeprecations) Check(ctx *analysis.Context, n syntax.Node) {
	var nameExpr syntax.Expr
	var args *syntax.ArgList
	switch c := n.(type) {
	case *syntax.MethodCall:
		nameExpr, args = c.Name, c.Args
	case *syntax.StaticCall:
		nameExpr, args = c.Name, c.Args
	}
	id, ok := nameExpr.(*syntax.Identifier)
	if !ok {
		return
	}
	ver := phpUnitVersion(ctx)
	if ver < 80 { // E1
		return
	}
	name := phpunit.PuName(id.Value) // method names are case-insensitive
	switch name {
	case "assertEquals", "assertNotEquals":
		if args != nil && !puForeignAssertion(ctx, n, name) {
			checkEqualsArgs(ctx, n, name, args.Args)
		}
	case "assertFileNotExists", "assertDirectoryNotExists":
		repl := renamedAssertions[name]
		if ver < 91 || puForeignAssertion(ctx, n, name) || puCompatibilityLayer(ctx, n, repl) { // E2, E5, E6
			return
		}
		span := id.Span()
		ctx.Report(span, name+"() was deprecated by PHPUnit 9.1; call "+repl+"() instead.", diagnostic.Fix{
			Title: "Rename to " + repl + "()",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: span, NewText: repl}}
			},
		})
	}
}

// puForeignAssertion reports whether the call may target another API
// than PHPUnit's assertions (custos): a receiver resolving to a class whose
// method `name` is declared outside PHPUnit (e.g. a comparator library's own
// assertEquals with another signature, or a storage class's own
// assertFileNotExists, which the rename would break), or an instance call on a receiver
// other than `$this` that does not resolve to PHPUnit's declaration
// (assertions are called on the test case itself).
func puForeignAssertion(ctx *analysis.Context, n syntax.Node, name string) bool {
	var classes []string
	switch c := n.(type) {
	case *syntax.MethodCall:
		classes = ctx.TypeOf(c.Var).Classes()
		if v, ok := c.Var.(*syntax.Variable); (!ok || v.Name != "this") && len(classes) == 0 {
			return true
		}
	case *syntax.StaticCall:
		if fqn := ctx.Types().ClassRef(c.Class); fqn != "" {
			classes = []string{fqn}
		}
	}
	for _, cl := range classes {
		m := ctx.Index().FindMethod(cl, name, ctx.PHP)
		if m == nil || strings.HasPrefix(strings.ToLower(strings.TrimPrefix(m.Class, `\`)), `phpunit\`) {
			return false
		}
	}
	return len(classes) > 0
}

var equalsReplacements = map[int]string{3: "WithDelta", 5: "Canonicalizing", 6: "IgnoringCase"}

// puCompatibilityLayer implements E6: the call belongs to code that keeps
// older PHPUnit versions working, so the replacement API repl may not exist
// there. It sits in a method named repl (a polyfill, which the replacement
// would turn into a method calling itself), or under an `if`, `elseif` or
// ternary whose condition asks `method_exists(…, 'repl')`.
func puCompatibilityLayer(ctx *analysis.Context, call syntax.Node, repl string) bool {
	for p := call.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		var conds []syntax.Expr
		switch x := p.(type) {
		case *syntax.If:
			conds = append(conds, x.Cond)
			for _, ei := range x.ElseIfs {
				conds = append(conds, ei.Cond)
			}
		case *syntax.Ternary:
			conds = append(conds, x.Cond)
		}
		for _, c := range conds {
			if puAsksMethodExists(ctx, c, repl) {
				return true
			}
		}
	}
	m, ok := syntax.EnclosingFuncLike(call).(*syntax.Method)
	return ok && strings.EqualFold(m.Name.Value, repl)
}

// puAsksMethodExists reports whether cond calls the builtin method_exists()
// with the literal method name repl.
func puAsksMethodExists(ctx *analysis.Context, cond syntax.Expr, repl string) bool {
	found := false
	syntax.Inspect(cond, func(n syntax.Node) bool {
		c, ok := n.(*syntax.FuncCall)
		if !ok || found || !ctx.IsGlobalFunctionCall(c, "method_exists") || c.Args == nil || len(c.Args.Args) < 2 {
			return !found
		}
		if arg, ok := c.Args.Args[1].(*syntax.Arg); ok {
			if lit, ok := arg.Value.(*syntax.Literal); ok && lit.LitKind == syntax.LitString && len(lit.Raw) > 2 {
				found = strings.EqualFold(lit.Raw[1:len(lit.Raw)-1], repl)
			}
		}
		return !found
	})
	return found
}

var equalsParamIndex = map[string]int{"delta": 3, "maxDepth": 4, "canonicalize": 5, "ignoreCase": 6}

// checkEqualsArgs implements D1/D2.
func checkEqualsArgs(ctx *analysis.Context, call syntax.Node, name string, args []syntax.Expr) {
	many := len(args) > 3
	for i, a := range args {
		idx := i
		if arg, ok := a.(*syntax.Arg); ok && arg.Name != nil {
			p, known := equalsParamIndex[arg.Name.Value]
			if !known {
				continue
			}
			idx = p
		} else if !many {
			continue
		}
		var msg string
		switch idx {
		case 3:
			msg = "PHPUnit 8.0 deprecated the delta argument; call " + name + "WithDelta() instead."
		case 4:
			msg = "PHPUnit 8.0 deprecated the maxDepth argument; drop it."
		case 5:
			msg = "PHPUnit 8.0 deprecated the canonicalize argument; call " + name + "Canonicalizing() instead."
		case 6:
			msg = "PHPUnit 8.0 deprecated the ignoreCase argument; call " + name + "IgnoringCase() instead."
		default:
			continue
		}
		if repl := equalsReplacements[idx]; repl != "" && puCompatibilityLayer(ctx, call, name+repl) { // E6
			continue
		}
		ctx.ReportNode(a, msg)
	}
}
