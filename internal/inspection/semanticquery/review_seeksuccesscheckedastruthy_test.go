package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewNativeConditionUseOwners(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`if($x){}elseif(readdir($d)){}`, "true"},
		{`if($x):elseif(!readdir($d)):endif;`, "true"},
		{`if($x){}else if(readdir($d)){}`, "true"},
		{`for(;readdir($d);){}`, "true"},
		{`for(;$x,($e=readdir($d));){}`, "true"},
		{`for(;readdir($d),$x;){}`, "false"},
		{`for(readdir($d);;readdir($d)){}`, "false,false"},
		{`if(readdir(...)){}`, "false"},
		{`if($x+=readdir($d)){}`, "false"},
		{`if($x??=readdir($d)){}`, "false"},
		{`if($x.=readdir($d)){}`, "false"},
		{`if($x[readdir($d)]=1){}`, "false"},
		{`if($x=readdir($d)){}`, "true"},
	} {
		t.Run(tc.src, func(t *testing.T) {
			var got []string
			r := localContractProbe{"conditions", &got}
			e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{Only: []string{r.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatalf("parse: %+v", f.Errors)
			}
			e.Analyze(f)
			if s := strings.Join(got, ","); s != tc.want {
				t.Fatalf("got %s want %s", s, tc.want)
			}
		})
	}
}
