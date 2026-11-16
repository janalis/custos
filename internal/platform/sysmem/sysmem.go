// Package sysmem tells how much memory the running process may use: the
// machine's physical memory, capped on Linux by the cgroup memory limit of
// the process when one is set. Total is defined per operating system and
// returns 0 when the amount is unknown; the decoding below is portable so
// that every platform's tests cover it.
package sysmem

import (
	"encoding/binary"
	"io/fs"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
)

// fromSysctl decodes the macOS hw.memsize value. syscall.Sysctl returns the
// raw little-endian uint64 as a string with one trailing zero byte dropped,
// so a short value is padded back with zeros.
func fromSysctl(raw string, err error) uint64 {
	if err != nil || raw == "" || len(raw) > 8 {
		return 0
	}
	var b [8]byte
	copy(b[:], raw)
	return binary.LittleEndian.Uint64(b[:])
}

// fromLinux reads the physical memory from proc/meminfo in fsys, the root
// of a Linux filesystem, and lowers it to the smallest cgroup memory limit
// that applies to the process.
func fromLinux(fsys fs.FS) uint64 {
	total := meminfoTotal(readText(fsys, "proc/meminfo"))
	for _, name := range cgroupFiles(readText(fsys, "proc/self/cgroup")) {
		read := parseLimit
		if path.Base(name) == "memory.stat" {
			read = statLimit
		}
		if limit := read(readText(fsys, name)); limit > 0 && (total == 0 || limit < total) {
			total = limit
		}
	}
	return total
}

func readText(fsys fs.FS, name string) string {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return ""
	}
	return string(b)
}

// meminfoTotal returns the MemTotal line of /proc/meminfo in bytes, or 0.
func meminfoTotal(meminfo string) uint64 {
	for line := range strings.Lines(meminfo) {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != "MemTotal:" || fields[2] != "kB" {
			continue
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || kb > math.MaxUint64/1024 {
			return 0
		}
		return kb * 1024
	}
	return 0
}

// cgroupFiles lists the memory limit files that may apply to the process,
// read from /proc/self/cgroup: a group is bounded by its own limit and by
// that of every ancestor, so each group's file is listed up to the root of
// its hierarchy (cgroup v2's memory.max for the "0::path" line, v1's
// memory.limit_in_bytes for a hierarchy holding the memory controller, plus
// the v1 group's memory.stat, whose hierarchical_memory_limit is that
// minimum as the kernel computes it). The roots of both hierarchies are
// always listed: that is where a container usually sees its own limit.
func cgroupFiles(self string) []string {
	var files []string
	for line := range strings.Lines(self) {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 3)
		if len(parts) != 3 || !strings.HasPrefix(parts[2], "/") {
			continue
		}
		switch {
		case parts[0] == "0" && parts[1] == "":
			files = appendAncestors(files, "sys/fs/cgroup", parts[2], "memory.max")
		case slices.Contains(strings.Split(parts[1], ","), "memory"):
			files = appendAncestors(files, "sys/fs/cgroup/memory", parts[2], "memory.limit_in_bytes")
			files = append(files, path.Join("sys/fs/cgroup/memory", parts[2], "memory.stat"))
		}
	}
	for _, root := range []string{"sys/fs/cgroup/memory.max", "sys/fs/cgroup/memory/memory.limit_in_bytes"} {
		if !slices.Contains(files, root) {
			files = append(files, root)
		}
	}
	return files
}

// appendAncestors appends the file named base of group, an absolute path
// in the hierarchy mounted at mount, then that of each ancestor up to the
// root.
func appendAncestors(files []string, mount, group, base string) []string {
	for {
		files = append(files, path.Join(mount, group, base))
		if group == "/" {
			return files
		}
		group = path.Dir(group)
	}
}

// statLimit returns the hierarchical_memory_limit of a cgroup v1
// memory.stat file in bytes, or 0 when it is missing or no real limit.
func statLimit(stat string) uint64 {
	for line := range strings.Lines(stat) {
		if v, ok := strings.CutPrefix(line, "hierarchical_memory_limit "); ok {
			return parseLimit(v)
		}
	}
	return 0
}

// noLimit is the smallest value read as no limit at all: cgroup v1 writes
// "unlimited" as the page-aligned maximum near MaxInt64, and no machine has
// 4 EiB of memory.
const noLimit = 1 << 62

// parseLimit reads a cgroup memory limit in bytes; "max" (no limit), the v1
// "unlimited" value, an empty file and anything that is not a positive
// integer give 0.
func parseLimit(s string) uint64 {
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil || n >= noLimit {
		return 0
	}
	return n
}
