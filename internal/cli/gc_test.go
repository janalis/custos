package cli

import (
	"io"
	"math"
	"runtime/debug"
	"strings"
	"testing"

	"custos/internal/platform/sysmem"
)

func TestPlanGC(t *testing.T) {
	const gib = uint64(1) << 30
	for name, tc := range map[string]struct {
		env        map[string]string
		total      uint64
		percent    int
		limit      int64
		readsTotal bool
	}{
		"defaults":           {nil, 16 * gib, gcPercent, int64(12 * gib), true},
		"GOGC set":           {map[string]string{"GOGC": "50"}, 16 * gib, 0, 0, false},
		"GOGC off":           {map[string]string{"GOGC": "off"}, 16 * gib, 0, 0, false},
		"GOMEMLIMIT set":     {map[string]string{"GOMEMLIMIT": "1GiB"}, 16 * gib, gcPercent, 0, false},
		"both set":           {map[string]string{"GOGC": "100", "GOMEMLIMIT": "1GiB"}, 16 * gib, 0, 0, false},
		"memory unknown":     {nil, 0, 0, 0, true},
		"cgroup sized":       {nil, 2 * gib, gcPercent, int64(3 * gib / 2), true},
		"beyond int64 range": {nil, math.MaxUint64, 0, 0, true},
	} {
		read := false
		getenv := func(k string) string { return tc.env[k] }
		total := func() uint64 { read = true; return tc.total }
		percent, limit := planGC(getenv, total)
		if percent != tc.percent || limit != tc.limit {
			t.Errorf("%s: planGC = (%d, %d), want (%d, %d)", name, percent, limit, tc.percent, tc.limit)
		}
		if read != tc.readsTotal {
			t.Errorf("%s: total read = %v, want %v", name, read, tc.readsTotal)
		}
	}
}

// TestTuneGCCommands checks that only the batch commands tune the GC.
func TestTuneGCCommands(t *testing.T) {
	calls := 0
	saved := tuneGC
	tuneGC = func() { calls++ }
	defer func() { tuneGC = saved }()
	for _, tc := range []struct {
		args  []string
		tunes bool
	}{
		{[]string{"analyse", "-h"}, true},
		{[]string{"analyze", "-h"}, true},
		{[]string{"fix", "-h"}, true},
		{[]string{"lsp", "-h"}, false},
		{[]string{"mcp", "-h"}, false},
		{[]string{"rules", "-h"}, false},
		{[]string{"version"}, false},
	} {
		before := calls
		Run("dev", tc.args, strings.NewReader(""), io.Discard, io.Discard)
		if got := calls > before; got != tc.tunes {
			t.Errorf("%v: tuned GC = %v, want %v", tc.args, got, tc.tunes)
		}
	}
}

// TestTuneRuntimeGC applies the plan to the test binary and restores the
// previous settings.
func TestTuneRuntimeGC(t *testing.T) {
	prevPercent := debug.SetGCPercent(100)
	debug.SetGCPercent(prevPercent)
	prevLimit := debug.SetMemoryLimit(-1)
	t.Cleanup(func() {
		debug.SetGCPercent(prevPercent)
		debug.SetMemoryLimit(prevLimit)
	})

	t.Setenv("GOGC", "")
	t.Setenv("GOMEMLIMIT", "")
	tuneRuntimeGC()
	wantPercent := gcPercent
	if sysmem.Total() == 0 {
		wantPercent = prevPercent // memory unknown: runtime defaults
	}
	if got := debug.SetGCPercent(prevPercent); got != wantPercent {
		t.Errorf("GC percent = %d, want %d", got, wantPercent)
	}
	if got := debug.SetMemoryLimit(-1); sysmem.Total() > 0 && got != int64(sysmem.Total()/4*3) {
		t.Errorf("memory limit = %d, want 3/4 of %d", got, sysmem.Total())
	}

	// Set in the environment, both are left alone.
	debug.SetMemoryLimit(math.MaxInt64)
	t.Setenv("GOGC", "100")
	t.Setenv("GOMEMLIMIT", "1GiB")
	tuneRuntimeGC()
	if got := debug.SetGCPercent(prevPercent); got != prevPercent {
		t.Errorf("GC percent = %d, want it left at %d", got, prevPercent)
	}
	if got := debug.SetMemoryLimit(-1); got != math.MaxInt64 {
		t.Errorf("memory limit = %d, want it left alone", got)
	}
}
