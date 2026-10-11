package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewSqliteForeignKeyProbe dispatches on method calls only, as the rule package does.
type reviewSqliteForeignKeyProbe struct{ nativeCheckProbe }

func (reviewSqliteForeignKeyProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewSqliteForeignKeySettingInsideTransaction(t *testing.T) {
	const (
		lite   = "function f($ok, $sql, $m){$db=new SQLite3(':memory:');"
		pdo    = "function f($ok, $sql){$db=new PDO('sqlite::memory:');"
		pragma = "$db->exec('PRAGMA foreign_keys=ON');}"
	)
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"open transaction", lite + "$db->exec('BEGIN');" + pragma, 1},
		{"statements keep it open", lite + "$db->exec('BEGIN');$db->exec('INSERT INTO t VALUES (1)');$db->query('SELECT 1');$db->lastInsertRowID();" + pragma, 1},
		{"conditional statement keeps it open", lite + "$db->exec('BEGIN');if($ok){$db->exec('INSERT INTO t VALUES (1)');}" + pragma, 1},
		{"opened by query", lite + "$db->query('BEGIN');" + pragma, 1},
		{"opened by querySingle", lite + "$db->querySingle('BEGIN IMMEDIATE');" + pragma, 1},
		{"opened in guard", lite + "if(!$db->exec('BEGIN')){return;}" + pragma, 1},
		{"reopened after branches", lite + "$db->exec('BEGIN');if($ok){$db->exec('COMMIT');}$db->exec('BEGIN');" + pragma, 1},
		{"reopened after unknown sql", lite + "$db->exec($sql);$db->exec('BEGIN');" + pragma, 1},
		{"harmless prepare", lite + "$db->exec('BEGIN');$s=$db->prepare('SELECT 1');" + pragma, 1},
		{"named argument", lite + "$db->exec(query: 'BEGIN');$db->exec(query: 'PRAGMA foreign_keys=ON');}", 1},
		{"pdo open transaction", pdo + "$db->beginTransaction();" + pragma, 1},
		{"pdo named argument", pdo + "$db->exec(statement: 'BEGIN');$db->exec(statement: 'PRAGMA foreign_keys=ON');}", 1},
		{"pdo opened by query", pdo + "$db->query('BEGIN');" + pragma, 1},
		{"rolled back on a returning branch", lite + "$db->exec('BEGIN');if($ok){$ok=2;$db->exec('ROLLBACK');return;}" + pragma, 1},
		{"rolled back on a throwing branch", lite + "$db->exec('BEGIN');if($ok){$db->exec('ROLLBACK');throw new Exception('x');}" + pragma, 1},
		{"rolled back on an exiting branch", lite + "$db->exec('BEGIN');if($ok){$db->exec('ROLLBACK');$ok=2;exit(1);}" + pragma, 1},
		{"pdo rolled back on a returning branch", pdo + "$db->beginTransaction();if($ok){$db->rollBack();return;}" + pragma, 1},

		{"committed in both branches", lite + "$db->exec('BEGIN');if($ok){$db->exec('COMMIT');}else{$db->exec('ROLLBACK');}" + pragma, 0},
		{"committed in try and catch", lite + "$db->exec('BEGIN');try{$db->exec('CREATE TABLE t (a)');$db->exec('COMMIT');}catch(Exception $e){$db->exec('ROLLBACK');}" + pragma, 0},
		{"committed in one branch", lite + "$db->exec('BEGIN');if($ok){$db->exec('COMMIT');}" + pragma, 0},
		{"rolled back on failure", lite + "$db->exec('BEGIN');$db->exec('INSERT INTO t VALUES (1)') or $db->exec('ROLLBACK');" + pragma, 0},
		{"early return then commit", lite + "$db->exec('BEGIN');if(!$db->exec('INSERT INTO t VALUES (1)')){$db->exec('ROLLBACK');return;}$db->exec('COMMIT');" + pragma, 0},
		{"opened in one branch", lite + "if($ok){$db->exec('BEGIN');}" + pragma, 0},
		{"committed by query", lite + "$db->exec('BEGIN');$db->query('COMMIT');" + pragma, 0},
		{"committed by querySingle", lite + "$db->exec('BEGIN');$db->querySingle('END');" + pragma, 0},
		{"multi-statement query", lite + "$db->exec('BEGIN');$db->query('SELECT 1; SELECT 2');" + pragma, 0},
		{"trailing semicolon query", lite + "$db->exec('BEGIN');$db->query('SELECT 1;');" + pragma, 1},
		{"unknown query sql", lite + "$db->exec('BEGIN');$db->query($sql);" + pragma, 0},
		{"unparsable sql", lite + "$db->exec('BEGIN');$db->exec('SELECT $body$');" + pragma, 0},
		{"prepared commit", lite + "$db->exec('BEGIN');$s=$db->prepare('COMMIT');$s->execute();" + pragma, 0},
		{"prepared begin", lite + "$s=$db->prepare('BEGIN');$db->exec('BEGIN');" + pragma, 0},
		{"prepared unknown", lite + "$db->exec('BEGIN');$s=$db->prepare($sql);" + pragma, 0},
		{"closed and reopened", lite + "$db->exec('BEGIN');$db->close();$db->open(':memory:');" + pragma, 0},
		{"dynamic method", lite + "$db->exec('BEGIN');$db->$m();" + pragma, 0},
		{"user method", "class Db extends SQLite3{function finish(){$this->exec('COMMIT');}} function f(){$db=new Db(':memory:');$db->exec('BEGIN');$db->finish();" + pragma, 0},
		{"unknown pragma sql", lite + "$db->exec('BEGIN');$db->exec($sql);}", 0},
		{"pdo committed by method in branch", pdo + "$db->beginTransaction();if($ok){$db->commit();}else{$db->rollBack();}" + pragma, 0},
		{"pdo opened in one branch", pdo + "if($ok){$db->beginTransaction();}" + pragma, 0},
		{"pdo committed by query", pdo + "$db->beginTransaction();$db->query('COMMIT');" + pragma, 0},
		{"pdo wrong argument name", pdo + "$db->exec(query: 'BEGIN');" + pragma, 0},
		{"rolled back before continue", lite + "$db->exec('BEGIN');foreach([1] as $_){if($ok){$db->exec('ROLLBACK');continue;}}" + pragma, 0},
		{"rolled back before a caught throw", lite + "$db->exec('BEGIN');try{if($ok){$db->exec('ROLLBACK');throw new Exception('x');}}catch(Exception $e){}" + pragma, 0},
		{"rolled back in a nested branch", lite + "$db->exec('BEGIN');if($ok){if($sql){$db->exec('ROLLBACK');}return;}" + pragma, 0},
		{"rolled back before a conditional return", lite + "$db->exec('BEGIN');if($ok){$db->exec('ROLLBACK');if($sql){return;}}" + pragma, 0},
		{"short-circuit rollback in the same block", lite + "$db->exec('BEGIN');$ok && $db->exec('ROLLBACK');$db->exec('PRAGMA foreign_keys=ON');return;}", 0},
		{"committed in a returning closure", lite + "$db->exec('BEGIN');$g=function()use($db){$db->exec('COMMIT');return;};$g();" + pragma, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewSqliteForeignKeyProbe{nativeCheckProbe{"SqliteForeignKeySettingInsideTransaction", CheckSqliteForeignKeySettingInsideTransaction}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"SqliteForeignKeySettingInsideTransaction"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
