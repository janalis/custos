//go:build unix && !linux

package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// watch waits for stop: without procfs, the peak comes from maxRSS.
func watch(_ int, stop <-chan struct{}) uint64 {
	<-stop
	return 0
}

// maxRSS returns the peak resident set of the process and its waited-for
// descendants (bytes on BSD and macOS).
func maxRSS(ps interface{ SysUsage() any }) uint64 { return rusageMax(ps) }

// machine describes the host from sysctl.
func machine() Machine {
	m := baseMachine()
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		m.CPU = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		b, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		m.MemoryGB = roundGB(b)
	}
	return m
}
