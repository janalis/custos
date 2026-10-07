// Package safeio reads files named by untrusted input (repository content,
// configuration) without letting them exhaust memory or block: only regular
// files are read (not devices, FIFOs or directories, also behind symlinks),
// and never more than a byte limit.
package safeio

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrNotRegular is returned for paths that are not regular files.
var ErrNotRegular = errors.New("not a regular file")

// ReadPrefix returns at most max bytes of the regular file at path (the
// whole file when it is smaller).
func ReadPrefix(path string, max int64) ([]byte, error) {
	f, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, max))
}

// ReadFile returns the content of the regular file at path, or an error
// when it is larger than max bytes.
func ReadFile(path string, max int64) ([]byte, error) {
	b, err := ReadPrefix(path, max+1)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("%s: larger than %d bytes", path, max)
	}
	return b, nil
}

// beforeOpen, when set (tests only), runs between the Stat and the Open so
// a test can swap the path the way a concurrent writer could.
var beforeOpen func(path string)

func open(path string) (*os.File, error) {
	// Stat first: opening a FIFO would block until a writer appears.
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, &os.PathError{Op: "read", Path: path, Err: ErrNotRegular}
	}
	if beforeOpen != nil {
		beforeOpen(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Re-check on the opened file (the path may have been swapped).
	if st, err := f.Stat(); err != nil || !st.Mode().IsRegular() {
		f.Close()
		return nil, &os.PathError{Op: "read", Path: path, Err: ErrNotRegular}
	}
	return f, nil
}
