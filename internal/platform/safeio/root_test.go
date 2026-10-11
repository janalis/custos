package safeio

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRootFile(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := root.WriteFile("file", []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := ReadRootFile(root, "file", 6)
	if err != nil || string(b) != "abcdef" {
		t.Fatal(string(b), err)
	}
	for _, tc := range []struct {
		name string
		max  int64
	}{{"file", 5}, {"missing", 10}, {".", 10}, {"../outside", 10}} {
		if _, err := ReadRootFile(root, tc.name, tc.max); err == nil {
			t.Fatal(tc)
		}
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRootFile(root, "link", 100); err == nil {
		t.Fatal("symlink escape")
	}
	beforeRootOpen = func(root *os.Root, name string) { _ = root.Remove(name); _ = root.Mkdir(name, 0o700) }
	t.Cleanup(func() { beforeRootOpen = nil })
	if _, err := ReadRootFile(root, "file", 100); err == nil {
		t.Fatal("swapped directory accepted")
	}
	beforeRootOpen = func(root *os.Root, name string) { _ = root.Remove(name) }
	if err := root.WriteFile("other", []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRootFile(root, "other", 100); err == nil {
		t.Fatal("removed file accepted")
	}
}

type simulatedRootFile struct {
	io.Reader
	info os.FileInfo
	err  error
}

func (f simulatedRootFile) Stat() (os.FileInfo, error) { return f.info, f.err }
func TestOpenedRootDescriptorFailures(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "file")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := readRootDescriptor(simulatedRootFile{Reader: strings.NewReader("abc"), info: info}, 2); err == nil {
		t.Fatal("growing file exceeded budget")
	}
	if _, err := readRootDescriptor(simulatedRootFile{err: errors.New("stat failed")}, 2); err == nil {
		t.Fatal("descriptor stat failed")
	}
}
