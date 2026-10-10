package libxmlerrorbuffernevercleared

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// TestAuditRegressions exercises independent valid programs and comment-preservation boundaries.
func TestAuditRegressions(t *testing.T) {
	for _, tc := range []struct {
		src         string
		want, fixes int
	}{
		{"libxml_use_internal_errors(true);$d=new DOMDocument();for($i=0;$i<3;libxml_clear_errors(),$i++){$d->loadXML($xml);}", 0, -1},
		{"libxml_use_internal_errors(true);$d=new DOMDocument();foreach([] as $xml){$d->loadXML($xml);}", 0, -1},
		{"libxml_use_internal_errors(true);for(;false;){simplexml_load_string($xml);}", 0, -1},
		{"libxml_use_internal_errors(true);while(false){simplexml_load_string($xml);}", 0, -1},
		{"libxml_use_internal_errors(true);while($more){simplexml_load_string(clearAndRead());}", 0, -1},
		{"libxml_use_internal_errors(true);$d=new DOMDocument(clearAndVersion());while($more){$d->loadXML($xml);}", 0, -1},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("bad regression source: %+v", f.Errors)
			}
			got := e.Analyze(f)
			if len(got) != tc.want {
				t.Fatalf("got %d findings want %d: %+v", len(got), tc.want, got)
			}
			if tc.fixes >= 0 {
				count := 0
				for _, d := range got {
					count += len(d.Fixes)
				}
				if count != tc.fixes {
					t.Fatalf("got %d fixes want %d", count, tc.fixes)
				}
			}
		})
	}
}

func BenchmarkNativeStateProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(`<?php libxml_use_internal_errors(true);while($more){simplexml_load_string($xml);}`), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(f)
	}
}
