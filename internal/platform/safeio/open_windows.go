//go:build windows

package safeio

import "os"

// open opens path for reading and returns it with its size, refusing
// anything but a regular file (checked before opening, so a device is never
// opened, and again on the opened handle in case the path was swapped).
func open(path string) (*os.File, int64, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}
	if !st.Mode().IsRegular() {
		return nil, 0, notRegular(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	if st, err = f.Stat(); err != nil || !st.Mode().IsRegular() {
		f.Close()
		return nil, 0, notRegular(path)
	}
	return f, st.Size(), nil
}
