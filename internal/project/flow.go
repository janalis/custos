package project

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// BuildFlowSnapshot builds project-local summaries without including vendor
// bodies. Parsing is bounded by the same source reader as ordinary analysis.
// Trees are retained only during bounded composition, then released.
func BuildFlowSnapshot(files []string, srcs [][]byte, opt syntax.Options, ix *index.Index) *flow.Snapshot {
	trees := make([]*syntax.File, len(files))
	for i, path := range files {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		if src == nil {
			var err error
			src, err = ReadSource(path)
			if err != nil {
				continue
			}
		}
		trees[i] = syntax.ParseBest(path, src, opt)
	}
	snapshot := flow.NewSnapshot()
	for round := 0; round < flow.MaxDepth; round++ {
		results := []*flow.File{}
		for _, file := range trees {
			if file != nil {
				results = append(results, flow.Extract(file, ix, opt.Version, snapshot))
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
