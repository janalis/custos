package project

import (
	"context"

	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// BuildFlowSnapshot builds project-local summaries without including vendor
// bodies. Parsing is bounded by the same source reader as ordinary analysis.
// Trees are retained only during bounded composition, then released.
func BuildFlowSnapshot(files []string, srcs [][]byte, opt syntax.Options, ix *index.Index) *flow.Snapshot {
	trees := make([]*syntax.File, len(files))
	_ = Each(context.Background(), len(files), func(i int) {
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
	snapshot := flow.NewSnapshot()
	for round := 0; round < flow.MaxDepth; round++ {
		// Every file of a round reads the previous round's summaries only.
		extracted := make([]*flow.File, len(trees))
		_ = Each(context.Background(), len(trees), func(i int) {
			if trees[i] != nil {
				extracted[i] = flow.Extract(trees[i], ix, opt.Version, snapshot)
			}
		})
		results := []*flow.File{}
		for _, file := range extracted {
			if file != nil {
				results = append(results, file)
			}
		}
		next := flow.NewSnapshot(results...)
		if snapshot.Equal(next) {
			return next
		}
		snapshot = next
	}
	return snapshot
}
