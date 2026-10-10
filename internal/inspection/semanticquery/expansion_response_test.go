package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionResponseProof(t *testing.T) {
	for _, tc := range []struct {
		source string
		known  bool
		size   int64
	}{
		{`header('Content-Length: 1'); echo 'x';`, true, 1},
		{`header('Content-Length: 9',true,206);header('Location: /next');echo 'x';`, false, 1},
		{`header('Content-Length: 9');header('Status: 204 No Content');echo 'x';`, false, 1},
		{`header('Content-Length: 1');echo str_repeat('x',1073741824),'x';`, false, 1073741824},
		{`header('Content-Encoding: gzip');header('Content-Length: 1');echo 'x';`, true, 1},
		{`header('Content-Length: 1');header('X: y',false);echo 'x';`, true, 1},
		{`header('Content-Length: 1'); echo 'é';`, true, 2},
		{`header('Content-Length: 1'); echo 'a'.'b';`, true, 2},
		{`header('Content-Length: 1'); echo str_repeat('ab',3);`, true, 6},
		{`header('Content-Length: 1'); echo $x;`, false, 0},
		{`header('Content-Length: 1'); echo str_repeat($x,3);`, false, 0},
		{`header('Content-Length: 1'); echo str_repeat('ab',-1);`, false, 0},
		{`header('Content-Length: 1'); echo str_repeat('ab',999999999999);`, false, 0},
		{`header('Content-Length: 1'); echo str_repeat('ab',1000000000);`, false, 0},
		{`header('Content-Length: 1'); echo strlen('ab');`, false, 0},
		{`header('Content-Length: 1'); echo 2+3;`, false, 0},
		{`header('Content-Length: 1'); echo 'a'.$x;`, false, 0},
		{`header('Content-Length: 1'); print 'abc';`, true, 3},
		{`header('Content-Length: 1'); print $x;`, false, 0},
		{`header('Content-Length: 1'); exit('abc'); echo 'z';`, true, 3},
		{`header('Content-Length: 1'); exit($x);`, false, 0},
		{`header('Content-Length: 1'); exit;`, true, 0},
		{`header('Content-Length: 1'); header('Transfer-Encoding: chunked'); echo 'x';`, false, 1},
		{`header('Content-Length: 1'); http_response_code(304);echo 'x';`, false, 1},
		{`header('Content-Length: 1'); http_response_code($x);echo 'x';`, false, 1},
		{`header('Content-Length: 1'); http_response_code();echo 'x';`, true, 1},
		{`header('Content-Length: 1',true,$x);echo 'x';`, false, 1},
		{`header('Content-Length: 1',true,0);echo 'x';`, true, 1},
		{`header('Content-Length: 1',true,206);echo 'x';`, true, 1},
		{`header('Content-Length: 1'); header_remove();echo 'x';`, true, 1},
		{`header('Content-Length: 1'); header_remove('content-length');echo 'x';`, true, 1},
		{`header('Content-Length: 1'); header_remove($x);echo 'x';`, false, 1},
		{`header('Content-Length: 1'); header($x);echo 'x';`, false, 1},
		{`header('Content-Length: 1'); unknown();echo 'x';`, false, 1},
		{`header('Content-Length: 1'); $a=1;echo 'x';`, false, 1},
		{`header('Content-Length: 1'); if($x)echo 'x';`, false, 0},
		{`namespace Test {header('Content-Length: 1');echo 'x';}`, true, 1},
		{`function f(){header('Content-Length: 1');echo 'x';}`, false, 1},
		{`function f(){header('Content-Length: 1');echo 'x';exit;}`, true, 1},
	} {
		t.Run(tc.source, func(t *testing.T) {
			seen := false
			p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
				k, _, _, ok := ExpansionHeader(ctx, c)
				if !ok || k != "content-length" {
					return
				}
				seen = true
				r := ExpansionResponseAt(ctx, c)
				if r.BodyKnown != tc.known || r.BodyLength != tc.size {
					t.Fatalf("got %+v want known=%v length=%d", r, tc.known, tc.size)
				}
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			e.Analyze(syntax.Parse("proof.php", []byte("<?php "+tc.source), syntax.Options{}))
			if !seen {
				t.Fatal("missing header")
			}
		})
	}
}

func TestExpansionHeaderProof(t *testing.T) {
	for _, tc := range []struct {
		source string
		known  bool
	}{
		{`header('X: y');`, true}, {`header('X: y',false);`, true}, {`header('X: y',$x);`, false}, {`header("X: y\r\nZ: q");`, false}, {`header('HTTP/1.1 200 OK');`, false}, {`strlen('x');`, false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
				_, _, _, ok := ExpansionHeader(ctx, c)
				if ok != tc.known {
					t.Fatalf("known=%v", ok)
				}
			}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				t.Fatal(err)
			}
			e.Analyze(syntax.Parse("header.php", []byte("<?php "+tc.source), syntax.Options{}))
		})
	}
}

func TestExpansionResponseBounds(t *testing.T) {
	source := `<?php header('Content-Length: 1');` + strings.Repeat(`echo '';`, 4097)
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ExpansionResponseAt(ctx, c).BodyKnown {
			t.Fatal("statement budget exceeded")
		}
		if _, ok := expansionByteLength(ctx, CallArgument(c.Args, 0, "header"), 9); ok {
			t.Fatal("depth budget exceeded")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("bounded.php", []byte(source), syntax.Options{}))
}

func TestExpansionDecimalLength(t *testing.T) {
	for _, tc := range []struct {
		input string
		ok    bool
	}{{"", false}, {"-1", false}, {"999999999999999999999", false}, {"001", true}} {
		_, ok := ExpansionDecimalLength(tc.input)
		if ok != tc.ok {
			t.Fatalf("%q: %v", tc.input, ok)
		}
	}
}

func BenchmarkExpansionResponse(b *testing.B) {
	file := syntax.Parse("bench.php", []byte(`<?php header('Content-Length: 9');header('Vary: Accept');echo str_repeat('x',9);`), syntax.Options{})
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) { ExpansionResponseAt(ctx, c) }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
