package cli

import (
	"math"
	"os"
	"runtime/debug"

	"custos/internal/platform/sysmem"
)

// gcPercent is the garbage collector target (GOGC) of the batch commands,
// analyse and fix: their heap is mostly live syntax trees, indexes and
// summaries, so collecting less often saves CPU time for a bounded growth of
// the peak heap. The long-running servers keep the runtime default.
const gcPercent = 200

// tuneGC is a seam: tests keep the runtime of the test binary untouched.
var tuneGC = tuneRuntimeGC

// tuneRuntimeGC applies planGC to the running program.
func tuneRuntimeGC() {
	percent, limit := planGC(os.Getenv, sysmem.Total)
	if percent > 0 {
		debug.SetGCPercent(percent)
	}
	if limit > 0 {
		debug.SetMemoryLimit(limit)
	}
}

// planGC returns the GC target percent and the soft memory limit in bytes a
// batch command runs with, 0 meaning to leave that setting alone. A GOGC or
// GOMEMLIMIT set in the environment wins over each. The soft limit, three
// quarters of the memory the process may use (total), makes the collector
// work harder as the heap nears it, so a high percent cannot push the peak
// past what the machine or container has.
func planGC(getenv func(string) string, total func() uint64) (percent int, limit int64) {
	if getenv("GOGC") == "" {
		percent = gcPercent
	}
	if getenv("GOMEMLIMIT") == "" {
		if l := total() / 4 * 3; l <= math.MaxInt64 {
			limit = int64(l)
		}
	}
	return percent, limit
}
