package passwordcomparedwithfreshhash

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestPreservePasswordFixEvaluation(t *testing.T) {
	for _, src := range []string{
		"password_hash($password, chooseAlgorithm()) === $stored;",
		"loadHash() === password_hash($password, PASSWORD_DEFAULT);",
		"password_hash($password, PASSWORD_DEFAULT, options()) === $stored;",
	} {
		t.Run(src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PasswordComparedWithFreshHash"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{}))
			if len(got) != 1 {
				t.Fatalf("got %d findings, want1: %+v", len(got), got)
			}
			if len(got[0].Fixes) != 0 {
				t.Fatal("fix must preserve evaluation and side effects")
			}
		})
	}
}
