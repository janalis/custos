package phpunit

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// phpUnitDeprecations reports PHPUnit assertion APIs deprecated in the
// configured PHPUnit version.
type phpUnitDeprecations struct{}

func init() { register(phpUnitDeprecations{}) }

func (phpUnitDeprecations) ID() string { return "PhpUnitDeprecations" }

func (phpUnitDeprecations) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

// phpUnitVersion parses PHP_UNIT_VERSION (`PHPUNIT91` -> 91); when unset it
// is inferred from the indexed PHPUnit, else (or when malformed) it is 80.
func phpUnitVersion(ctx *analysis.Context) int {
	if !ctx.OptionSet("PHP_UNIT_VERSION") {
		if v, ok := detectedPHPUnitVersion(ctx); ok {
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
	name := puName(id.Value) // method names are case-insensitive
	switch name {
	case "assertEquals", "assertNotEquals":
		if args != nil {
			checkEqualsArgs(ctx, name, args.Args)
		}
	case "assertFileNotExists", "assertDirectoryNotExists":
		if ver < 91 { // E2
			return
		}
		repl := renamedAssertions[name]
		span := id.Span()
		ctx.Report(span, name+"() was deprecated by PHPUnit 9.1; call "+repl+"() instead.", analysis.Fix{
			Title: "Rename to " + repl + "()",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{{Span: span, NewText: repl}}
			},
		})
	}
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
