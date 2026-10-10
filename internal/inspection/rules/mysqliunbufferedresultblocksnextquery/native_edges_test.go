package mysqliunbufferedresultblocksnextquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliUnbufferedResultBlocksNextQuery"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"function empty(mysqli $db){$db->query(\"SELECT 1\");}", 0},
		{"$db=new mysqli();echo 1;$db->query(\"SELECT 1\");", 0},
		{"$db=new mysqli();$x=4;$db->query(\"SELECT 1\");", 0},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$other=new mysqli();$other->query(\"SELECT 1\");", 0},
		{"$db=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);function nested($db){$db->query(\"SELECT 1\");}", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestOperationalAdditionalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliUnbufferedResultBlocksNextQuery"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$db=new mysqli();$object->r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$db->query(\"SELECT 1\");", 0},
		{"$db=new mysqli();$other=new mysqli();$r=$db->query(\"SELECT 1\",MYSQLI_USE_RESULT);$other->query(\"SELECT 1\");", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
