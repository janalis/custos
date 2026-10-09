// Package testbudget scales wall-clock limits in tests by the speed of the
// machine at test time, so complexity-regression tests (whose broken
// behaviour costs orders of magnitude more than the limit) do not fail
// spuriously on a loaded machine.
package testbudget

import (
	"runtime"
	"time"
)

// reference is how long calibrate's workload takes on an idle development
// machine; budgets passed to Of are expressed for that speed.
const reference = 4 * time.Millisecond

// Of returns d scaled by how much slower than the reference the machine is
// right now (never less than d itself). It calibrates on every call (about
// 10 ms), so call it right after the timed work: the load then is the load
// the work ran under, not that of a quieter moment earlier in the run.
func Of(d time.Duration) time.Duration {
	return time.Duration(float64(d) * scale(calibrate(), reference))
}

// scale is the slowdown factor of measured against ref, at least 1.
func scale(measured, ref time.Duration) float64 {
	if f := float64(measured) / float64(ref); f > 1 {
		return f
	}
	return 1
}

// calibrate times a fixed CPU-bound workload (best of three runs).
func calibrate() time.Duration {
	best := time.Duration(1<<63 - 1)
	for r := 0; r < 3; r++ {
		start := time.Now()
		x := uint64(88172645463325252)
		for i := 0; i < 2_000_000; i++ {
			x ^= x << 13
			x ^= x >> 7
			x ^= x << 17
		}
		runtime.KeepAlive(x) // keep the loop from being optimised away
		if d := time.Since(start); d < best {
			best = d
		}
	}
	return best
}
