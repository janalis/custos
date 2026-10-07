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
