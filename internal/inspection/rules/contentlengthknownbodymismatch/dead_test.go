package contentlengthknownbodymismatch

import (
	"os"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
)

func TestUnreachableCall(t *testing.T) {
	raw, err := os.ReadFile("../../../../testdata/rules/ContentLengthKnownBodyMismatch/basic.php")
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := conformance.ParseMarkup(raw)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.TrimPrefix(string(src), "<?php")
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ContentLengthKnownBodyMismatch"}})
	if err != nil {
		t.Fatal(err)
	}
	file := syntax.Parse("dead.php", []byte("<?php function unreachable(){return;"+body+"}"), syntax.Options{})
	if len(file.Errors) != 0 {
		t.Fatal(file.Errors)
	}
	if got := e.Analyze(file); len(got) != 0 {
		t.Fatalf("dead call: %+v", got)
	}
}
