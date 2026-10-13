package project

import (
	"runtime"
	"sync"

	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// BuildFlowSnapshot builds project-local summaries without including vendor
// bodies. Parsing is bounded by the same source reader as ordinary analysis.
// Trees are retained only during bounded composition, then released.
func BuildFlowSnapshot(files []string, srcs [][]byte, opt syntax.Options, ix *index.Index) *flow.Snapshot {
	trees := make([]*syntax.File, len(files))
	each(len(files), runtime.GOMAXPROCS(0), func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		if src == nil {
			var err error
			src, err = ReadSource(files[i])
			if err != nil {
				return
			}
		}
		trees[i] = syntax.ParseBest(files[i], src, opt)
	})
	// A file is summarized again only when a contract its summaries were
	// computed from changed in the last round. The files of a round read the
	// snapshot of the round before: they are summarized side by side.
	snapshot := flow.NewSnapshot()
	var earlier *flow.Snapshot
	results := make([]*flow.File, len(trees))
	for round := 0; round < flow.MaxDepth; round++ {
		each(len(trees), max(1, runtime.GOMAXPROCS(0)/2), func(i int) {
			if file := trees[i]; file != nil && (results[i] == nil || !results[i].Current(earlier, snapshot)) {
				results[i] = flow.Extract(file, ix, opt.Version, snapshot)
			}
		})
		next := flow.NewSnapshot(results...)
		if snapshot.Equal(next) {
			return next
		}
		earlier, snapshot = snapshot, next
	}
	return snapshot
}

// each runs work for every index below n on that many goroutines and returns
// when all are done. Each call of work owns its index.
func each(n, workers int, work func(i int)) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				work(i)
			}
		}()
	}
	for i := range n {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}
