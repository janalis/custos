package astquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestMemberRemovalSpans(t *testing.T) {
	src := "<?php\nclass A {\n    public $x;\n    /** doc */\n    // note\n    public function m() {}\n    public $y;\n}\n"
	f := Parse(t, src)
	var m *syntax.Method
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if mm, ok := n.(*syntax.Method); ok {
			m = mm
		}
		return true
	})
	out := []byte(src)
	spans := MemberRemovalSpans(f, m)
	for i := len(spans) - 1; i >= 0; i-- {
		s := spans[i]
		out = append(out[:s.Start:s.Start], out[s.End:]...)
	}
	want := "<?php\nclass A {\n    public $x;\n    \n    // note\n    public $y;\n}\n"
	if string(out) != want {
		t.Fatalf("got %q", out)
	}
}
