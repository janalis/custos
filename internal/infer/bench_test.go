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

// BenchmarkTypeOfVariables types every variable of a file mixing guards,
// assertion calls, method-template calls and untyped properties (the paths
// narrowing walks for each use).
func BenchmarkTypeOfVariables(b *testing.B) {
	var sb strings.Builder
	sb.WriteString(`<?php
class Assert {
    /** @psalm-assert string $v */
    public static function string($v) {}
    /**
     * @template T
     * @param class-string<T> $c
     * @return T
     */
    public static function make(string $c) {}
}
class Foo { public ?int $id = null; }
final class Subject {
    private $count = 0;
    private $items = [];
`)
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&sb, `    public function m%d(?string $s, int|string $u, array $xs) {
        $this->count++;
        $this->items[] = $s;
        Assert::string($u);
        $f = Assert::make(Foo::class);
        if ($s === null) { return $f->id; }
        foo($s, $u, $xs, $this->count, $this->items);
        $x = $s . $u;
        foreach ($xs as $k => $v) { bar($k, $v, $x, $f); }
        return $u;
    }
`, i)
	}
	sb.WriteString("}\n")
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpver.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		f := syntax.Parse("t.php", src, opt)
		ix := index.New(stubs.Index())
		ix.Add(index.Extract(f))
		env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Variable, *syntax.PropertyFetch, *syntax.MethodCall, *syntax.StaticCall:
				env.TypeOf(n.(syntax.Expr))
			}
			return true
		})
	}
}
