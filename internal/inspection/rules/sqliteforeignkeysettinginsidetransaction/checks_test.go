package sqliteforeignkeysettinginsidetransaction

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SqliteForeignKeySettingInsideTransaction"}})
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
		{"\n$db = new SQLite3(':memory:'); $db->exec('BEGIN; PRAGMA foreign_keys=ON;');\n", 1},
		{"\n$db = new SQLite3(':memory:'); $db->exec('PRAGMA foreign_keys=ON; BEGIN;');\n", 0},
		{"$d=new SQLite3(':memory:');$d->exec('BEGIN');$d->exec('PRAGMA foreign_keys=ON');", 1},
		{"$d=new SQLite3(':memory:');$d->exec('BEGIN;COMMIT;PRAGMA foreign_keys=ON');", 0},
		{"$d=new PDO('sqlite::memory:');$d->beginTransaction();$d->exec('PRAGMA foreign_keys=ON');", 1},
		{"$d=new PDO('mysql:host=x');$d->beginTransaction();$d->exec('PRAGMA foreign_keys=ON');", 0},
		{"if(false){\n$db = new SQLite3(':memory:'); $db->exec('BEGIN; PRAGMA foreign_keys=ON;');\n}", 0},
		{"$d=new PDO('sqlite::memory:');$d->beginTransaction();$d->commit();$d->exec('PRAGMA foreign_keys=ON');", 0},
		{"$d=new SQLite3(':memory:');$d->exec($sql);$d->exec('PRAGMA foreign_keys=ON');", 0},
		{"$d=new SQLite3(':memory:');$d->exec('SELECT $body$');$d->exec('PRAGMA foreign_keys=ON');", 0},
		{"$d=new SQLite3(':memory:');$d->exec($sql);", 0},
		{"$d=new SQLite3(':memory:');$d->exec('SELECT $body$');", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
