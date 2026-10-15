package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestUnavailableNativeConstruction(t *testing.T) {
	for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP55} {
		calls := 0
		p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
			name, ok := c.Name.(*syntax.Name)
			if !ok || name.Value != "probe" {
				return
			}
			calls++
			n := CallArgument(c.Args, 0, "").(*syntax.New)
			if got := NativeNewClass(ctx, n); (got != nil) != (version >= phpversion.PHP55) {
				t.Fatalf("version%s class%+v", version, got)
			}
		}}
		e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}, PHP: version})
		if err != nil {
			t.Fatal(err)
		}
		e.Analyze(syntax.Parse("new.php", []byte(`<?php probe(new IntlGregorianCalendar());$class='IntlGregorianCalendar';probe(new $class());`), syntax.Options{}))
		if calls != 2 {
			t.Fatal(calls)
		}
	}
}

func BenchmarkNativeMethodAvailability(b *testing.B) {
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		name, ok := c.Name.(*syntax.Name)
		if !ok || name.Value != "probe" {
			return
		}
		method := CallArgument(c.Args, 0, "").(*syntax.MethodCall)
		NativeMethod(ctx, method, "IntlGregorianCalendar", "set")
		b.ReportAllocs()
		for b.Loop() {
			NativeMethod(ctx, method, "IntlGregorianCalendar", "set")
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}, PHP: phpversion.PHP84})
	if err != nil {
		b.Fatal(err)
	}
	e.Analyze(syntax.Parse("native.php", []byte(`<?php $calendar=new IntlGregorianCalendar();probe($calendar->set(2026,0,1));`), syntax.Options{}))
}
