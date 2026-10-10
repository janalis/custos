package untrustedxmlwriterrawcontent

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"function f(XMLWriter $w){$w->writeRaw('safe');$w->writeRaw($a['x']);$w->writeRaw($a->value['x']);}", 0},
		{"function f(XMLWriter $w){$v=$_GET['x'];$w->writeRaw($v);}", 1},
		{"function f(XMLWriter $w){$_POST['x']='safe';$w->writeRaw($_POST['x']);}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedXmlWriterRawContent"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("edges.php", []byte("<?php "+tc.src), syntax.Options{})
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
