package parsedqueryvaluedecodedtwice

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
		{"parse_str(\"user.name=a%252Fb\",$q);echo urldecode($q[\"user_name\"]);", 1, ""},
		{"echo urldecode($q[\"user_name\"]);", 0, ""},
		{"if($ok){}echo urldecode($q[\"user_name\"]);", 0, ""},
		{"$q=[];echo urldecode($q[\"user_name\"]);", 0, ""},
		{"other();echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$other);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str($query,$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user[]=a\",$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user.name=a&user_name=b\",$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user%ZZ=a\",$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user.name=%ZZ\",$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"user.name\",$q);echo urldecode($q[\"user_name\"]);", 0, ""},
		{"parse_str(\"x=a%252Fb\",$q);echo rawurldecode($q[\"x\"]);", 1, ""},
		{"parse_str(\"x=plain\",$q);echo urldecode($q[\"x\"]);", 0, ""},
		{"parse_str(\"x=%25ZZ\",$q);echo urldecode($q[\"x\"]);", 0, ""},
		{"parse_str(\"x=a\",$q);echo urldecode($q[\"other\"]);", 0, ""},
		{"parse_str(\"x=a\",$q);echo urldecode($x);", 0, ""},
		{"parse_str(\"x=a%252Fb\",$q);echo urldecode($obj->q[\"x\"]);", 0, ""},
		{"parse_str(\"x=a%252Fb\",$q);echo urldecode($q[$key]);", 0, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ParsedQueryValueDecodedTwice"}, PHP: phpversion.MustParse(tc.php)}
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
