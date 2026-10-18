package syntax

// Superglobal reports whether name (without '$') is one of PHP's
// superglobal arrays, visible in every scope without a `global` statement.
func Superglobal(name string) bool {
	switch name {
	case "GLOBALS", "_SERVER", "_GET", "_POST", "_FILES", "_COOKIE", "_SESSION", "_REQUEST", "_ENV":
		return true
	}
	return false
}
