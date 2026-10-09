package throwrawexception

import (
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// isBuiltinClass reports whether c is the declaration from the embedded PHP
// stubs (as opposed to a user/project declaration).
func isBuiltinClass(c *index.Class, ver phpversion.Version) bool {
	if c == nil {
		return false
	}
	return stubs.Index().Class(c.FQN, ver) == c
}
