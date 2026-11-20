package config

import (
	"path/filepath"
	"sync"

	"github.com/janalis/custos/internal/format"
)

// AnalysisFormat returns analysis.Config.Format for the configuration: the
// settings custos format gives each file without command-line options, and
// false for a file format.exclude leaves out. Relative paths are taken from
// base when it is not empty, else from the working directory; with a base,
// no .editorconfig above it is read. When the
// settings do not resolve, no file is formatted: custos format reports why.
func (c *Config) AnalysisFormat(base string) func(path string) (any, bool) {
	resolver := sync.OnceValues(func() (*FormatResolver, error) { return c.FormatResolver(FormatFlags{Within: base}) })
	// The findings cache key and the rule both ask for a file's settings:
	// resolve each path once for the life of the configuration.
	var memo sync.Map // path → *format.Settings, nil when not formatted
	return func(path string) (any, bool) {
		if base != "" && !filepath.IsAbs(path) {
			path = filepath.Join(base, filepath.FromSlash(path))
		}
		if v, ok := memo.Load(path); ok {
			s := v.(*format.Settings)
			return s, s != nil
		}
		var s *format.Settings
		if r, err := resolver(); err == nil && !c.FormatExcluded(path) {
			if res, err := r.Settings(path); err == nil {
				s = &res.Settings
			}
		}
		memo.Store(path, s)
		return s, s != nil
	}
}
