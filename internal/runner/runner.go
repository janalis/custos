// Package runner discovers PHP files and analyses them in parallel.
package runner

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

// FileResult is the outcome for one file.
type FileResult struct {
	Path     string // as discovered (relative to the working directory when possible)
	Src      []byte
	Findings []analysis.Finding
	Errors   []syntax.Error // syntax errors
	Err      error          // I/O error
}

// Discover lists *.php files under paths, skipping excluded directory names
// or relative paths.
func Discover(paths, exclude []string) ([]string, error) {
	return DiscoverWith(paths, exclude, nil)
}

// DiscoverWith is Discover that also lists files whose base name matches one
// of patterns (path.Match syntax, e.g. "composer.json"), as requested by
// rules implementing analysis.FilePatternRule (see Engine.FilePatterns).
func DiscoverWith(paths, exclude, patterns []string) ([]string, error) {
	ex := map[string]bool{}
	for _, e := range exclude {
		ex[filepath.ToSlash(strings.TrimSuffix(e, "/"))] = true
	}
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
					if ex[d.Name()] || ex[filepath.ToSlash(rel)] {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if strings.HasSuffix(p, ".php") || matchesAny(d.Name(), patterns) {
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

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
func Run(e *analysis.Engine, files []string, opt syntax.Options) []FileResult {
	results := make([]FileResult, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	workers := runtime.GOMAXPROCS(0)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = analyzeFile(e, files[i], opt)
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

func analyzeFile(e *analysis.Engine, path string, opt syntax.Options) FileResult {
	src, err := os.ReadFile(path)
	if err != nil {
		return FileResult{Path: path, Err: err}
	}
	f := syntax.ParseBest(path, src, opt)
	return FileResult{Path: path, Src: src, Findings: e.Analyze(f), Errors: f.Errors}
}
