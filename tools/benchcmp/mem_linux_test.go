//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGroupMemory(t *testing.T) {
	root := t.TempDir()
	// pid (comm) state ppid pgrp, 18 more fields, then rss.
	stat := func(pgrp, rss string) string {
		return "1 (a b) S 1 " + pgrp + strings.Repeat(" 0", 18) + " " + rss + " 0\n"
	}
	write(t, filepath.Join(root, "self", "stat"), stat("7", "1")) // not a pid
	write(t, filepath.Join(root, "10", "stat"), stat("7", "2"))   // resident set: 2 pages
	write(t, filepath.Join(root, "11", "stat"), stat("7", "9"))   // proportional set: 3 kB
	write(t, filepath.Join(root, "11", "smaps_rollup"), "Rss: 9 kB\nPss: 3 kB\n")
	write(t, filepath.Join(root, "12", "stat"), stat("8", "5"))   // other group
	write(t, filepath.Join(root, "13", "stat"), "13 (short) S\n") // truncated
	write(t, filepath.Join(root, "14", "stat"), stat("7", "1"))   // no Pss line: 1 page
	write(t, filepath.Join(root, "14", "smaps_rollup"), "Rss: 9 kB\n")
	write(t, filepath.Join(root, "15", "stat"), stat("7", "0")) // bad Pss: resident set
	write(t, filepath.Join(root, "15", "smaps_rollup"), "Pss: many kB\n")
	if err := os.Mkdir(filepath.Join(root, "16"), 0o755); err != nil { // exited
		t.Fatal(err)
	}
	page := uint64(os.Getpagesize())
	if got, want := groupMemory(root, 7), 3*page+3*1024; got != want {
		t.Errorf("groupMemory = %d, want %d", got, want)
	}
}

func TestWatch(t *testing.T) {
	exe, _ := os.Executable()
	s := execute([]string{exe}, t.TempDir(), fakeEnv("FAKE_SLEEP=200ms", "FAKE_CHILD=1"), time.Minute)
	if s.peak < 32<<20 {
		t.Errorf("peak = %d, want the fake's 32 MB at least", s.peak)
	}
	if got := maxRSS(noUsage{}); got != 0 {
		t.Errorf("maxRSS = %d", got)
	}
}

func TestLinuxMachine(t *testing.T) {
	saved := procRoot
	t.Cleanup(func() { procRoot = saved })
	procRoot = t.TempDir()
	if m := machine(); m.CPU != "unknown" || m.MemoryGB != 0 {
		t.Errorf("machine without procfs = %+v", m)
	}
	write(t, filepath.Join(procRoot, "cpuinfo"), "processor\t: 0\nmodel name\t: Test CPU\nmodel name\t: Other\n")
	write(t, filepath.Join(procRoot, "meminfo"), "MemTotal:       16777216 kB\nMemFree: 1 kB\n")
	if m := machine(); m.CPU != "Test CPU" || m.MemoryGB != 16 {
		t.Errorf("machine = %+v", m)
	}
}
