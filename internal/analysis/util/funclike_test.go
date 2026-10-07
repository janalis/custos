package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestEquivalentFuncLike(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{"<?php $a === $a;", true},
		{"<?php $a['k'] === $a [ 'k' ];", true},
		{"<?php $a->b === $a /* c */ ->b;", true},
		{"<?php $a === $b;", false},
		{"<?php ($a) === $a;", false},
		{"<?php $a['k'] === $a['j'];", false},
		{"<?php f(1) === f(1, 2);", false},
	}
	for _, c := range cases {
		f := parse(t, c.src)
		b := firstExpr(t, f).(*syntax.Binary)
		if got := Equivalent(f, b.Left, b.Right); got != c.want {
			t.Errorf("%s: Equivalent = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestEnclosingFuncLike(t *testing.T) {
	f := parse(t, "<?php function f() { $g = function () { return $x; }; } $y;")
	var x, y *syntax.Variable
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok {
			switch v.Name {
			case "x":
				x = v
			case "y":
				y = v
			}
		}
		return true
	})
	if fn, ok := syntax.EnclosingFuncLike(x).(*syntax.Closure); !ok || syntax.FuncLikeBody(fn) == nil {
		t.Errorf("$x: want enclosing closure, got %T", syntax.EnclosingFuncLike(x))
	}
	if fn := syntax.EnclosingFuncLike(y); fn != nil {
		t.Errorf("$y: want nil, got %T", fn)
	}
}

func TestQuotedStringContentAndLastNamePart(t *testing.T) {
	f := parse(t, `<?php f('a@', "b", 1);`)
	call := firstExpr(t, f).(*syntax.FuncCall)
	want := []struct {
		s  string
		ok bool
	}{{"a@", true}, {"b", true}, {"", false}}
	for i, a := range call.Args.Args {
		s, ok := QuotedStringContent(a.(*syntax.Arg).Value)
		if s != want[i].s || ok != want[i].ok {
			t.Errorf("arg %d: got (%q, %v)", i, s, ok)
		}
	}
	for in, out := range map[string]string{`\A\b`: "b", "c": "c", `Ns\d`: "d"} {
		if got := LastNamePart(in); got != out {
			t.Errorf("LastNamePart(%q) = %q", in, got)
		}
	}
}

func TestQuotedStringContentPrefixed(t *testing.T) {
	f := parse(t, `<?php b'x';`)
	if _, ok := QuotedStringContent(firstExpr(t, f)); ok {
		t.Fatal("binary-prefixed literal is not plainly quoted")
	}
}
