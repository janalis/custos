package unescapedhtmloutput

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestTextTransformOutput(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   int
	}{
		{`echo '<p>'.htmlspecialchars_decode(htmlspecialchars($_GET['v']));`, 1},
		{`echo '<p>'.html_entity_decode(htmlentities($_GET['v']));`, 1},
		{`echo '<p>'.urldecode(htmlspecialchars($_GET['v']));`, 1},
		{`echo '<p>'.str_replace('&lt;','<',htmlspecialchars($_GET['v']));`, 1},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.htmlspecialchars_decode($_GET['v']);`, 1},
		{`$_GET['v']=rawurlencode($_GET['v']);echo '<p>'.urldecode($_GET['v']);`, 1},
		{`function esc(&$v){$v=htmlspecialchars($v);}esc($_GET['v']);echo '<p>'.htmlspecialchars_decode($_GET['v']);`, 1},
		{`function decode($v){return htmlspecialchars_decode($v);}echo '<p>'.decode(htmlspecialchars($_GET['v']));`, 1},
		{`$_GET['v']='safe';echo '<p>'.htmlspecialchars_decode($_GET['v']);`, 0},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.$_GET['v'];`, 0},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.strtoupper($_GET['v']);`, 0},
		{`echo '<p>'.htmlspecialchars(html_entity_decode($_GET['v']));`, 0},
		{`echo '<p>'.rawurlencode($_GET['v']);`, 0},
		{`echo '<p>'.base64_encode($_GET['v']);`, 0},
		{`echo '<p>'.bin2hex($_GET['v']);`, 0},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.strtoupper(string:$_GET['v']);`, 0},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.strtoupper(...$args);`, 0},
		{`namespace App;function strtoupper($string){return html_entity_decode($string);}$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.strtoupper($_GET['v']);`, 1},
		{`$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.` + strings.Repeat("strtoupper(", 9) + `$_GET['v']` + strings.Repeat(")", 9) + `;`, 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnescapedHtmlOutput"}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("output.php", []byte("<?php "+tc.source), syntax.Options{})
			if len(f.Errors) > 0 {
				t.Fatal(f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %+v want %d", got, tc.want)
			}
		})
	}
}
