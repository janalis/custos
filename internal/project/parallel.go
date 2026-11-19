package project

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/janalis/custos/internal/diagnostic"
)

// parallel runs work(i) for every i below n on up to workers goroutines (one
// per CPU when workers < 1) and returns when every started item is done.
// Items start in index order. Cancellation is observed between items: once
// ctx is done the rest are not started and ctx.Err() is returned. A panic in
// work(i) is recovered and handed to crashed(i, value), or dropped when
// crashed is nil (work then reports its own failures and leaves a crashed
// item unset). A panic on a worker goroutine would otherwise end the whole
// process (CLI run, language server or MCP server), which no guard on the
// calling goroutine prevents.
//
// Workers claim the next item from a shared counter instead of receiving it
// over a channel: a hand-off per item parked and woke a goroutine each time,
// and the idle processors spun looking for work meanwhile.
func parallel(ctx context.Context, n, workers int, work func(int), crashed func(int, any)) error {
	if workers < 1 {
		workers = runtime.GOMAXPROCS(0)
	}
	var next atomic.Int64
	var canceled atomic.Bool
	var wg sync.WaitGroup
	for range min(n, workers) {
		wg.Go(func() {
			for !canceled.Load() {
				i := int(next.Add(1) - 1)
				if i >= n {
					return
				}
				if ctx.Err() != nil {
					canceled.Store(true)
					return
				}
				isolate(i, work, crashed)
			}
		})
	}
	wg.Wait()
	if canceled.Load() {
		return ctx.Err()
	}
	return nil
}

func isolate(i int, work func(int), crashed func(int, any)) {
	defer func() {
		if r := recover(); r != nil && crashed != nil {
			crashed(i, r)
		}
	}()
	work(i)
}

// Each runs fn(i) for every i below n on one worker per CPU. Only the
// dispatcher observes cancellation, between items. A panic in fn fails the
// call (the first one is returned as an error) instead of the process; the
// other items still run.
func Each(ctx context.Context, n int, fn func(int)) error {
	var mu sync.Mutex
	var crash error
	err := parallel(ctx, n, 0, fn, func(i int, r any) {
		mu.Lock()
		defer mu.Unlock()
		if crash == nil {
			crash = fmt.Errorf("internal failure on item %d: %v", i, r)
		}
	})
	if crash != nil {
		return crash
	}
	return err
}

// internalFailure is the result of a file whose analysis crashed outside
// any rule (parser, shared inference, the engine itself): an internal
// finding and an internal-failure gap, so the run goes on and the file
// cannot read as completely analyzed.
func internalFailure(path string, src []byte, r any) diagnostic.FileResult {
	return diagnostic.FileResult{
		Path: path, Src: src,
		Findings: []diagnostic.Finding{{
			Rule: "internal", Severity: diagnostic.SeverityError,
			Message: fmt.Sprintf("custos crashed on this file and skipped it: %v", r),
		}},
		Incomplete: []diagnostic.AnalysisIssue{{Reason: "internal-failure"}},
	}
}
