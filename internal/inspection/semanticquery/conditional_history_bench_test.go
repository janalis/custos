package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkConditionalHistoryConsumers(b *testing.B) {
	for _, scenario := range []struct {
		name, source string
		dom          bool
	}{
		{"PDOUnconditional", `$s=(new PDO('sqlite::memory:'))->prepare('SELECT :x');$s->bindValue(':x',1);probe($s->execute());`, false},
		{"DOMUnconditional", `$d=new DOMDocument();$d->loadXML('<x><y/></x>');probe($d->documentElement);`, true},
		{"PDO", `$s=(new PDO('sqlite::memory:'))->prepare('SELECT :x');$s->bindValue(':x',1);if($flag){$s->bindValue(':x',2);}probe($s->execute());`, false},
		{"DOM", `$d=new DOMDocument();$d->loadXML('<x><y/></x>');if($flag){$d->loadXML('<a><b/></a>');}probe($d->documentElement);`, true},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			p := nativeProbe{check: func(ctx *analysis.Context, at *syntax.FuncCall) {
				name, ok := at.Name.(*syntax.Name)
				if !ok || name.Value != "probe" {
					return
				}
				value := CallArgument(at.Args, 0, "")
				// Prime the lazy analysis before measuring repeated consumer queries.
				ctx.Flow().Value(value)
				b.ReportAllocs()
				for b.Loop() {
					if scenario.dom {
						ExpansionDOMElementChildren(ctx, value, at)
					} else {
						ExecuteBindings(ctx, value.(*syntax.MethodCall))
					}
				}
			}}
			engine, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
			if err != nil {
				b.Fatal(err)
			}
			engine.Analyze(syntax.Parse("history.php", []byte("<?php "+scenario.source), syntax.Options{}))
		})
	}
}
