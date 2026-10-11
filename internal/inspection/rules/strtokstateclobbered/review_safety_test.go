package strtokstateclobbered

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const nested = "<?php function f($csv) { $row = strtok($csv, \"\\n\"); while ($row !== false) { $cell = strtok($row, ',');" +
		" while ($cell !== false) { echo $cell; $cell = strtok(','); } $row = strtok(\"\\n\"); } }"
	for _, tc := range []struct {
		name, source string
		findings     int
	}{
		{"nested tokenizer", nested, 1},
		{"statement between", "<?php $a = strtok($x, ','); $b = strtok($y, ';'); echo $a, $b; $c = strtok(',');", 1},
		{"return", "<?php function f($x, $y) { strtok($x, ','); strtok($y, ';'); return strtok(','); }", 1},
		{"alternative syntax", "<?php strtok($x, ','); while ($q): strtok($y, ';'); strtok(','); endwhile;", 1},
		{"named token", "<?php strtok($x, ','); strtok($y, ';'); strtok(token: ';');", 0},
		{"named token clobbered", "<?php strtok($x, ','); strtok($y, ';'); strtok(token: ',');", 1},
		{"leading token", "<?php $s = strtok($u, ':'); $n = strtok($f, '.'); $e = strtok('.');", 0},
		{"interpolated delimiter", "<?php strtok($x, ','); strtok($y, \"$d\"); strtok(\"$d\");", 1},
		{"continuation before second start", "<?php strtok($x, ','); if ($q) { strtok(','); } strtok($y, ';'); strtok(',');", 0},
		{"continuation header before second start", "<?php strtok($x, ','); strtok($y, ';'); while (strtok(';')) { strtok(','); }", 1},
		{"continuation header between starts", "<?php strtok($x, ','); while (strtok(',')) { strtok($y, ';'); strtok(','); }", 0},
		{"elseif branch", "<?php strtok($x, ','); strtok($y, ';'); if ($q) {} elseif ($r) { strtok(','); }", 0},
		{"static call between", "<?php strtok($x, ','); strtok($y, ';'); A::b(); strtok(',');", 0},
		{"new between", "<?php strtok($x, ','); strtok($y, ';'); new A; strtok(',');", 0},
		{"include between", "<?php strtok($x, ','); strtok($y, ';'); include 'a.php'; strtok(',');", 0},
		{"arrow function between", "<?php strtok($x, ','); strtok($y, ';'); $f = fn() => 1; strtok(',');", 0},
		{"user function", "<?php namespace N; function strtok($a, $b = null) {} strtok($x, ','); strtok($y, ';'); strtok(',');", 0},
		{"single start in function", "<?php function f($y) { strtok($y, ';'); return strtok(','); }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(tc.source), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
