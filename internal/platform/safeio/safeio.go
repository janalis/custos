// Package safeio reads files named by untrusted input (repository content,
// configuration) without letting them exhaust memory or block: only regular
// files are read (not devices, FIFOs or directories, also behind symlinks),
// and never more than a byte limit.
package safeio

import (
	"bytes"
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
	f, size, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Size the buffer from the opened file (one allocation for the common
	// case); a file that grows meanwhile is still read up to max.
	if size > max {
		size = max
	}
	// bytes.Buffer.ReadFrom keeps MinRead bytes free before each read, so
	// that much headroom avoids any reallocation for an unchanged file.
	buf := bytes.NewBuffer(make([]byte, 0, size+bytes.MinRead))
	_, err = buf.ReadFrom(io.LimitReader(f, max))
	return buf.Bytes(), err
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

// notRegular is the error for a path that is not a regular file.
func notRegular(path string) error {
	return &os.PathError{Op: "read", Path: path, Err: ErrNotRegular}
}
