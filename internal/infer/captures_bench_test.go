package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func BenchmarkTypeOfCapturesAndMutations(b *testing.B) {
	for _, scenario := range []struct{ name, body string }{
		{"captures", `$x = 1; $x = 's'; $a = ['k' => 1]; $a['v'] = $x;
            if ($object !== null) { $f = fn() => $object; $f(); }
            $f = function () use ($x, $a) { return [$x, $a]; };
            $g = fn() => fn() => $x; $f(); ($g())();`},
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
			file := syntax.Parse("bench.php", []byte(source.String()), syntax.Options{Version: phpver.PHP84})
			ix := index.New(stubs.Index())
			ix.Add(index.Extract(file))
			resolver := names.New(file)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				env := infer.NewEnv(file, resolver, ix, phpver.PHP84)
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
