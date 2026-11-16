package analysis

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// FailNamesForTest injects a shared-data failure for integration tests.
func FailNamesForTest() func() {
	old := newResolver
	newResolver = func(*syntax.File) *names.Resolver { panic("injected resolver failure") }
	return func() { newResolver = old }
}

// FailFlowForTest injects a failure of the flow environment's build.
func FailFlowForTest() func() {
	old := newLocalFlow
	newLocalFlow = func(*syntax.File, *infer.Env, *flow.Snapshot, flow.Options, func(func())) *flow.Env {
		panic("injected flow failure")
	}
	return func() { newLocalFlow = old }
}
