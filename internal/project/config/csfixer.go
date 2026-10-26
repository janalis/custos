package config

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// csFixerFiles are the PHP-CS-Fixer configuration files, in the order the
// tool itself looks for them.
var csFixerFiles = []string{".php-cs-fixer.php", ".php-cs-fixer.dist.php"}

// csFixerYodaSets are the rule sets that enable `yoda_style`.
var csFixerYodaSets = map[string]bool{
	"@symfony": true, "@symfony:risky": true, "@phpcsfixer": true, "@phpcsfixer:risky": true,
}

// csFixerStyleRead returns the comparison style the project's PHP-CS-Fixer
// configuration enforces: its `yoda_style` rule when the file names it
// (`false`, or a list setting both `equal` and `identical` to `false`,
// means regular order; anything else yoda), else yoda when it enables a
// rule set that includes that rule (`@Symfony`, `@PhpCsFixer`). ok is false
// without such a file or when it says nothing about the style.
func csFixerStyleRead(read func(string) ([]byte, error)) (analysis.ComparisonStyle, bool) {
	for _, name := range csFixerFiles {
		b, err := read(name)
		if err != nil {
			continue
		}
		f := syntax.ParseBest(name, b, syntax.Options{})
		explicit, set := "", false
		syntax.InspectFile(f, func(n syntax.Node) bool {
			item, ok := n.(*syntax.ArrayItem)
			if !ok {
				return true
			}
			switch key := strings.ToLower(csFixerString(item.Key)); {
			case key == "yoda_style":
				explicit = csFixerYoda(item.Value)
			case csFixerYodaSets[key]:
				set = set || csFixerConst(item.Value, "true")
			}
			return true
		})
		switch {
		case explicit == "regular":
			return analysis.StyleRegular, true
		case explicit == "yoda" || set:
			return analysis.StyleYoda, true
		}
		return analysis.StyleRegular, false
	}
	return analysis.StyleRegular, false
}

// csFixerYoda reads the configuration of the `yoda_style` rule: "regular"
// when it is off or keeps the regular order for both `==` and `===`,
// "yoda" otherwise.
func csFixerYoda(v syntax.Expr) string {
	if csFixerConst(v, "false") {
		return "regular"
	}
	if arr, ok := v.(*syntax.Array); ok {
		off := 0
		for _, it := range arr.Items {
			if it == nil {
				continue
			}
			switch csFixerString(it.Key) {
			case "equal", "identical":
				if csFixerConst(it.Value, "false") {
					off++
				}
			}
		}
		if off == 2 {
			return "regular"
		}
	}
	return "yoda"
}

// csFixerConst reports whether v is the constant name (`true`, `false`).
func csFixerConst(v syntax.Expr, name string) bool {
	c, ok := v.(*syntax.ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), name)
}

// csFixerString returns the value of a quoted string literal, "" for
// anything else.
func csFixerString(e syntax.Expr) string {
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 || lit.Raw[0] != '\'' && lit.Raw[0] != '"' {
		return ""
	}
	return lit.Raw[1 : len(lit.Raw)-1]
}
