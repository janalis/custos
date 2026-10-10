package parsestrkeynormalizationmismatch

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
		{"parse_str(\"user.name=a%252Fb\",$q);echo $q[\"user.name\"];", 1, ""},
		{"echo $q[\"user.name\"];", 0, ""},
		{"if($ok){}echo $q[\"user.name\"];", 0, ""},
		{"$q=[];echo $q[\"user.name\"];", 0, ""},
		{"other();echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$other);echo $q[\"user.name\"];", 0, ""},
		{"parse_str($query,$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user[]=a\",$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name=a&user_name=b\",$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user%ZZ=a\",$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name=%ZZ\",$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name\",$q);echo $q[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$q);echo isset($q[\"user.name\"]);", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$q);echo ($q[\"user.name\"]??\"\");", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$q);echo $q[\"plain\"];", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$q);echo $other[\"user.name\"];", 0, ""},
		{"parse_str(\"user.name=a%252Fb\",$q);$key=\"user.name\";echo $q[$key];", 0, ""},
		{"parse_str(\"x=a\",$q);echo $obj->q[\"user.name\"];", 0, ""},
		{"parse_str(\"x=a\",$q);echo $q[$key];", 0, ""},
		{"$key=\"user.name\";parse_str(\"user.name=a\",$q);echo $q[$key];", 1, ""},
	} {
		t.Run(tc.source+tc.php, func(t *testing.T) {
			cfg := analysis.Config{Only: []string{"ParseStrKeyNormalizationMismatch"}, PHP: phpversion.MustParse(tc.php)}
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
