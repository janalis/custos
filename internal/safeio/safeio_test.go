package safeio

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestReadLimits(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, err := ReadPrefix(p, 4); err != nil || string(b) != "0123" {
		t.Fatalf("prefix: %q %v", b, err)
	}
	if b, err := ReadFile(p, 10); err != nil || string(b) != "0123456789" {
		t.Fatalf("whole: %q %v", b, err)
	}
	if _, err := ReadFile(p, 9); err == nil {
		t.Fatal("oversized file accepted")
	}
	if _, err := ReadFile(dir, 100); !errors.Is(err, ErrNotRegular) {
		t.Fatalf("directory: %v", err)
	}
	// a symlink to a device is refused, not read forever
	if runtime.GOOS != "windows" {
		l := filepath.Join(dir, "zero")
		if err := os.Symlink("/dev/zero", l); err == nil {
			if _, err := ReadFile(l, 1<<20); !errors.Is(err, ErrNotRegular) {
				t.Fatalf("/dev/zero: %v", err)
			}
		}
		fifo := filepath.Join(dir, "fifo")
		if err := syscall.Mkfifo(fifo, 0o644); err == nil {
			done := make(chan error, 1)
			go func() { _, err := ReadFile(fifo, 100); done <- err }()
			select {
			case err := <-done:
				if !errors.Is(err, ErrNotRegular) {
					t.Fatalf("fifo: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("reading a FIFO blocked")
			}
		}
	}
}

func TestMissingFile(t *testing.T) {
	if _, err := ReadPrefix(filepath.Join(t.TempDir(), "missing"), 10); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing: %v", err)
	}
}

// A file larger than its buffer estimate (here: max exceeds the size) and a
// file read exactly to its size both come back whole.
func TestReadSizes(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f")
	data := make([]byte, 100000)
	for i := range data {
		data[i] = byte('a' + i%26)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, max := range []int64{100000, 1 << 30, 99999} {
		b, err := ReadPrefix(p, max)
		want := data
		if max < int64(len(data)) {
			want = data[:max]
		}
		if err != nil || string(b) != string(want) {
			t.Fatalf("max %d: %d bytes, %v", max, len(b), err)
		}
	}
}
