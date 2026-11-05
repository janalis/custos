package flow

import "testing"

func BenchmarkExtendedCallableContracts(b *testing.B) {
	for name, source := range map[string]string{
		"map":       `$xs=array_map('htmlspecialchars',[$_GET['x'],$_GET['y']]);probe($xs[0]);`,
		"names":     `$cb=$flag?'strtolower':'htmlspecialchars';probe($cb($_GET['x']));`,
		"invokable": `final class R{function __invoke($x){return $x;}}$cb=new R;probe($cb($_GET['x']));`,
		"returned":  `function factory(string $prefix){return fn($x)=>$prefix.$x;}$cb=factory('prefix');probe($cb($_GET['x']));`,
		"guard":     `function f(){$h=tmpfile();$cb=fclose(...);if($cb($h)){probe($h);}}`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				env, file := summarized(b, source)
				env.Value(argument(callsNamed(file, "probe")[0]))
			}
		})
	}
}
