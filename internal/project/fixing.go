package project

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"

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
// the engine is safe for concurrent use). Outcomes are in input order so
// diffs and messages stay deterministic.
func PrepareFixes(engine *analysis.Engine, files []string, parse syntax.Options) []PreparedFix {
	outs := make([]PreparedFix, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < runtime.GOMAXPROCS(0); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				src, perm, err := ReadFixable(files[i])
				if err != nil {
					outs[i].Err = err
					continue
				}
				res := fixing.FixSource(engine, files[i], src, fixing.Options{Parse: parse})
				if res.Applied == 0 || string(res.Source) == string(src) {
					continue
				}
				outs[i] = PreparedFix{Source: src, Output: res.Source, Perm: perm, Applied: res.Applied}
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return outs
}

// ErrSkipped marks files fix leaves alone without failing.
var ErrSkipped = errors.New("not a regular file")

// ReadFixable reads a file fix may rewrite and returns its permission bits.
// Symlinks (and other non-regular files) are skipped, in dry runs too:
// writing through a link may lead outside the project, and the analysed
// repository is untrusted.
func ReadFixable(path string) ([]byte, os.FileMode, error) {
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		err = fmt.Errorf("%s: %w", path, ErrSkipped)
	}
	if err != nil {
		return nil, 0, err
	}
	src, err := ReadSource(path)
	return src, info.Mode().Perm(), err
}
