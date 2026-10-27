// Package project discovers PHP files and analyses them in parallel.
package project

import (
	"context"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Discover lists *.php files under paths, skipping excluded directory names
// or relative paths, hidden directories (see hidden), and symlinks that do
// not lead to a regular file inside the scanned path.
func Discover(paths, exclude []string) ([]string, error) {
	return DiscoverWith(paths, exclude, nil)
}

// DiscoverWith is Discover that also lists files whose base name matches one
// of patterns (path.Match syntax, e.g. "composer.json"), as requested by
// rules implementing analysis.FilePatternRule (see Engine.FilePatterns).
func DiscoverWith(paths, exclude, patterns []string) ([]string, error) {
	var out []string
	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			out = append(out, root)
			continue
		}
		err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != root {
					rel, _ := filepath.Rel(root, p)
					if excluded(exclude, filepath.ToSlash(rel)) || hidden(d.Name()) {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if (strings.HasSuffix(p, ".php") || matchesAny(d.Name(), patterns)) && (d.Type()&fs.ModeSymlink == 0 || linkInside(root, p)) {
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	// A file named on its own and again inside a named directory is listed
	// once: fixing it twice would report its edits twice.
	seen := make(map[string]bool, len(out))
	uniq := out[:0]
	for _, p := range out {
		if c := filepath.Clean(p); !seen[c] {
			seen[c] = true
			uniq = append(uniq, p)
		}
	}
	sort.Strings(uniq)
	return uniq, nil
}

// linkInside reports whether the symlink p resolves to a regular file under
// the scanned path root. A link leading out of it would make the analysis
// depend on files outside the project.
func linkInside(root, p string) bool {
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	target, err := filepath.EvalSymlinks(p)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(real, target)
	if err != nil || !filepath.IsLocal(rel) {
		return false
	}
	info, err := os.Stat(target)
	return err == nil && info.Mode().IsRegular()
}

// hidden reports a directory name starting with a dot (.git, .idea,
// .phpunit.cache, …): tools keep their state and generated code there, so
// discovery leaves it out below a scanned path (a path may still name one).
func hidden(name string) bool { return strings.HasPrefix(name, ".") }

func matchesAny(base string, patterns []string) bool {
	for _, pat := range patterns {
		if ok, _ := path.Match(pat, base); ok {
			return true
		}
	}
	return false
}

// Run parses and analyses files with one worker per CPU. Results are
// returned in input order.
func Run(e *analysis.Engine, files []string, opt syntax.Options) []diagnostic.FileResult {
	return RunSources(e, files, nil, opt)
}

// RunSources is Run with sources already read (srcs[i] for files[i]; a nil
// or missing entry is read from disk), e.g. by BuildIndexKeep.
func RunSources(e *analysis.Engine, files []string, srcs [][]byte, opt syntax.Options) []diagnostic.FileResult {
	return runSources(e, files, srcs, opt, false)
}

// RunReport is RunSources for reporting only: each finding keeps its fix
// titles (so it still reads as fixable) but not the edit closures, which
// hold the file's syntax tree and semantic data alive until the end of
// the run (gigabytes on large legacy trees with many findings).
func RunReport(e *analysis.Engine, files []string, srcs [][]byte, opt syntax.Options) []diagnostic.FileResult {
	return runSources(e, files, srcs, opt, true)
}

func runSources(e *analysis.Engine, files []string, srcs [][]byte, opt syntax.Options, report bool) []diagnostic.FileResult {
	results := make([]diagnostic.FileResult, len(files))
	// AnalyzeBuffer isolates crashes of its own; nothing else here can fail.
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		results[i] = analyzeFile(e, files[i], src, opt)
		if report {
			dropEdits(results[i].Findings)
		}
	}, nil)
	return results
}

func analyzeFile(e *analysis.Engine, path string, src []byte, opt syntax.Options) diagnostic.FileResult {
	if src == nil {
		var err error
		if src, err = ReadSource(path); err != nil {
			return diagnostic.FileResult{Path: path, Err: err}
		}
	}
	return AnalyzeBuffer(e, path, src, opt)
}

func dropEdits(fs []diagnostic.Finding) {
	for i := range fs {
		for j := range fs[i].Fixes {
			fs[i].Fixes[j].Edits = nil
		}
	}
}
