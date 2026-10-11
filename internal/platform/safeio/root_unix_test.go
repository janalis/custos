//go:build !windows

package safeio

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRootFIFOAndPrivacy(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !PrivateDirectory(dir, st) {
		t.Fatal("private directory rejected")
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err = os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if PrivateDirectory(dir, st) {
		t.Fatal("public directory accepted")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := syscall.Mkfifo(filepath.Join(dir, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRootFile(root, "pipe", 100); err == nil {
		t.Fatal("FIFO accepted")
	}
	if err := root.WriteFile("regular", []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	beforeRootOpen = func(root *os.Root, name string) {
		_ = root.Remove(name)
		if err := syscall.Mkfifo(filepath.Join(dir, name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { beforeRootOpen = nil })
	if _, err := ReadRootFile(root, "regular", 100); err == nil {
		t.Fatal("swapped FIFO accepted")
	}
}
