package sessionmutationafterclose

import (
	"os"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestIndependentNativeFalsePositiveFixture(t *testing.T) {
	src, err := os.ReadFile("../../../../testdata/rules/SessionMutationAfterClose/false-positives.php")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SessionMutationAfterClose"}})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("negative.php", src, syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("unexpected findings: %+v", got)
	}
}
