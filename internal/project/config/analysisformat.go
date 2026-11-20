package config

import (
	"path/filepath"
	"sync"
)

// AnalysisFormat returns analysis.Config.Format for the configuration: the
// settings custos format gives each file without command-line options, and
// false for a file format.exclude leaves out. Relative paths are taken from
// base when it is not empty, else from the working directory; with a base,
// no .editorconfig above it is read. When the
// settings do not resolve, no file is formatted: custos format reports why.
func (c *Config) AnalysisFormat(base string) func(path string) (any, bool) {
	resolver := sync.OnceValues(func() (*FormatResolver, error) { return c.FormatResolver(FormatFlags{Within: base}) })
	return func(path string) (any, bool) {
		if base != "" && !filepath.IsAbs(path) {
			path = filepath.Join(base, filepath.FromSlash(path))
		}
		r, err := resolver()
		if err != nil || c.FormatExcluded(path) {
			return nil, false
		}
		res, err := r.Settings(path)
		if err != nil {
			return nil, false
		}
		return &res.Settings, true
	}
}
