//go:build unix

package main

import (
	"math"
	"runtime"
	"syscall"
)

func baseMachine() Machine {
	return Machine{OS: runtime.GOOS + "/" + runtime.GOARCH, CPU: "unknown", Cores: runtime.NumCPU()}
}

func roundGB(bytes float64) float64 { return math.Round(bytes/(1<<30)*10) / 10 }

// rusageMax is the ru_maxrss of a finished process (0 when unknown).
func rusageMax(ps interface{ SysUsage() any }) uint64 {
	if ru, ok := ps.SysUsage().(*syscall.Rusage); ok {
		return uint64(ru.Maxrss)
	}
	return 0
}
