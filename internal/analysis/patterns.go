package analysis

// FilePatternRule is implemented by rules that inspect files other than PHP
// sources (e.g. composer.json). FilePatterns returns base-name glob patterns
// (path.Match syntax); the runner discovers matching files in addition to
// *.php when the rule is enabled.
type FilePatternRule interface {
	FilePatterns() []string
}

// FilePatterns returns the extra base-name patterns requested by enabled
// rules (deduplicated, in rule order).
func (e *Engine) FilePatterns() []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range e.rules {
		fp, ok := r.rule.(FilePatternRule)
		if !ok {
			continue
		}
		for _, p := range fp.FilePatterns() {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
}
