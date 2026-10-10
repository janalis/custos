package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestExpansionDFiles(t *testing.T) {
	for _, tc := range []struct {
		body          string
		success, want bool
	}{
		{"return true;", true, true},
		{"return false;", false, true},
		{"throw new Exception;", false, true},
		{"throw new Exception;", true, false},
		{"return true;", false, false},
		{"return false;", true, false},
		{"", true, false},
		{"echo 1;", true, false},
		{"return $x;", false, false},
		{"return null;", false, false},
		{"return false; return true;", true, false},
	} {
		probeNative(t, "probe(function(){"+tc.body+"});", func(ctx *analysis.Context, c *syntax.FuncCall) {
			closure := CallArgument(c.Args, 0, "").(*syntax.Closure)
			if got := ExpansionDFinalOutcome(ctx, closure.Body, tc.success); got != tc.want {
				t.Fatalf("%s: %v", tc.body, got)
			}
		})
	}
	probeNative(t, "probe();", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ExpansionDPathCall(ctx, c) {
			t.Fatal("unknown function is not path API")
		}
		if len(ExpansionDLocalEvents(ctx, c)) != 1 {
			t.Fatal("scope events missing")
		}
	})
	for _, name := range []string{"file_get_contents", "file_put_contents", "fopen", "file", "is_file", "is_dir", "file_exists", "unlink", "mkdir", "rmdir", "scandir", "opendir"} {
		probeNative(t, "probe("+name+"($p));", func(ctx *analysis.Context, c *syntax.FuncCall) {
			if !ExpansionDPathCall(ctx, CallArgument(c.Args, 0, "").(*syntax.FuncCall)) {
				t.Fatal(name)
			}
		})
	}
	probeNative(t, strings.Repeat("$v=1;", 513)+"probe();", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if ExpansionDLocalEvents(ctx, c) != nil {
			t.Fatal("exhausted event budget must remain unknown")
		}
	})
	probeNative(t, "$cb=function(){unlink($p);};probe();", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if len(ExpansionDLocalEvents(ctx, c)) != 2 {
			t.Fatal("nested scopes leaked")
		}
	})
}

func BenchmarkExpansionDLocalEvents(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) { ExpansionDLocalEvents(ctx, c) }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php " + strings.Repeat("$h=fopen($p,'r');probe($h);", 40))
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.Parse("bench.php", src, syntax.Options{}))
	}
}

func TestExpansionDCallbackTermination(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{"return 'x';", true},
		{"if (true) return 'x';", true},
		{"if (1) { echo 'x'; return 'x'; }", true},
		{"if (false) echo 'x'; else return 'x';", true},
		{"if (false) return 'x';", false},
		{"if ($unknown) return 'x';", false},
		{"if (true) echo 'x'; elseif ($x) return 'x';", false},
		{"echo 'x';", false},
		{strings.Repeat("if (true) {", 65) + "return 'x';" + strings.Repeat("}", 65), false},
	} {
		probeNative(t, "probe(function(){"+tc.body+"});", func(ctx *analysis.Context, c *syntax.FuncCall) {
			body := CallArgument(c.Args, 0, "").(*syntax.Closure).Body
			if got := ExpansionDCallbackTerminates(ctx, body); got != tc.want {
				t.Fatalf("%s: got %v want %v", tc.body, got, tc.want)
			}
		})
	}
}

func BenchmarkExpansionDCallbackTermination(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		body := CallArgument(c.Args, 0, "").(*syntax.Closure).Body
		ExpansionDCallbackTerminates(ctx, body)
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php probe(function(){if (true) return 'x';});"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(file)
	}
}
