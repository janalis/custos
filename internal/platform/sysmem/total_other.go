//go:build !darwin && !linux

package sysmem

// Total returns 0: the memory of this operating system is not read.
func Total() uint64 { return 0 }
