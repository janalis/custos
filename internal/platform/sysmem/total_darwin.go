package sysmem

import "syscall"

// Total returns the machine's physical memory in bytes, or 0 when unknown.
func Total() uint64 { return fromSysctl(syscall.Sysctl("hw.memsize")) }
