package util

import (
	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/stubs"
)

// IsBuiltinClass reports whether c is the declaration from the embedded PHP
// stubs (as opposed to a user/project declaration).
func IsBuiltinClass(c *index.Class, ver phpver.Version) bool {
	if c == nil {
		return false
	}
	return stubs.Index().Class(c.FQN, ver) == c
}
