package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

func BenchmarkTypeOfCapturesAndMutations(b *testing.B) {
	for _, scenario := range []struct{ name, body string }{
		{"captures", `$x = 1; $x = 's'; $a = ['k' => 1]; $a['v'] = $x;
            if ($object !== null) { $f = fn() => $object; $f(); }
            $f = function () use ($x, $a) { return [$x, $a]; };
            $g = fn() => fn() => $x; $f(); ($g())();`},
		{"elementMutations", `$a = ['x' => null, 'count' => 1];
            $before = $a['x']++; ++$a['count']; $stored = $a['x'];
            $copy = $a; $read = $copy['count']; $f = fn() => $a['x']; $f();`},
		{"mutations", `$x = null; $y = ++$x; $z = $x++;
            if ($c) { unset($x); } $result = $x;
            $text = '99'; ++$text; $last = $text;`},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			var source strings.Builder
			source.WriteString("<?php class CaptureObject {}\n")
			for i := 0; i < 100; i++ {
				fmt.Fprintf(&source, "function f%d(bool $c, ?CaptureObject $object) { %s }\n", i, scenario.body)
			}
			file := syntax.Parse("bench.php", []byte(source.String()), syntax.Options{Version: phpversion.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(file))
			resolver := names.New(file)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := infer.NewEnv(file, resolver, ix, phpversion.PHP84)
				syntax.InspectFile(file, func(n syntax.Node) bool {
					if v, ok := n.(*syntax.Variable); ok {
						env.TypeOf(v)
					}
					return true
				})
			}
		})
	}
}
