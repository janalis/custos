package flow

import "testing"

func BenchmarkEngineCallableContracts(b *testing.B) {
	for name, source := range map[string]string{
		"native":       `$f=strtolower(...);probe($f($_GET['x']));`,
		"method":       `final class R{public static function raw($x){return $x;}}$f=R::raw(...);probe($f($_GET['x']));`,
		"alternatives": `$f=$flag?strtolower(...):htmlspecialchars(...);probe($f($_GET['x']));`,
		"resource":     `function openStream(){return tmpfile();}$h=openStream();if($h!==false){fclose($h);probe(fread($h,1));}`,
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				e, file := summarized(b, source)
				e.Value(argument(callsNamed(file, "probe")[0]))
			}
		})
	}
}
