package sysmem

import "os"

// Total returns the machine's physical memory in bytes, lowered to the
// process's cgroup memory limit when one is set, or 0 when unknown.
func Total() uint64 { return fromLinux(os.DirFS("/")) }
