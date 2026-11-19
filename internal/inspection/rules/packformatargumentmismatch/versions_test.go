package packformatargumentmismatch

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

func TestVersion161(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PackFormatArgumentMismatch"}, PHP: phpversion.MustParse("5.3")})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("test.php", []byte("<?php pack('J2',1);"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("unsupported contract: %+v", got)
	}
}
