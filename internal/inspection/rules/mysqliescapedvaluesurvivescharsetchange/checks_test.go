package mysqliescapedvaluesurvivescharsetchange

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func engine(t *testing.T) *analysis.Engine {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"MysqliEscapedValueSurvivesCharsetChange"}})
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
		{"\n$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { $db->query(\"SELECT '$v'\"); }\n", 1},
		{"\n$db = new mysqli(); $db->set_charset('gbk'); $v = $db->real_escape_string($input); $db->query(\"SELECT '$v'\");\n", 0},
		{"function f(mysqli $db){$db->set_charset('utf8mb4');$v=$db->real_escape_string($x);if($db->set_charset('utf8mb4')){$db->query(\"SELECT '$v'\");}}", 0},
		{"function f(mysqli $db){$db->set_charset('utf8mb4');$v=$db->real_escape_string($x);$db->set_charset('gbk');$db->query(\"SELECT '$v'\");}", 0},
		{"if(false){\n$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { $db->query(\"SELECT '$v'\"); }\n}", 0},
		{"function f(mysqli $d){$d->query();}", 0},
		{"function f(mysqli $d){$d->query(function(){});}", 0},
		{"function f(mysqli $d,mysqli $e){$d->set_charset('utf8mb4');$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$e->query(\"SELECT '$v'\");}}", 0},
		{"function f(mysqli $d){$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$d->query(\"SELECT '$v'\");}}", 0},
		{"function f(mysqli $d){$d->set_charset($charset);$v=$d->real_escape_string($x);if($d->set_charset('gbk')){$d->query(\"SELECT '$v'\");}}", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			got := e.Analyze(syntax.Parse("check.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
