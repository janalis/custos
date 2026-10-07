package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestEquivalentFoldNames(t *testing.T) {
	for src, want := range map[string]bool{
		`<?php [f($x), F($x)];`:                 true,
		`<?php [\Ns\f($x), \NS\F($x)];`:         true,
		`<?php [$o->get($x), $o->GET($x)];`:     true,
		`<?php [Foo::make($x), FOO::Make($x)];`: true,
		`<?php [new Foo($x), NEW foo($x)];`:     true,
		`<?php [f(A::K), F(a::K)];`:             true,
		`<?php [f($x), f($X)];`:                 false,
		`<?php [$o->p, $o->P];`:                 false,
		`<?php [f(K), f(k)];`:                   false,
		`<?php [f(A::K), f(A::k)];`:             false,
		`<?php [f('a'), f('A')];`:               false,
		`<?php [f($x), g($x)];`:                 false,
	} {
		f := parse(t, src)
		arr := firstExpr(t, f).(*syntax.Array)
		a := arr.Items[0].Value
		b := arr.Items[1].Value
		if got := EquivalentFoldNames(f, a, b); got != want {
			t.Errorf("%s: got %v, want %v", src, got, want)
		}
	}
}

func TestEquivalentFoldNamesNested(t *testing.T) {
	f := parse(t, `<?php [MD5(StrToLower($user)), md5(strtolower($user))];`)
	arr := firstExpr(t, f).(*syntax.Array)
	if !EquivalentFoldNames(f, arr.Items[0].Value, arr.Items[1].Value) {
		t.Fatal("nested")
	}
}
