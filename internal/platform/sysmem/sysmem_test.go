package sysmem

import (
	"errors"
	"runtime"
	"slices"
	"testing"
	"testing/fstest"
)

func TestTotal(t *testing.T) {
	got := Total()
	switch runtime.GOOS {
	case "darwin", "linux":
		if got == 0 {
			t.Fatalf("Total() = 0 on %s", runtime.GOOS)
		}
	default:
		if got != 0 {
			t.Fatalf("Total() = %d, want 0 on %s", got, runtime.GOOS)
		}
	}
}

func TestFromSysctl(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		err  error
		want uint64
	}{
		// 16 GiB: the trailing zero byte syscall.Sysctl drops is restored.
		{"\x00\x00\x00\x00\x04\x00\x00", nil, 16 << 30},
		{"\x01\x02\x03\x04\x05\x06\x07\x08", nil, 0x0807060504030201},
		{"", nil, 0},
		{"\x01\x02\x03\x04\x05\x06\x07\x08\x09", nil, 0},
		{"\x01", errors.New("no such sysctl"), 0},
	} {
		if got := fromSysctl(tc.raw, tc.err); got != tc.want {
			t.Errorf("fromSysctl(%q, %v) = %d, want %d", tc.raw, tc.err, got, tc.want)
		}
	}
}

func TestMeminfoTotal(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want uint64
	}{
		{"MemFree:  100 kB\nMemTotal:       16384 kB\n", 16384 * 1024},
		{"MemTotal: 2 MB\nMemTotal: 3 kB\n", 3 * 1024},
		{"MemTotal: lots kB\n", 0},
		{"MemTotal: 18446744073709551615 kB\n", 0},
		{"", 0},
	} {
		if got := meminfoTotal(tc.in); got != tc.want {
			t.Errorf("meminfoTotal(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseLimit(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want uint64
	}{
		{"1073741824\n", 1 << 30},
		{"max\n", 0},
		{"", 0},
		{"garbage", 0},
		{"-1", 0},
		{"0", 0},
		{"9223372036854771712", 0}, // v1 "unlimited", page-aligned near MaxInt64
		{"4611686018427387904", 0}, // 1<<62: no real limit
		{"4611686018427387903", 1<<62 - 1},
	} {
		if got := parseLimit(tc.in); got != tc.want {
			t.Errorf("parseLimit(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestCgroupFiles(t *testing.T) {
	self := "12:cpu,memory:/docker/abc\n0::/user.slice/app.scope\n3:pids:/x\nnonsense\n4:memory:relative\n"
	want := []string{
		"sys/fs/cgroup/memory/docker/abc/memory.limit_in_bytes",
		"sys/fs/cgroup/memory/docker/memory.limit_in_bytes",
		"sys/fs/cgroup/memory/memory.limit_in_bytes",
		"sys/fs/cgroup/memory/docker/abc/memory.stat",
		"sys/fs/cgroup/user.slice/app.scope/memory.max",
		"sys/fs/cgroup/user.slice/memory.max",
		"sys/fs/cgroup/memory.max",
	}
	if got := cgroupFiles(self); !slices.Equal(got, want) {
		t.Fatalf("cgroupFiles = %q, want %q", got, want)
	}
	// The root group, or no group line, reads the roots of both hierarchies.
	want = []string{"sys/fs/cgroup/memory.max", "sys/fs/cgroup/memory/memory.limit_in_bytes"}
	if got := cgroupFiles("0::/\n"); !slices.Equal(got, want) {
		t.Fatalf("cgroupFiles(root) = %q, want %q", got, want)
	}
	if got := cgroupFiles(""); !slices.Equal(got, want) {
		t.Fatalf("cgroupFiles(none) = %q, want %q", got, want)
	}
}

func TestStatLimit(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want uint64
	}{
		{"cache 1\nhierarchical_memory_limit 1073741824\nrss 2\n", 1 << 30},
		{"hierarchical_memory_limit 9223372036854771712\n", 0},
		{"hierarchical_memory_limit\n", 0},
		{"rss 2\n", 0},
	} {
		if got := statLimit(tc.in); got != tc.want {
			t.Errorf("statLimit(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFromLinux(t *testing.T) {
	meminfo := &fstest.MapFile{Data: []byte("MemTotal: 8388608 kB\n")} // 8 GiB
	file := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	for name, tc := range map[string]struct {
		fs   fstest.MapFS
		want uint64
	}{
		"no cgroup": {fstest.MapFS{"proc/meminfo": meminfo}, 8 << 30},
		"v2 unlimited": {fstest.MapFS{
			"proc/meminfo":             meminfo,
			"proc/self/cgroup":         file("0::/\n"),
			"sys/fs/cgroup/memory.max": file("max\n"),
		}, 8 << 30},
		"v2 own group": {fstest.MapFS{
			"proc/meminfo":                    meminfo,
			"proc/self/cgroup":                file("0::/app\n"),
			"sys/fs/cgroup/app/memory.max":    file("2147483648\n"),
			"sys/fs/cgroup/memory.max":        file("max\n"),
			"sys/fs/cgroup/memory/whatever.x": file("1"),
		}, 2 << 30},
		"v1 unlimited": {fstest.MapFS{
			"proc/meminfo": meminfo,
			"sys/fs/cgroup/memory/memory.limit_in_bytes": file("9223372036854771712\n"),
		}, 8 << 30},
		"v1 limited": {fstest.MapFS{
			"proc/meminfo": meminfo,
			"sys/fs/cgroup/memory/memory.limit_in_bytes": file("1073741824\n"),
		}, 1 << 30},
		"garbage limit": {fstest.MapFS{
			"proc/meminfo":             meminfo,
			"sys/fs/cgroup/memory.max": file("lots\n"),
		}, 8 << 30},
		"limit without meminfo": {fstest.MapFS{
			"sys/fs/cgroup/memory.max": file("1073741824\n"),
		}, 1 << 30},
		"nothing": {fstest.MapFS{}, 0},
		"v2 parent slice": {fstest.MapFS{
			"proc/meminfo":     meminfo,
			"proc/self/cgroup": file("0::/user.slice/user-1000.slice/session-2.scope\n"),
			"sys/fs/cgroup/user.slice/user-1000.slice/session-2.scope/memory.max": file("max\n"),
			"sys/fs/cgroup/user.slice/user-1000.slice/memory.max":                 file("1073741824\n"),
			"sys/fs/cgroup/user.slice/memory.max":                                 file("4294967296\n"),
		}, 1 << 30},
		"v1 parent group": {fstest.MapFS{
			"proc/meminfo":     meminfo,
			"proc/self/cgroup": file("5:memory:/batch/job\n"),
			"sys/fs/cgroup/memory/batch/job/memory.limit_in_bytes": file("9223372036854771712\n"),
			"sys/fs/cgroup/memory/batch/memory.limit_in_bytes":     file("2147483648\n"),
		}, 2 << 30},
		"v1 hierarchical limit": {fstest.MapFS{
			"proc/meminfo":                         meminfo,
			"proc/self/cgroup":                     file("5:memory:/job\n"),
			"sys/fs/cgroup/memory/job/memory.stat": file("hierarchical_memory_limit 1073741824\n"),
		}, 1 << 30},
		"v1 unlimited without meminfo": {fstest.MapFS{
			"sys/fs/cgroup/memory/memory.limit_in_bytes": file("9223372036854771712\n"),
		}, 0},
	} {
		if got := fromLinux(tc.fs); got != tc.want {
			t.Errorf("%s: fromLinux = %d, want %d", name, got, tc.want)
		}
	}
}
