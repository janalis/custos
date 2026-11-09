//go:build !windows

package selfupdate

import "os"

// swap moves the new binary tmp over exe; a running binary may be replaced.
func swap(tmp, exe string) error { return os.Rename(tmp, exe) }
