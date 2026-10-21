package curlreadcallbackexceedsrequestedsize

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewCurlReadCallbackExceedsRequestedSizeRebinding(t *testing.T) {
	p := expansionCryptoDBProbe{id: "CurlReadCallbackExceedsRequestedSize", kinds: []syntax.NodeKind{syntax.KFuncCall}, check: func(ctx *analysis.Context, n syntax.Node) {
		check(ctx, n, "finding")
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}, PHP: phpversion.PHP85})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"untouched", "", 1},
		{"plain reads", "if ($limit > 4) { echo $limit, \"asked $limit\", isset($limit); } $copy = $limit; $list = [$limit]; $limit + 1;", 1},
		{"expression argument", "error_log('asked ' . $limit); f($limit + 1);", 1},
		{"other variable written", "$other = 1; $other++; foreach ($rows as $k => $row) {} unset($other); f($other);", 1},
		{"value capture", "$f = function () use ($limit) { $limit = 0; }; $g = fn () => $limit--; $f(); $g();", 1},
		{"other reference capture", "$f = function () use (&$other) { $other = 0; };", 1},
		{"other catch", "try { f(); } catch (Exception $e) {} try { f(); } catch (Exception) {}", 1},
		{"value array item", "$pair = [$limit, &$other];", 1},
		{"assignment", "$limit = 0;", 0},
		{"compound assignment", "$limit -= 8;", 0},
		{"destructuring", "[$a, $limit] = [1, 2];", 0},
		{"increment", "$limit--;", 0},
		{"reference argument", "clamp($limit);", 0},
		{"parenthesized argument", "clamp(($limit));", 0},
		{"named argument", "clamp(n: $limit);", 0},
		{"method argument", "$o->clamp($limit);", 0},
		{"static argument", "C::clamp($limit);", 0},
		{"constructor argument", "new C($limit);", 0},
		{"builtin reference argument", "settype($limit, 'string');", 0},
		{"foreach value", "foreach ([$limit - 8] as $limit) {}", 0},
		{"foreach key", "foreach ([1] as $limit => $v) {}", 0},
		{"reference assignment", "$alias = &$limit; $alias = 0;", 0},
		{"reference array item", "$box = [&$limit]; $box[0] = 0;", 0},
		{"reference capture", "$f = function () use (&$limit) { $limit = 0; }; $f();", 0},
		{"global", "global $limit;", 0},
		{"static", "static $limit = 0;", 0},
		{"unset", "unset($limit);", 0},
		{"catch", "try { f(); } catch (Exception $limit) {}", 0},
		{"variable variable", "$name = 'limit'; $$name = 0;", 0},
		{"extract", "extract(['limit' => 0]);", 0},
		{"parse_str", "parse_str('limit=0');", 0},
		{"include", "include 'clamp.php';", 0},
		{"eval", "eval('$limit = 0;');", 0},
		{"write after return in loop", "while (true) { if ($limit < 9) { return str_repeat('x', $limit + 1); } $limit--; }", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "<?php curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { " + tc.body + " return str_repeat('x', $limit + 1); });"
			f := syntax.ParseBest("review.php", []byte(src), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

type expansionCryptoDBProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p expansionCryptoDBProbe) ID() string { return p.id }

func (p expansionCryptoDBProbe) Semantic() {}

func (p expansionCryptoDBProbe) Flow() {}

func (p expansionCryptoDBProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p expansionCryptoDBProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
