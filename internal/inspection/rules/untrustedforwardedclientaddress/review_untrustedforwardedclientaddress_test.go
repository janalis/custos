package untrustedforwardedclientaddress

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewUntrustedForwardedClientAddress(t *testing.T) {
	const decl = "/** @custos-protected */ function grant(){} "
	const use = "if($_SERVER['HTTP_X_FORWARDED_FOR']===$ip){grant();}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"unguarded", "function f($ip){" + use + "}", 1},
		{"unguarded at top level", use, 1},
		{"reversed operands", "function f($ip){if($ip==$_SERVER['HTTP_X_FORWARDED_FOR']){grant();}}", 1},
		{"other server key", "function f($ip){if($_SERVER['HTTP_CLIENT_IP']===$ip){grant();}}", 0},
		{"other array", "function f($ip,$h){if($h['HTTP_X_FORWARDED_FOR']===$ip){grant();}}", 0},
		{"dynamic array", "function f($ip){if(headers()['HTTP_X_FORWARDED_FOR']===$ip){grant();}}", 0},
		{"dynamic key", "function f($ip,$k){if($_SERVER[$k]===$ip){grant();}}", 0},

		{"enclosing peer check", "function f($ip){if($_SERVER['REMOTE_ADDR']==='10.0.0.1'){" + use + "}}", 0},
		{"enclosing builtin on peer", "function f($ip,$p){if(in_array($_SERVER['REMOTE_ADDR'],$p,true)){" + use + "}}", 0},
		{"enclosing project call", "function trusted(){return true;} function f($ip){if(trusted()){" + use + "}}", 0},
		{"enclosing unknown call", "function f($ip){if(isTrustedProxy()){" + use + "}}", 0},
		{"enclosing method call", "function f($ip,$r){if($r->isFromTrustedProxy()){" + use + "}}", 0},
		{"enclosing static call", "function f($ip){if(Proxy::trusted()){" + use + "}}", 0},
		{"enclosing else arm", "function f($ip){if(!isTrustedProxy()){deny();}else{" + use + "}}", 0},
		{"enclosing elseif condition", "function f($ip,$a){if($a){deny();}elseif(isTrustedProxy()){" + use + "}}", 0},
		{"outer enclosing check", "function f($ip,$a){if(isTrustedProxy()){if($a){" + use + "}}}", 0},
		{"enclosing plain flag", "function f($ip,$a){if($a){" + use + "}}", 1},
		{"enclosing builtin call", "function f($ip){if(is_string($ip)){" + use + "}}", 1},
		{"enclosing other server key", "function f($ip){if($_SERVER['HTTPS']==='on'){" + use + "}}", 1},
		{"enclosing loop", "function f($ip,$l){foreach($l as $_){" + use + "}}", 1},

		{"preceding guard", "function f($ip){if(!isTrustedProxy($_SERVER['REMOTE_ADDR'])){return;}" + use + "}", 0},
		{"preceding peer guard throws", "function f($ip){if($_SERVER['REMOTE_ADDR']!=='10.0.0.1'){throw new Exception();}" + use + "}", 0},
		{"earlier preceding guard", "function f($ip){if(!isTrustedProxy()){return;}$a=1;" + use + "}", 0},
		{"preceding guard in else", "function f($ip){if(isTrustedProxy()){$a=1;}else{return;}" + use + "}", 0},
		{"preceding guard in elseif", "function f($ip,$a){if($a){$b=1;}elseif(!isTrustedProxy()){return;}" + use + "}", 0},
		{"preceding guard of outer list", "function f($ip,$a){if(!isTrustedProxy()){return;}if($a){" + use + "}}", 0},
		{"preceding guard at top level", "if(!isTrustedProxy()){exit;}" + use, 0},
		{"preceding check falls through", "function f($ip){if(!isTrustedProxy()){log_it();}" + use + "}", 1},
		{"preceding elseif falls through", "function f($ip,$a){if($a){$b=1;}elseif(!isTrustedProxy()){$b=2;}else{$b=3;}" + use + "}", 1},
		{"preceding plain guard", "function f($ip,$a){if(!$a){return;}" + use + "}", 1},
		{"preceding builtin guard", "function f($ip){if(!is_string($ip)){return;}" + use + "}", 1},
		{"following guard", "function f($ip){" + use + "if(!isTrustedProxy()){return;}}", 1},
		{"guard outside closure", "function f($ip){if(!isTrustedProxy()){return;}return function()use($ip){" + use + "};}", 1},
		{"check outside closure", "function f($ip){if(isTrustedProxy()){return function()use($ip){" + use + "};}}", 1},
		{"guard before switch", "function f($ip,$a){if(!isTrustedProxy()){return;}switch($a){case 1:" + use + "}}", 0},
		{"guard inside case", "function f($ip,$a){switch($a){case 1:if(!isTrustedProxy()){break;}" + use + "}}", 0},
		{"switch without guard", "function f($ip,$a){switch($a){case 1:$b=1;" + use + "}}", 1},
		{"unbraced body", "function f($ip){if(isTrustedProxy())" + use + "}", 0},
		{"unbraced unguarded", "function f($ip,$a){if($a)" + use + "}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, decl+tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedForwardedClientAddress"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
