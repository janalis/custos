package formencodingusedforrfc3986signature

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestEdges(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
		php    string
	}{
		{"other(); $x->other(); echo $x[0];", 0, ""},
		{"function signRfc3986($k){$q=http_build_query([\"x\"=>\"a b\"]);return hash_hmac(\"sha256\",$q,$k);}", 1, ""},
		{"function otherSignature($k){$q=http_build_query([\"x\"=>\"a b\"]);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=http_build_query([\"x\"=>\"plain\"]);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=http_build_query([\"x\"=>$value]);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=http_build_query($data);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=json_encode([\"x\"=>\"a b\"]);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=http_build_query([\"x\"=>\"a b\"],\"\",\"&\",$encoding);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"function signRfc3986($k){$q=http_build_query([\"x\"=>\"a b\"],\"\",\"&\",PHP_QUERY_RFC3986);return hash_hmac(\"sha256\",$q,$k);}", 0, ""},
		{"hash_hmac(\"sha256\",$q,$k);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"FormEncodingUsedForRfc3986Signature"}, PHP: phpversion.MustParse(tc.php)}
			cfg.Rules = map[string]analysis.RuleConfig{"FormEncodingUsedForRfc3986Signature": {Options: map[string]any{"signatureFunctions": []string{"signRfc3986"}}}}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("edge.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
