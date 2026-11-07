//go:build linux

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// procRoot is the procfs mount point.
var procRoot = "/proc"

// pollEvery is the memory sampling interval.
const pollEvery = 20 * time.Millisecond

// watch samples the memory of process group pgid until stop is closed and
// returns its peak in bytes. PHP tools run parallel workers, so the sum over
// the group is what the machine has to provide.
func watch(pgid int, stop <-chan struct{}) uint64 {
	tick := time.NewTicker(pollEvery)
	defer tick.Stop()
	var peak uint64
	for {
		peak = max(peak, groupMemory(procRoot, pgid))
		select {
		case <-stop:
			return peak
		case <-tick.C:
		}
	}
}

// groupMemory sums the proportional set size of the processes of group
// pgid: pages shared by forked workers count once overall, not once per
// worker. It falls back to the resident set when smaps_rollup is missing.
func groupMemory(root string, pgid int) uint64 {
	entries, _ := os.ReadDir(root)
	var total uint64
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		dir := filepath.Join(root, e.Name())
		data, err := os.ReadFile(filepath.Join(dir, "stat"))
		if err != nil {
			continue // exited meanwhile
		}
		// Fields after the parenthesized command name, which may hold
		// spaces: state ppid pgrp ... with rss (pages) the 22nd.
		i := strings.LastIndexByte(string(data), ')')
		f := strings.Fields(string(data)[i+1:])
		if len(f) < 22 || f[2] != strconv.Itoa(pgid) {
			continue
		}
		if pss, ok := pssBytes(filepath.Join(dir, "smaps_rollup")); ok {
			total += pss
			continue
		}
		pages, _ := strconv.ParseUint(f[21], 10, 64)
		total += pages * uint64(os.Getpagesize())
	}
	return total
}

// pssBytes reads the "Pss:" line (kilobytes) of an smaps_rollup file.
func pssBytes(path string) (uint64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if f := strings.Fields(line); len(f) >= 2 && f[0] == "Pss:" {
			kb, err := strconv.ParseUint(f[1], 10, 64)
			return kb * 1024, err == nil
		}
	}
	return 0, false
}

// maxRSS returns the peak resident set of the process and its waited-for
// descendants (kilobytes on Linux).
func maxRSS(ps interface{ SysUsage() any }) uint64 {
	return rusageMax(ps) * 1024
}

// machine describes the host from /proc.
func machine() Machine {
	m := baseMachine()
	if f, err := os.Open(filepath.Join(procRoot, "cpuinfo")); err == nil {
		defer f.Close()
		for s := bufio.NewScanner(f); s.Scan(); {
			if k, v, ok := strings.Cut(s.Text(), ":"); ok && strings.TrimSpace(k) == "model name" {
				m.CPU = strings.TrimSpace(v)
				break
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(procRoot, "meminfo")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if f := strings.Fields(line); len(f) >= 2 && f[0] == "MemTotal:" {
				kb, _ := strconv.ParseFloat(f[1], 64)
				m.MemoryGB = roundGB(kb * 1024)
			}
		}
	}
	return m
}
