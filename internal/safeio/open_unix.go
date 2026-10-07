//go:build !windows

package safeio

import (
	"os"
	"syscall"
)

// open opens path for reading and returns it with its size, refusing
// anything but a regular file. The path is checked before opening, so a
// device is never opened (opening one can have side effects: a tape
// rewinds, a serial line toggles, a terminal becomes the controlling tty).
// O_NONBLOCK (a FIFO swapped in after the check does not block the open)
// and O_NOCTTY cover a path replaced between the check and the open, and
// the type is checked again on the opened descriptor.
func open(path string) (*os.File, int64, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}
	if !st.Mode().IsRegular() {
		return nil, 0, notRegular(path)
	}
	if beforeOpen != nil {
		beforeOpen(path)
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, 0, err
	}
	if st, err = f.Stat(); err != nil || !st.Mode().IsRegular() {
		f.Close()
		return nil, 0, notRegular(path)
	}
	return f, st.Size(), nil
}

// beforeOpen, when set (tests only), runs between the check and the open so
// a test can swap the path the way a concurrent writer could.
var beforeOpen func(path string)
