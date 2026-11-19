package pattern

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

// BenchmarkRules runs twenty pattern rules over a file of calls,
// comparisons and instantiations, most of which they do not match.
func BenchmarkRules(b *testing.B) {
	patterns := []string{
		`\dump(...)`, `\dd(...)`, `$X == null`, `$X != null`, `new \DateTime(...)`, `\in_array($N, $H)`,
		`\count($X) == 0`, `\strlen($X) == 0`, `$this->app->make($X)`, `\json_encode($X)`, `\sprintf($F, ...$A)`,
		`\str_replace($A, $B, $C)`, `\is_null($X)`, `$X === $X`, `\array_merge(...)`, `\call_user_func(...)`,
		`\func_get_args()`, `\array_key_exists($K, $A)`, `$X->save()`, `\App\helper($X)`,
	}
	var rules []analysis.Rule
	for i, p := range patterns {
		r, err := Compile(Definition{ID: fmt.Sprintf("R%d", i), Match: Patterns{p}, Message: "m"})
		if err != nil {
			b.Fatal(err)
		}
		rules = append(rules, r)
	}
	e, err := analysis.NewEngine(rules, analysis.Config{})
	if err != nil {
		b.Fatal(err)
	}
	var src strings.Builder
	src.WriteString("<?php\nnamespace App;\nfunction f($a, $b) {\n")
	for i := range 200 {
		fmt.Fprintf(&src, "    $x%d = strtolower($a) . sprintf('%%d', %d) . $b->name($a == null ? 1 : 2);\n    $o%d = new \\stdClass();\n", i, i, i)
	}
	src.WriteString("}\n")
	f := syntax.Parse("bench.php", []byte(src.String()), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(f)
	}
}
