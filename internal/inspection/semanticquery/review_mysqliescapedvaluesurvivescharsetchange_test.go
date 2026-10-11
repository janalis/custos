package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewMysqliEscapedValueSurvivesCharsetChange(t *testing.T) {
	const head = "<?php function f(mysqli $d,mysqli $e,$x,$c){$d->set_charset('utf8mb4');"
	const escape = "$v=$d->real_escape_string($x);"
	const query = "$d->query(\"SELECT '$v'\");"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"changed", head + escape + "if($d->set_charset('gbk')){" + query + "}}", 1},
		{"changed behind failure guard", head + escape + "if(!$d->set_charset('gbk')){return;}" + query + "}", 1},
		{"changed twice", head + escape + "if($d->set_charset('gbk')){if($d->set_charset('latin1')){" + query + "}}}", 1},
		{"restored", head + escape + "if($d->set_charset('latin1')){$d->query('SELECT 1');if($d->set_charset('utf8mb4')){" + query + "}}}", 0},
		{"restored in another case", head + escape + "if($d->set_charset('latin1')){if($d->set_charset('UTF8MB4')){" + query + "}}}", 0},
		{"restored without a check", head + escape + "if($d->set_charset('gbk')){$d->set_charset('utf8mb4');" + query + "}}", 0},
		{"maybe restored", head + escape + "if($d->set_charset('gbk')){if($c){$d->set_charset('utf8mb4');}" + query + "}}", 0},
		{"later unknown charset", head + escape + "if($d->set_charset('gbk')){if($d->set_charset($c)){" + query + "}}}", 0},
		{"other connection restored", head + escape + "if($d->set_charset('gbk')){if($e->set_charset('utf8mb4')){" + query + "}}}", 1},
		{"selection after the query", head + escape + "if($d->set_charset('gbk')){" + query + "$d->set_charset('utf8mb4');}}", 1},
		{"maybe reselected before escaping", head + "if($c){$d->set_charset('gbk');}" + escape + "if($d->set_charset('gbk')){" + query + "}}", 0},
		{"reselected before escaping", head + "$d->set_charset('gbk');" + escape + "if($d->set_charset('gbk')){" + query + "}}", 0},
		{"reselected then changed", head + "$d->set_charset('gbk');" + escape + "if($d->set_charset('utf8mb4')){" + query + "}}", 1},
		{"checked selection before escaping", "<?php function f(mysqli $d,$x){if($d->set_charset('utf8mb4')){" + escape + "if($d->set_charset('gbk')){" + query + "}}}", 1},
		{"no selection before escaping", "<?php function f(mysqli $d,$x){" + escape + "if($d->set_charset('gbk')){" + query + "}}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := expansionCryptoDBProbe{id: "MysqliEscapedValueSurvivesCharsetChange", kinds: []syntax.NodeKind{syntax.KMethodCall}, check: func(ctx *analysis.Context, n syntax.Node) {
				CheckMysqliEscapedValueSurvivesCharsetChange(ctx, n, "finding")
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("case.php", []byte(tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v\n%s", len(got), tc.want, got, tc.src)
			}
		})
	}
}
