package astquery

import (
	"strings"
)

// IsTestClassFQN reports whether a class FQN (with or without leading
// backslash) denotes a test class: it ends with Test or contains a \Tests\
// or \Test\ namespace segment.
func IsTestClassFQN(fqn string) bool {
	if fqn == "" {
		return false
	}
	if !strings.HasPrefix(fqn, `\`) {
		fqn = `\` + fqn
	}
	return strings.HasSuffix(fqn, "Test") || strings.Contains(fqn, `\Tests\`) || strings.Contains(fqn, `\Test\`)
}
