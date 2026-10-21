package mysqlimultiqueryresultsnotdrained

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliMultiQueryResultsNotDrained"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"\n$db = new mysqli(); if ($db->multi_query('SELECT 7; SELECT 11')) { $db->query('SELECT 14'); }\n", 1},
		{"\n$db = new mysqli(); $db->query('SELECT 7'); $db->query('SELECT 14');\n", 0},
		{"function f(mysqli $db){if($db->multi_query('SELECT 1;')){$db->query('SELECT 2');}}", 0},
		{"function f(mysqli $db){if($db->multi_query('SELECT 1;SELECT 2')){while($db->more_results()){$db->next_result();}$db->query('SELECT 3');}}", 0},
		{"function f(mysqli $db){if($db->multi_query('SELECT 1;SELECT 2;SELECT 3')){$db->next_result();$db->query('SELECT 4');}}", 1},
		{"if(false){\n$db = new mysqli(); if ($db->multi_query('SELECT 7; SELECT 11')) { $db->query('SELECT 14'); }\n}", 0},
		{"function f(mysqli $d){$d->multi_query('SELECT 1;SELECT 2');$d->query('SELECT 3');}", 0},
		{"function f(mysqli $d){if($d->multi_query($sql)){$d->query('SELECT 3');}}", 0},
		{"function f(mysqli $d){if($d->multi_query('SELECT $body$')){$d->query('SELECT 3');}}", 0},
		{"function f(mysqli $d){if($d->multi_query('SELECT 1;SELECT 2')){$pending=$d->more_results();$d->query('SELECT 3');}}", 1},
		{`function f(mysqli $d){if(!$d->multi_query('SELECT 1;SELECT 2')){return;}$d->query('SELECT 3');}`, 1},
		{`function f(mysqli $d){if($d->multi_query('SELECT 1;SELECT 2')===true){$d->query('SELECT 3');}}`, 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
