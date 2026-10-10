package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

func TestNativeRegexCaptures(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		count   int
		known   bool
	}{
		{"", 0, false},
		{"x", 0, false},
		{"abc", 0, false},
		{"\\xx\\", 0, false},
		{" x ", 0, false},
		{"/x", 0, false},
		{"{x}", 0, false},
		{"/x/", 0, true},
		{"/(x)/", 1, true},
		{"/\\(x\\)/", 0, true},
		{"/[(]/", 0, true},
		{"/(?:x)(?=y)(?!z)/", 0, true},
		{"/(?<label>x)/", 1, true},
		{"/(?<=x)(?<!y)/", 0, true},
		{"/(?'label'x)/", 1, true},
		{"/(?P<label>x)/", 1, true},
		{"/(?/", 0, false},
		{"/(?</", 0, false},
		{"/(?)/", 0, false},
		{"/(?P)/", 0, false},
		{"/(?P=x)/", 0, false},
		{"/(?<)/", 0, false},
		{"/(?|x)/", 0, false},
		{"/)/", 0, false},
		{"/(/", 1, false},
		{"/[/", 0, false},
		{"/(*SKIP)/", 0, false},
		{"/(x)/n", 0, false},
		{"/(x)/x", 0, false},
		{`/\Q(a)\E(b)/`, 0, false},
		{`/\E/`, 0, false},
		{`/\c(a)/`, 0, false},
		{`/\/`, 0, false},
		{`/a/b/`, 0, false},
		{`/[[:alpha:](]/`, 0, false},
		{`/[]()]/`, 0, false},
		{`/[^]()]/`, 0, false},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			count, known := NativeRegexCaptures(tc.pattern)
			if count != tc.count || known != tc.known {
				t.Fatalf("got(%d,%v),want(%d,%v)", count, known, tc.count, tc.known)
			}
		})
	}
}

func TestNativeDateContracts(t *testing.T) {
	for _, tc := range []struct{ source, class string }{
		{"probe(new DateTime());", "DateTime"},
		{"probe(new DateTimeImmutable());", "DateTimeImmutable"},
		{"probe(new DateInterval('P1D'));", "DateInterval"},
		{"probe(new stdClass());", ""},
		{"probe($unknown);", ""},
	} {
		probeNative(t, tc.source, func(ctx *analysis.Context, c *syntax.FuncCall) {
			if got := NativeDateClass(ctx, CallArgument(c.Args, 0, "value")); got != tc.class {
				t.Fatalf("class=%q,want%q", got, tc.class)
			}
		})
	}
	for _, tc := range []struct {
		source string
		static bool
		format string
	}{
		{"probe(DateTimeImmutable::getLastErrors());", true, ""},
		{"probe(Other::getLastErrors());", false, ""},
		{"probe(DateTimeImmutable::createFromFormat('Y',$x));", false, ""},
		{"probe(Unknown::$method());", false, ""},
		{"probe(date('Y'));", false, "'Y'"},
		{"probe(gmdate('o-W'));", false, "'o-W'"},
		{"probe((new DateTime())->format('c'));", false, "'c'"},
		{"probe((new DateInterval('P1D'))->format('%d'));", false, ""},
		{"probe((new DateTime())->modify('+1 day'));", false, ""},
		{"probe((new DateTime())->{$method}());", false, ""},
		{"probe(42);", false, ""},
	} {
		probeNative(t, tc.source, func(ctx *analysis.Context, c *syntax.FuncCall) {
			arg := CallArgument(c.Args, 0, "value")
			if got := NativeDateStatic(ctx, arg, "getLastErrors") != nil; got != tc.static {
				t.Errorf("static=%v", got)
			}
			if got := ctx.Text(NativeDateFormat(ctx, arg)); got != tc.format {
				t.Errorf("format=%q,want%q", got, tc.format)
			}
		})
	}
	tokens := NativeFormatTokens(`Y-W \Y H:i:s\`)
	if !tokens['Y'] || !tokens['W'] || tokens['\\'] {
		t.Fatalf("tokens=%v", tokens)
	}
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		arg := CallArgument(c.Args, 0, "value")
		if NativeDateClass(ctx, arg) != "" {
			t.Error("unavailable date class resolved")
		}
		if NativeDateStatic(ctx, arg, "getLastErrors") != nil {
			t.Error("unavailable static API resolved")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{PHP: phpversion.PHP53, Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.Analyze(syntax.Parse("old.php", []byte("<?php probe(new DateTimeImmutable());probe(DateTimeImmutable::getLastErrors());"), syntax.Options{}))
	p.check = func(ctx *analysis.Context, c *syntax.FuncCall) {
		arg := CallArgument(c.Args, 0, "value")
		if NativeDateClass(ctx, arg) != "" {
			t.Error("missing class declaration resolved")
		}
		if NativeDateStatic(ctx, arg, "getLastErrors") != nil {
			t.Error("missing static declaration resolved")
		}
	}
	e, err = analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	e.WithIndex(index.New(nil)).Analyze(syntax.Parse("missing.php", []byte("<?php probe(new DateTime());probe(DateTime::getLastErrors());"), syntax.Options{}))
}

func TestNativeRegexFlags(t *testing.T) {
	for _, tc := range []struct {
		source     string
		has, known bool
	}{
		{"probe(PREG_OFFSET_CAPTURE);", true, true},
		{"probe(PREG_UNMATCHED_AS_NULL);", false, true},
		{"probe(PREG_OFFSET_CAPTURE|PREG_UNMATCHED_AS_NULL);", true, true},
		{"probe(PREG_OFFSET_CAPTURE|$unknown);", true, false},
		{"probe(256);", true, true},
		{"probe(0);", false, true},
		{"probe(0|0);", false, true},
		{"probe($unknown);", false, false},
	} {
		probeNative(t, tc.source, func(ctx *analysis.Context, c *syntax.FuncCall) {
			has, known := NativeFlagContains(ctx, CallArgument(c.Args, 0, "value"), "PREG_OFFSET_CAPTURE")
			if has != tc.has || known != tc.known {
				t.Fatalf("got(%v,%v),want(%v,%v)", has, known, tc.has, tc.known)
			}
		})
	}
	probeNative(t, "probe(4);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		if _, known := NativeFlagContains(ctx, CallArgument(c.Args, 0, "value"), "unknown"); known {
			t.Error("unknown flag contract")
		}
	})
}

func BenchmarkNativeRegexCaptures(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		NativeRegexCaptures(`/^(?<name>[a-z]+)(?:,\s*([0-9]+))?$/u`)
	}
}

func TestNativeRegexNumericSplitFlag(t *testing.T) {
	probeNative(t, "probe(2);", func(ctx *analysis.Context, c *syntax.FuncCall) {
		has, known := NativeFlagContains(ctx, CallArgument(c.Args, 0, "value"), "PREG_SPLIT_DELIM_CAPTURE")
		if !has || !known {
			t.Error("known numeric flag lost")
		}
	})
}
