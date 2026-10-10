package readloopprocesseseoffailure

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ReadLoopProcessesEofFailure"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$h=tmpfile();while(!feof($h)){strlen(fgets($h));}", 0},
		{"declare(strict_types=1);while(!is_null($h)){strlen(fgets($h));}", 0},
		{"declare(strict_types=1);while(!feof($o->h)){strlen(fgets($o->h));}", 0},
		{"declare(strict_types=1);$h=tmpfile();while(!feof($h))echo fgets($h);", 0},
		{"declare(strict_types=1);$h=tmpfile();while(!feof($h)){$o->line=fgets($h);strlen($o->line);}", 0},
		{"declare(strict_types=1);$h=tmpfile();while(!feof($h)){strlen(\"line\");$x=1;}", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
