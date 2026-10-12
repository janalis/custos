package unescapedhtmloutput

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestIndependentNativeEdges(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"echo '<script>let label=' . json_encode($_GET['label']) . ';</script>';", 0},
		// Parts that cannot hold markup do not hide the request text behind them.
		{"$x = $_GET['a']; echo '<i>' . htmlspecialchars($x) . '</i><b>' . $x . '</b>';", 1},
		{"$x = $_GET['a']; $y = (int) $x; echo \"<i>$y</i><b>$x</b>\";", 1},
		{"$x = $_GET['a']; echo '<i>', strlen($x), '</i><b>', $x, '</b>', $x;", 1},
		{"$x = $_GET['a']; echo '<i>' . htmlspecialchars($x) . '</i><b>' . (int) $x . '</b>';", 0},
		{"$x = $_GET['a']; echo '<i>' . $row['t'] . '</i><b>' . $x . '</b>';", 0},
		{"$x = $_GET['a']; echo \"<a title='\" . htmlspecialchars($x) . \"'>\" . $x;", 0},
		{"$x = $_GET['a']; echo '<i>' . (int) $x . '</i><b>' . $x . '</b>' . done();", 0},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnescapedHtmlOutput"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
