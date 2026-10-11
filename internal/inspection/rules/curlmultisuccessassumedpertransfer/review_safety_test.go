package curlmultisuccessassumedpertransfer

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"CurlMultiSuccessAssumedPerTransfer"}})
	if err != nil {
		t.Fatal(err)
	}
	const setup = "$h=curl_multi_init();curl_multi_add_handle($h,curl_init());"
	const check = "if(curl_multi_exec($h,$running)===CURLM_OK && !$running){return true;}"
	for _, tc := range []struct {
		name, between string
		findings      int
	}{
		{"plain", "", 1},
		{"lifecycle builtins", "curl_multi_setopt($h,CURLMOPT_PIPELINING,0);curl_multi_select($h);curl_multi_errno($h);", 1},
		{"unrelated helper", "helper($other);$o->read($other);", 1},
		{"reassigned handle", "$h=curl_multi_init();curl_multi_add_handle($h,curl_init());", 1},
		{"user helper", "if(failed($h)>0){return false;}", 0},
		{"method helper", "$this->drain($h);", 0},
		{"static helper", "Transfers::drain($h);", 0},
		{"constructor", "new Drain($h);", 0},
		{"info read", "curl_multi_info_read($h);", 0},
		{"removed transfer", "curl_multi_remove_handle($h,$c);", 0},
		{"closed handle", "curl_multi_close($h);", 0},
		{"spread argument", "curl_multi_select(...$h);", 0},
		{"alias", "$copy=$h;drain($copy);", 0},
		{"reference alias", "$copy=&$h;", 0},
		{"reference target", "$h=&$other;", 0},
		{"stored copy", "$this->handle=$h;$this->drain();", 0},
		{"array member", "$all=[$h];", 0},
		{"closure capture", "$f=function() use ($h){return curl_multi_info_read($h);};$f();", 0},
		{"arrow capture", "$f=fn()=>curl_multi_info_read($h);$f();", 0},
		{"variable function", "$f($h);", 0},
		{"variable variable", "$n='h';drain($$n);", 0},
		{"compact", "drain(compact('h'));", 0},
		{"defined variables", "drain(get_defined_vars());", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, source := range []string{
				"<?php " + setup + tc.between + check,
				"<?php function f($other,$o,$c){" + setup + tc.between + check + "}",
			} {
				f := syntax.ParseBest("review.php", []byte(source), syntax.Options{})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe: %v", f.Errors)
				}
				for _, g := range e.Analyze(f) {
					if g.Rule != "CurlMultiSuccessAssumedPerTransfer" {
						t.Fatalf("rule crashed: %+v", g)
					}
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("%s: got %v, want %d", source, got, tc.findings)
				}
			}
		})
	}
	// A handle that is not held in a plain variable cannot be followed.
	f := syntax.ParseBest("review.php", []byte("<?php if(curl_multi_exec(curl_multi_init(),$running)===CURLM_OK && !$running){return true;}"), syntax.Options{})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("got %v, want none", got)
	}
}
