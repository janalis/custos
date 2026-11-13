package analysis

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/names"
)

// FailNamesForTest injects a shared-data failure for integration tests.
func FailNamesForTest() func() {
	old := newResolver
	newResolver = func(*syntax.File) *names.Resolver { panic("injected resolver failure") }
	return func() { newResolver = old }
}
