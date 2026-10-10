package sessionregenerationfailureunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "SessionRegenerationFailureUnchecked" || len(r.Kinds()) == 0 {
		t.Fatal("invalid rule contract")
	}
	if s, ok := r.(interface{ Semantic() }); ok {
		s.Semantic()
	}
	if f, ok := r.(interface{ Flow() }); ok {
		f.Flow()
	}
}

func TestProofBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{{"strlen(\"x\");", 0}, {"session_regenerate_id();", 0}, {"echo session_regenerate_id();", 0}, {"session_regenerate_id();echo 1;", 0}, {"session_regenerate_id();$x=true;", 0}, {"session_regenerate_id();$_POST[\"authenticated\"]=true;", 0}, {"session_regenerate_id();$_SESSION[$unknown]=true;", 0}, {"session_regenerate_id();$_SESSION[\"other\"]=true;", 0}, {"session_regenerate_id();$_SESSION[\"authenticated\"]=false;", 0}, {"if($x) session_regenerate_id();", 0}, {"session_regenerate_id();strlen(\"x\");", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SessionRegenerationFailureUnchecked"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
