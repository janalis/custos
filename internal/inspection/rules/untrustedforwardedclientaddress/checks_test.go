package untrustedforwardedclientaddress

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UntrustedForwardedClientAddress"}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestChecks counts the findings on short sources.
func TestChecks(t *testing.T) {
	e := engine(t)
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){$f=fn()=>access();}", 0},
		{"if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){echo 'logged';}", 0},
		{"/** @custos-protected */ function access(){} if($log){echo $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1';access();}", 0},
		{"if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){grantAccess();}", 0},
		{"/** @custos-protected */ function access(){} if($_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", 1},
		{"/** @custos-protected */ function access(){} if(trustedProxy() && $_SERVER['HTTP_X_FORWARDED_FOR']==='127.0.0.1'){access();}", 0},
		{"harmless(); $v=1+2; ++$v; echo ''; foreach([] as $entry){} function benign(){return 1;} $object->unrelated();", 0},
		{"if($a===$b){}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
