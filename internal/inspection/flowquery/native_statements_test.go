package flowquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestNativePriorStatements(t *testing.T) {
	file := syntax.Parse("test.php", []byte("<?php $first=1; if($flag){$inner=2; probe();} finish();"), syntax.Options{})
	calls := map[string]*syntax.FuncCall{}
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.FuncCall); ok {
			calls[call.Name.(*syntax.Name).Value] = call
		}
		return true
	})
	if prior := NativePriorStatements(file, calls["probe"]); len(prior) != 1 {
		t.Fatalf("nested call must see only its block: %+v", prior)
	}
	if prior := NativePriorStatements(file, calls["finish"]); len(prior) != 2 {
		t.Fatalf("top-level statements: %+v", prior)
	}
	if prior := NativePriorStatements(file, calls["probe"]); len(prior) != 1 {
		t.Fatalf("cached lookup changed: %+v", prior)
	}
	if NativePriorStatements(file, nil) != nil {
		t.Fatal("an absent node has no containing statement")
	}
}
