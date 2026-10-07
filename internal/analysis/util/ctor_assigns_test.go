package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestCtorAssignsProperty(t *testing.T) {
	src := `<?php
class A {
    private $a; private $b; private $c; private $d;
    public function __construct($x) {
        $this->a = [];
        $this->b = null;
        if ($x) { $this->c = 1; }
        $this->d .= 'x';
    }
}`
	f := parse(t, src)
	var c *syntax.ClassLike
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if cl, ok := n.(*syntax.ClassLike); ok {
			c = cl
		}
		return true
	})
	for name, want := range map[string]bool{"a": true, "b": false, "c": false, "d": false, "e": false} {
		if got := CtorAssignsProperty(c, name); got != want {
			t.Errorf("%s: got %v want %v", name, got, want)
		}
	}
}
