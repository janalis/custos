package cli

import (
	"math"
	"os"
	"runtime/debug"

	"github.com/janalis/custos/internal/platform/sysmem"
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
// batch command runs with, 0 meaning to leave that setting alone. The soft
// limit, three quarters of the memory the process may use (total), makes
// the collector work harder as the heap nears it, so the higher percent
// cannot push the peak past what the machine or container has; when that
// memory is unknown, the GC keeps the runtime defaults. A GOGC in the
// environment, whatever its value (off included), is the user's choice of
// collector and leaves both settings alone; a GOMEMLIMIT alone replaces
// only the soft limit.
func planGC(getenv func(string) string, total func() uint64) (percent int, limit int64) {
	if getenv("GOGC") != "" {
		return 0, 0
	}
	if getenv("GOMEMLIMIT") != "" {
		return gcPercent, 0
	}
	l := total() / 4 * 3
	if l == 0 || l > math.MaxInt64 {
		return 0, 0
	}
	return gcPercent, int64(l)
}
