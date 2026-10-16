package reflectioncompositetypeassumednamed

import (
	"fmt"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

func TestAssertionAssumption(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, description := range []string{"", ", $missing", ", rewrite()", ", 'message'"} {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ReflectionCompositeTypeAssumedNamed"}, Flow: flow.Options{Assertions: enabled}})
			if err != nil {
				t.Fatal(err)
			}
			src := fmt.Sprintf(`<?php function consume(int|string $x){} function f(){ $r=new ReflectionFunction('consume');$t=$r->getParameters()[0]->getType();assert($t instanceof ReflectionNamedType%s);$t->getName();}`, description)
			got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
			want := 1
			if enabled && (description == "" || description == ", 'message'") {
				want = 0
			}
			if len(got) != want {
				t.Fatalf("enabled=%v description=%s findings=%v", enabled, description, got)
			}
		}
	}
}
