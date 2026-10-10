package nonblockingemptyreadtreatedaseof

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NonBlockingEmptyReadTreatedAsEof"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$h=tmpfile();stream_set_blocking($h,false);if(\"\"===fread($h,1)){fclose($h);}", 1},
		{"$h=tmpfile();if(fread($h,1)===\"x\"){fclose($h);}", 0},
		{"$h=tmpfile();if(fread($h,1)===\"\")echo 1;", 0},
		{"$h=tmpfile();if(fread($h,1)===\"\"){return;}", 0},
		{"$h=tmpfile();if(fread($h,1)===\"\"){fclose($other);}", 0},
		{"$h=tmpfile();if(fread($h,1)===\"\"){fclose($h);}", 0},
		{"$h=tmpfile();stream_set_blocking($h,$unknown);if(fread($h,1)===\"\"){fclose($h);}", 0},
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
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NonBlockingEmptyReadTreatedAsEof"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"$h=tmpfile();if(fread($h,1)===\"\"){$x=4;}", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
