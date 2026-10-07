//go:build !windows

package safeio

import (
	"os"
	"syscall"
)

// open opens path for reading and returns it with its size, refusing
// anything but a regular file. O_NONBLOCK makes opening a FIFO return at
// once instead of waiting for a writer; the type check then runs on the
// opened descriptor (fstat), so a path swapped concurrently cannot slip a
// device or FIFO through, and no separate path lookup (stat) is needed.
func open(path string) (*os.File, int64, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, 0, err
	}
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() {
		f.Close()
		return nil, 0, notRegular(path)
	}
	return f, st.Size(), nil
}
