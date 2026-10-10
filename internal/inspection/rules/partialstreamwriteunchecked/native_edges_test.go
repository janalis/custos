package partialstreamwriteunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"if(false===fwrite($h,$s)){return false;} return true;", 0},
		{"function send($h,$s){if(fwrite($h,$s)===false){return false;}return true;}", 1},
		{"function send($h,$s){if(fwrite($h,$s)===true){return false;}return true;}", 0},
		{"function send($h,$s){if(fwrite($h,$s)===0){return false;}return true;}", 0},
		{"$ok=fwrite($h,$s)===false;", 0},
		{"if(fwrite($h,$s)===false){return false;}", 0},
		{"if(fwrite($h,$s)===false){}else{} return true;", 0},
		{"if(fwrite($h,$s)===false){} echo 'x';", 0},
		{"if(fwrite($h,$s)===false){} return false;", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"PartialStreamWriteUnchecked"}})
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
