package util

// IsSuperglobal reports whether name (without `$`) is one of PHP's
// superglobals, which every scope sees without a declaration.
func IsSuperglobal(name string) bool {
	switch name {
	case "GLOBALS", "_SERVER", "_GET", "_POST", "_FILES", "_COOKIE", "_SESSION", "_REQUEST", "_ENV":
		return true
	}
	return false
}
