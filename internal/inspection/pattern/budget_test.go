package pattern

import (
	"strings"
	"testing"
	"time"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/testing/testbudget"
)

// Patterns and globs come from the analyzed project: however they are
// written, matching stays bounded and says when it gave up.
func TestHostilePatterns(t *testing.T) {
	r, err := Compile(Definition{
		ID: "R", Message: "m",
		Match: Patterns{`$_; \a(); $_; $_; \a(); $_; \a(); $_; \a(); $_; \b();`},
		Paths: []string{strings.Repeat("**/", 60) + "x.php"},
	})
	if err != nil {
		t.Fatal(err)
	}
	seq := r.(*semanticSeqRule).rule.patterns[0].stmts
	if len(seq) != 9 { // the leading $_; dropped, the doubled one merged
		t.Fatalf("statements: %d", len(seq))
	}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{ProjectPath: func(string) string { return strings.Repeat("d/", 50) + "x.php" }})
	if err != nil {
		t.Fatal(err)
	}
	src := "<?php\nfunction f() {\n" + strings.Repeat("    a();\n", 400) + "}\n"
	start := testbudget.CPU()
	res := e.AnalyzeResult(syntax.Parse("x.php", []byte(src), syntax.Options{}))
	if d := testbudget.CPU() - start; d > testbudget.Of(2*time.Second) {
		t.Fatalf("took %s", d)
	}
	if len(res.Findings) != 0 || len(res.Incomplete) == 0 || res.Incomplete[0].Reason != "query-budget" {
		t.Fatalf("result: %+v", res)
	}
	if matchGlob(strings.Repeat("**/", 60)+"y.php", strings.Repeat("d/", 50)+"x.php") {
		t.Fatal("glob matched")
	}
}
