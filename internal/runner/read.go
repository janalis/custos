package runner

import (
	"custos/internal/safeio"
	"custos/internal/syntax"
)

// ReadSource reads a source file, but never more than syntax.MaxFileSize+1
// bytes: a larger (hostile or sparse) file yields a prefix that
// syntax.Parse then reports as too large, instead of being loaded whole
// into memory. Only regular files are read (a symlink to a device or FIFO
// fails instead of blocking). Callers must not write the result back as
// the file content.
func ReadSource(path string) ([]byte, error) {
	return safeio.ReadPrefix(path, syntax.MaxFileSize+1)
}
