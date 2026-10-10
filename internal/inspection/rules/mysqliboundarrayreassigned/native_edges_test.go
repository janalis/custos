package mysqliboundarrayreassigned

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliBoundArrayReassigned"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"function empty(mysqli_stmt $s){$s->execute();}", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");echo 1;echo 2;$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");echo 1;$row=[];$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$row=[];$row=[];$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$s->bind_param(\"i\",...$ids);$row=[];$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$s->bind_param(\"i\",$id);$row=[];$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$row=[];$other=[];$s->bind_param(\"i\",$other[0]);$row=[];$s->execute();", 0},
		{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$row=[];$s->bind_param(\"i\",$row[0]);$row=[];$other=$db->prepare(\"SELECT ?\");$other->execute();", 0},
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
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliBoundArrayReassigned"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"$db=new mysqli();$s=$db->prepare(\"SELECT ?\");$row=[];$s->bind_param(\"i\",$id,...$others);$row=[];$s->execute();", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
