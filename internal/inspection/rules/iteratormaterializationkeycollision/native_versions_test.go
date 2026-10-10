package iteratormaterializationkeycollision

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	v "custos/internal/php/version"
)

func TestBeforeLanguageFeature(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"IteratorMaterializationKeyCollision"}, PHP: v.PHP54})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("old.php", []byte("<?php function rows(){yield 'k'=>1;yield 'k'=>2;} iterator_to_array(rows());"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("old version produced findings: %+v", got)
	}
}
