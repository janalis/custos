package pgreturnedrowsusedasaffectedrows

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PgReturnedRowsUsedAsAffectedRows"}})
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
		{"\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_num_rows($r); }\n", 1},
		{"\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_affected_rows($r); }\n", 0},
		{"$r=pg_query($db,'UPDATE x SET y=1 RETURNING y');if($r!==false){pg_num_rows($r);}", 0},
		{"$r=pg_query($db,\"UPDATE x SET y='RETURNING'\");if($r!==false){pg_num_rows($r);}", 1},
		{"$r=pg_query($db,'SELECT 1');if($r!==false){pg_num_rows($r);}", 0},
		{"$r=pg_query($db,'UPDATE x SET y=1');pg_num_rows($r);", 0},
		{"if(false){\n$r = pg_query($db, 'UPDATE jobs SET done = true'); if ($r !== false) { echo pg_num_rows($r); }\n}", 0},
		{"$r=pg_query('UPDATE x SET y=1');if($r!==false){pg_num_rows($r);}", 1},
		{"$r=pg_query($d,$sql);if($r!==false){pg_num_rows($r);}", 0},
		{"$r=pg_query($d,'SELECT $body$');if($r!==false){pg_num_rows($r);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
