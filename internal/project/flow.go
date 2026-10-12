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
	// A file is summarized again only when a contract its summaries were
	// computed from changed in the last round.
	snapshot := flow.NewSnapshot()
	var earlier *flow.Snapshot
	results := make([]*flow.File, len(trees))
	for round := 0; round < flow.MaxDepth; round++ {
		for i, file := range trees {
			if file != nil && (results[i] == nil || !results[i].Current(earlier, snapshot)) {
				results[i] = flow.Extract(file, ix, opt.Version, snapshot)
			}
		}
		next := flow.NewSnapshot(results...)
		if snapshot.Equal(next) {
			return next
		}
		earlier, snapshot = snapshot, next
	}
	return snapshot
}
