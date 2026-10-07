package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestUnstableVariableInAndVarOccurrences(t *testing.T) {
	f := parse(t, `<?php function g() { $a = 1; $b++; $c .= 'x'; $$d = 2; }`)
	body := f.Stmts[0].(*syntax.Function).Body
	for _, file := range []*syntax.File{f, nil} {
		for name, want := range map[string]bool{"a": false, "b": true, "c": true, "": false} {
			if got := UnstableVariableIn(file, body, name); got != want {
				t.Errorf("file %v: UnstableVariableIn(%q) = %v", file != nil, name, got)
			}
		}
		occ := VarOccurrences(file, body)
		if len(occ["a"]) != 1 || len(occ["b"]) != 1 || len(occ["d"]) != 1 {
			t.Errorf("file %v: VarOccurrences = %v", file != nil, occ)
		}
	}
	if UnstableVariableIn(f, nil, "a") {
		t.Fatal("nil root")
	}
	if len(VarOccurrences(f, nil)) != 0 {
		t.Fatal("nil root occurrences")
	}
}
