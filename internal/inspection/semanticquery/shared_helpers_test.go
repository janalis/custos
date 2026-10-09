package semanticquery

import (
	"testing"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestArgBindsByRef(t *testing.T) {
	params := []index.Param{{Name: "$a"}, {Name: "$b", ByRef: true}}
	variadic := []index.Param{{Name: "$a"}, {Name: "$rest", ByRef: true, Variadic: true}}
	for _, c := range []struct {
		src     string
		params  []index.Param
		callRef bool
		want    bool
	}{
		{`<?php f(1);`, params, false, false},
		{`<?php f(1, $x);`, params, false, true},
		{`<?php f(b: $x);`, params, false, true},
		{`<?php f(a: $x);`, params, false, false},
		{`<?php f(1, 2, 3);`, variadic, false, true},
		{`<?php f(&$x);`, params, false, false},
		{`<?php f(&$x);`, params, true, true},
	} {
		f := Parse(t, c.src)
		var list *syntax.ArgList
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok {
				list = call.Args
			}
			return true
		})
		if got := ArgBindsByRef(list, c.params, c.callRef); got != c.want {
			t.Errorf("%s (callTimeRef=%v): got %v", c.src, c.callRef, got)
		}
	}
	if ArgBindsByRef(nil, params, true) {
		t.Error("nil list")
	}
}

func TestDocHasAnnotation(t *testing.T) {
	f := Parse(t, `<?php class C {
	/** @var int */
	private $a;
	/** @ORM\Column */
	private $b;
	private $c;
}`)
	var got []bool
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if p, ok := n.(*syntax.Property); ok {
			got = append(got, DocHasAnnotation(f, p))
		}
		return true
	})
	if len(got) != 3 || got[0] || !got[1] || got[2] {
		t.Errorf("got %v", got)
	}
}
