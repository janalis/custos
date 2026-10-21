//go:build unix

package testbudget

import (
	"syscall"
	"time"
)

// CPU returns the CPU time (user + system) the process has used so far.
// Unlike wall-clock time, it does not grow while other processes hold the
// CPU, so a limit on compute-bound work holds on a loaded machine; time
// such work as `start := testbudget.CPU(); …; testbudget.CPU() - start`,
// with no other test of the package running in parallel.
func CPU() time.Duration {
	var ru syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &ru) // cannot fail for RUSAGE_SELF
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}
