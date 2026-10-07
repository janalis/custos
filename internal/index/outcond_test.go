package index

import "testing"

func TestParamOutAndConditionalReturn(t *testing.T) {
	fs := extract(t, "a.php", `<?php
namespace App;
/**
 * @param-out int $a
 * @phpstan-param-out list<Foo> $b
 * @param-out string $byVal
 * @phpstan-return ($a is string ? Foo : null)
 * @return Foo|null
 */
function f(&$a, &$b, $byVal, &$c) {}
class C {
    /**
     * @template T
     * @param T $v
     * @psalm-return (T is int ? int : string)
     */
    public function m($v) {}
    /** @return (T is int ? int : string) */
    public function noSubject($v) {}
}
`)
	f := fs.Functions[0]
	if f.Params[0].Out != "int" || f.Params[1].Out != `\App\Foo[]` || f.Params[2].Out != "" || f.Params[3].Out != "" {
		t.Errorf("param-out: %+v", f.Params)
	}
	if f.CondReturn != `($a is string ? \App\Foo : null)` {
		t.Errorf("function conditional: %q", f.CondReturn)
	}
	c := fs.Classes[0]
	if got := c.Methods["m"].CondReturn; got != "($v is int ? int : string)" {
		t.Errorf("template subject: %q", got)
	}
	if got := c.Methods["nosubject"].CondReturn; got != "" {
		t.Errorf("unknown subject: %q", got)
	}
}
