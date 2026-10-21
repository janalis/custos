//go:build !unix

package testbudget

import "time"

// started anchors CPU where the process CPU time is not available.
var started = time.Now()

// CPU is the wall-clock time since the package was loaded where the
// process CPU time is not available.
func CPU() time.Duration { return time.Since(started) }
