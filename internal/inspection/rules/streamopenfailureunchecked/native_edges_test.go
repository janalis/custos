package streamopenfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestIndependentNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"$h=fopen('file','rb');if($h!==false){fread($h,4);fclose($h);}", 0},
		{"$h=fopen('file','rb');if(!$h){throw new Exception('x');}fread($h,4);", 0},
		{"if($h=fopen('file','rb')){fgets($h);}", 0},
		{"$h=fopen('file','rb');if($h){fread($h,4);}", 0},
		{"$h=fopen('file','rb');if(!is_resource($h)){return;}fread($h,4);", 0},
		{"$h=fopen('file','rb') or die('x');fread($h,4);", 0},
		{"$h=fopen('file','rb');if(!$h){echo 'x';}fread($h,4);", 1},
		{"$h=fopen('file','rb');if(!$h){return;}$h=fopen('other','rb');fread($h,4);", 1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"StreamOpenFailureUnchecked"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
