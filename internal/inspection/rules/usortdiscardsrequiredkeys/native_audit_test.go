package usortdiscardsrequiredkeys

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestAuditKeyWritesAndBuiltinFix(t *testing.T) {
	for _, tc := range []struct {
		src           string
		want          int
		fixedContains string
	}{
		{"$a=['k'=>1];usort($a,$cmp);$a['k']=9;", 0, ""},
		{"$a=['k'=>1];usort($a,$cmp);unset($a['k']);", 0, ""},
		{"$a=['k'=>1];usort($a,$cmp);$exists=isset($a['k']);", 0, ""},
		{"$a=['k'=>1];usort($a,$cmp);$blank=empty($a['k']);", 0, ""},
		{"$a=['k'=>1];usort($a,$cmp);echo $a['k']??'fallback';", 0, ""},
		{"namespace Audit;function uasort(&$a,$c){} $a=['k'=>1];usort($a,$cmp);echo $a['k'];", 1, "\\uasort($a,$cmp)"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UsortDiscardsRequiredKeys"}})
			if err != nil {
				t.Fatal(err)
			}
			src := []byte("<?php " + tc.src)
			got := e.Analyze(syntax.Parse("audit.php", src, syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
			if tc.fixedContains != "" {
				edits := got[0].Fixes[0].Edits()
				edit := edits[0]
				fixed := string(src[:edit.Span.Start]) + edit.NewText + string(src[edit.Span.End:])
				if !strings.Contains(fixed, tc.fixedContains) {
					t.Fatalf("wrong target in %s", fixed)
				}
			}
		})
	}
}
