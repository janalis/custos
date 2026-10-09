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
		if args != nil && !puForeignAssertEquals(ctx, n, name) {
			checkEqualsArgs(ctx, name, args.Args)
		}
	case "assertFileNotExists", "assertDirectoryNotExists":
		if ver < 91 { // E2
			return
		}
		repl := renamedAssertions[name]
		span := id.Span()
		ctx.Report(span, name+"() was deprecated by PHPUnit 9.1; call "+repl+"() instead.", diagnostic.Fix{
			Title: "Rename to " + repl + "()",
			Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: span, NewText: repl}}
			},
		})
	}
}

// puForeignAssertEquals reports whether the call may target another API
// than PHPUnit's assertions (custos): a receiver resolving to a class whose
// method `name` is declared outside PHPUnit (e.g. a comparator library's own
// assertEquals with another signature), or an instance call on a receiver
// other than `$this` that does not resolve to PHPUnit's declaration
// (assertions are called on the test case itself).
func puForeignAssertEquals(ctx *analysis.Context, n syntax.Node, name string) bool {
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

var equalsParamIndex = map[string]int{"delta": 3, "maxDepth": 4, "canonicalize": 5, "ignoreCase": 6}

// checkEqualsArgs implements D1/D2.
func checkEqualsArgs(ctx *analysis.Context, name string, args []syntax.Expr) {
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
		ctx.ReportNode(a, msg)
	}
}
