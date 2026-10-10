package execoutputarrayaccumulates

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ExecOutputArrayAccumulates"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"exec(\"x\",$object->out);", 0},
		{"exec(\"x\",$out);", 0},
		{"echo 1;echo 2;exec(\"x\",$out);", 0},
		{"echo 1;$out=[];exec(\"x\",$out);", 0},
		{"$out=[];exec(\"x\",$other);exec(\"y\",$out);", 0},
		{"echo 1;exec(\"x\",$out);exec(\"y\",$out);", 0},
		{"$x=1;exec(\"x\",$out);exec(\"y\",$out);", 0},
		{"$other=[];exec(\"x\",$out);exec(\"y\",$out);", 0},
		{"$out=[1];exec(\"x\",$out);exec(\"y\",$out);", 0},
		{"$out=[];exec(\"x\",$out);$last=exec(\"y\",$out);", 0},
		{"$out=[];exec(\"x\",$out);exec(\"y\",$out);", 0},
		{"$out=[];exec(\"x\",$out);exec(\"y\",$out);foreach($out as $line){echo $line;}", 1},
		{"$out=[];exec(\"x\",$out);exec(\"y\",$out);return;", 0},
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
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ExecOutputArrayAccumulates"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"$out=&$alias;exec(\"x\",$out);exec(\"y\",$out);echo count($out);", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
