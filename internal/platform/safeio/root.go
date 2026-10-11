package safeio

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// ReadRootFile reads a bounded regular file through a confined root handle.
// Nonblocking opens prevent a FIFO swapped in after the first check from
// hanging the server. Both path and opened descriptor must be regular files.
func ReadRootFile(root *os.Root, name string, max int64) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("invalid relative path")
	}
	st, err := root.Stat(name)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, notRegular(name)
	}
	if beforeRootOpen != nil {
		beforeRootOpen(root, name)
	}
	f, err := root.OpenFile(name, os.O_RDONLY|rootReadFlags, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readRootDescriptor(f, max)
}

type rootDescriptor interface {
	io.Reader
	Stat() (os.FileInfo, error)
}

func readRootDescriptor(f rootDescriptor, max int64) ([]byte, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	if st.Size() > max {
		return nil, fmt.Errorf("input exceeds byte budget")
	}
	buffer := bytes.NewBuffer(make([]byte, 0, st.Size()+bytes.MinRead))
	_, err = buffer.ReadFrom(io.LimitReader(f, max+1))
	b := buffer.Bytes()
	if int64(len(b)) > max {
		return nil, fmt.Errorf("input exceeds byte budget")
	}
	return b, err
}

// Test seam for deterministic path swaps between stat and open.
var beforeRootOpen func(*os.Root, string)
