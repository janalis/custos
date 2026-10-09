package syntax

import (
	"testing"

	"custos/internal/phpver"
)

func TestCloneArguments(t *testing.T) {
	for _, tc := range []struct {
		src          string
		count        int
		object, with string
	}{
		{`clone($a, $b)`, 2, "$a", "$b"},
		{`clone(withProperties: $b, object: $a)`, 2, "$a", "$b"},
		{`clone(object: $a)`, 1, "$a", ""},
		{`clone($a)`, 1, "$a", ""},
		{`clone()`, 0, "", ""},
		{`clone(...)`, 1, "", ""},
		{`clone(...$args)`, 1, "", ""},
		{`clone($a, ...$args)`, 2, "", ""},
		{`clone($a, $b, $c)`, 3, "$a", "$b"},
		{`clone(unknown: $a)`, 1, "", ""},
		{`clone(&$a)`, 1, "$a", ""},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := parse(t, "<?php "+tc.src+";", phpver.PHP85)
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			n := firstExpr(t, f).(*Clone)
			if n.Args == nil || len(n.Args.Args) != tc.count {
				t.Fatal("arguments lost")
			}
			for _, pair := range []struct {
				expr Expr
				want string
			}{{n.Expr, tc.object}, {n.With, tc.with}} {
				got := ""
				if pair.expr != nil {
					got = string(f.Src[pair.expr.Span().Start:pair.expr.Span().End])
				}
				if got != pair.want {
					t.Errorf("binding: got %q want %q", got, pair.want)
				}
			}
		})
	}
}

func TestClonePostfixGrammar(t *testing.T) {
	for _, tc := range []struct {
		src   string
		valid bool
	}{
		{`clone($a)->x`, true},
		{`(clone($a, []))->x`, true},
		{`clone($a, [])->x`, false},
		{`clone(object: $a)->x`, false},
		{`clone($a,)->x`, false},
		{`clone(...)->x`, false},
		{`clone $a->x`, true},
	} {
		f := parse(t, "<?php "+tc.src+";", phpver.PHP85)
		if (len(f.Errors) == 0) != tc.valid {
			t.Errorf("%s: %v", tc.src, f.Errors)
		}
	}
	n := firstExpr(t, parse(t, `<?php clone($a)->x;`, phpver.PHP85)).(*Clone)
	if _, ok := n.Expr.(*PropertyFetch); !ok {
		t.Fatalf("postfix binds to %T", n.Expr)
	}
	f := parse(t, `<?php clone($a);`, phpver.PHP84)
	if firstExpr(t, f).(*Clone).Args != nil {
		t.Fatal("function form before PHP 8.5")
	}
}

func BenchmarkCloneArguments(b *testing.B) {
	src := []byte(`<?php clone(withProperties: ['x' => $x], object: $object);`)
	b.ReportAllocs()
	for b.Loop() {
		Parse("clone.php", src, Options{Version: phpver.PHP85})
	}
}
