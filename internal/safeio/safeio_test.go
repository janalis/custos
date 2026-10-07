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

// TestSwappedPath swaps the path between the Stat and the Open, as a
// concurrent writer could: a vanished file fails the Open, a file replaced
// by a directory fails the re-check on the opened descriptor.
func TestSwappedPath(t *testing.T) {
	defer func() { beforeOpen = nil }()
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	write := func() {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := ReadPrefix(filepath.Join(dir, "missing"), 10); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing: %v", err)
	}

	write()
	beforeOpen = func(path string) { os.Remove(path) }
	if _, err := ReadFile(p, 10); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("vanished: %v", err)
	}

	write()
	beforeOpen = func(path string) {
		os.Remove(path)
		os.Mkdir(path, 0o755)
	}
	if _, err := ReadFile(p, 10); !errors.Is(err, ErrNotRegular) {
		t.Fatalf("swapped for a directory: %v", err)
	}
}
