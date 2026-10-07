package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestIsFuncNamedFold(t *testing.T) {
	src := "<?php ARRAY_Search($a, $b); \\Foo\\Count($x); $o->count();"
	f := syntax.Parse("x.php", []byte(src), syntax.Options{})
	var got []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		for _, want := range []string{"array_search", "count"} {
			if c, ok := IsFuncNamedFold(n, want); ok {
				_, p, _ := FuncNamePart(c)
				got = append(got, p)
			}
		}
		return true
	})
	if len(got) != 2 || got[0] != "ARRAY_Search" || got[1] != "Count" {
		t.Fatalf("got %v", got)
	}
}
