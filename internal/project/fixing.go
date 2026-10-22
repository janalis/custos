package project

import (
	"context"
	"errors"
	"fmt"
	"os"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// PreparedFix is the result of fixing one file; out is nil when nothing
// changed.
type PreparedFix struct {
	Source, Output []byte
	Perm           os.FileMode
	Applied        int
	Err            error
}

// PrepareFixes fixes every file with one worker per CPU (files are independent;
// the engine is safe for concurrent use). sources holds bytes already read
// for the first files (nil where unread); the rest are read here. Outcomes
// are in input order so diffs and messages stay deterministic.
func PrepareFixes(engine *analysis.Engine, files []string, sources [][]byte, parse syntax.Options) []PreparedFix {
	outs := make([]PreparedFix, len(files))
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		var src []byte
		if i < len(sources) {
			src = sources[i]
		}
		src, perm, err := readFixable(files[i], src)
		if err != nil {
			outs[i].Err = err
			return
		}
		res := fixSource(engine, files[i], src, fixing.Options{Parse: parse})
		if res.Applied == 0 || string(res.Source) == string(src) {
			return
		}
		outs[i] = PreparedFix{Source: src, Output: res.Source, Perm: perm, Applied: res.Applied}
	}, func(i int, r any) {
		outs[i] = PreparedFix{Err: fmt.Errorf("%s: internal failure, left unchanged: %v", files[i], r)}
	})
	return outs
}

// fixSource is fixing.FixSource; tests replace it to simulate a crash.
var fixSource = fixing.FixSource

// ErrSkipped marks files fix leaves alone without failing.
var ErrSkipped = errors.New("not a regular file")

// readFixable reads a file fix may rewrite, unless src already holds its
// bytes, and returns its permission bits.
// Symlinks (and other non-regular files) are skipped, in dry runs too:
// writing through a link may lead outside the project, and the analysed
// repository is untrusted.
func readFixable(path string, src []byte) ([]byte, os.FileMode, error) {
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		err = fmt.Errorf("%s: %w", path, ErrSkipped)
	}
	if err != nil {
		return nil, 0, err
	}
	if src == nil {
		src, err = ReadSource(path)
	}
	return src, info.Mode().Perm(), err
}
