//go:build !windows

package safeio

import (
	"os"
	"syscall"
)

// PrivateDirectory requires owner-only access and ownership by the server user.
func PrivateDirectory(_ string, info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return info.IsDir() && info.Mode()&os.ModeSymlink == 0 && info.Mode().Perm()&0o077 == 0 && ok && stat.Uid == uint32(os.Geteuid())
}
