package astquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestStableArrayRead(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{`echo $a[0];`, true},
		{`echo $a[0][1];`, true},
		{`echo foo()[0];`, false},
		{`$a[0]=1;`, false},
		{`unset($a[0]);`, false},
		{`$a[0]++;`, false},
		{`f($a[0]);`, false},
		{`echo count($a=[]),$a[0];`, false},
		{`echo $a++,$a[0];`, false},
		{`echo f($a),$a[0];`, false},
		{`echo $o->f($a),$a[0];`, false},
		{`echo C::f($a),$a[0];`, false},
		{`echo new C($a),$a[0];`, false},
		{`foreach($x as $a[0]) {}`, false},
		{`foreach($x as $a[0]=>$v) {}`, false},
		{`echo function() {$a=[];},$a[0];`, true},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := Parse(t, "<?php "+tc.src)
			var fetch *syntax.ArrayDimFetch
			syntax.InspectFile(f, func(n syntax.Node) bool {
				if x, ok := n.(*syntax.ArrayDimFetch); ok && fetch == nil {
					fetch = x
				}
				return true
			})
			if got := StableArrayRead(f, fetch); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			if got := StableArrayRead(f, fetch); got != tc.want {
				t.Fatal("cached result changed")
			}
		})
	}
	f := Parse(t, "<?php echo $a[0];")
	if StableArrayRead(f, &syntax.ArrayDimFetch{Var: &syntax.Variable{Name: "a"}}) {
		t.Fatal("unattached fetch")
	}
}

func BenchmarkStableArrayRead(b *testing.B) {
	f := syntax.Parse("bench.php", []byte("<?php echo $a[0];"), syntax.Options{})
	var fetch *syntax.ArrayDimFetch
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if x, ok := n.(*syntax.ArrayDimFetch); ok {
			fetch = x
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		StableArrayRead(f, fetch)
	}
}
