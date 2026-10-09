package project

import (
	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
)

// ReadSource reads a source file, but never more than syntax.MaxFileSize+1
// bytes: a larger (hostile or sparse) file yields a prefix that
// syntax.Parse then reports as too large, instead of being loaded whole
// into memory. Only regular files are read (a symlink to a device or FIFO
// fails instead of blocking). Callers must not write the result back as
// the file content.
func ReadSource(path string) ([]byte, error) {
	readers <- struct{}{}
	defer func() { <-readers }()
	return safeio.ReadPrefix(path, syntax.MaxFileSize+1)
}

// readers caps concurrent file reads: with one reader per core, opens
// contend in the kernel and slow the parsing workers too. On a vendor tree
// (7.7k files, 16-core M-series Mac) a full run went from 0.78–0.93 s to
// 0.74–0.77 s with 6 readers, and system CPU time from 2–4 s to 1.2 s;
// 4 readers starved the index pass.
var readers = make(chan struct{}, 6)
