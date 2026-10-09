package infer

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
)

func TestUnattachedHookValueType(t *testing.T) {
	f := syntax.Parse("empty.php", []byte("<?php"), syntax.Options{Version: phpversion.PHP84})
	e := NewEnv(f, names.New(f), index.New(nil), phpversion.PHP84)
	if got := e.hookValueType(&syntax.PropertyHook{}); !got.IsUnknown() {
		t.Fatalf("unattached hook: %s", got)
	}
}

func TestPlainParameterHookDocSpansDoNotAllocate(t *testing.T) {
	scope := &syntax.Method{Params: []*syntax.Param{{}, {}}}
	if allocs := testing.AllocsPerRun(100, func() {
		if got := promotedHookDocSpans(scope); got != nil {
			t.Fatalf("ordinary parameters have hook spans: %v", got)
		}
	}); allocs != 0 {
		t.Fatalf("ordinary parameters allocated %g times", allocs)
	}
}
