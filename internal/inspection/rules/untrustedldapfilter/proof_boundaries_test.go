package untrustedldapfilter

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestRuleContract(t *testing.T) {
	r := New()
	if r.ID() != "UntrustedLdapFilter" || len(r.Kinds()) == 0 {
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
	}{{"strlen(\"x\");", 0}, {"ldap_search($ldap,$base,$unknown);", 0}, {"ldap_search($ldap,$base,\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"x\".\"safe\");", 0}, {"ldap_search($l,$b,$o->p[\"user\"]);", 0}, {"ldap_search($l,$b,$custom[\"user\"]);", 0}} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedLdapFilter"}})
		if err != nil {
			t.Fatal(err)
		}
		findings := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
		if len(findings) != tc.want {
			t.Errorf("%s: got %d findings, want %d", tc.src, len(findings), tc.want)
		}
	}
}
