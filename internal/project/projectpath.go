package project

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"custos/internal/project/config"
)

// ProjectPaths returns analysis.Config.ProjectPath: a file path relative to
// its project root, with forward slashes. The root is the nearest ancestor
// directory holding custos.json, else the outermost one holding
// composer.json (a package or fixture application nested in a project stays
// part of it), else fallback. A file outside its root keeps its path.
// exists reports whether a file exists; it is asked once per directory.
func ProjectPaths(exists func(name string) bool, fallback string) func(path string) string {
	var mu sync.Mutex
	roots := map[string]string{}
	rootOf := func(dir string) string {
		mu.Lock()
		defer mu.Unlock()
		if root, ok := roots[dir]; ok {
			return root
		}
		root := fallback
		for d := dir; ; d = filepath.Dir(d) {
			if exists(filepath.Join(d, config.FileName)) {
				root = d
				break
			}
			if exists(filepath.Join(d, "composer.json")) {
				root = d
			}
			if filepath.Dir(d) == d {
				break
			}
		}
		roots[dir] = root
		return root
	}
	return func(path string) string {
		rel, err := filepath.Rel(rootOf(filepath.Dir(path)), path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			rel = path
		}
		return strings.ReplaceAll(filepath.ToSlash(rel), `\`, "/")
	}
}

// DiskProjectPaths is ProjectPaths over the filesystem for absolute or
// working-directory-relative paths, falling back to fallback (made
// absolute).
func DiskProjectPaths(fallback string) func(path string) string {
	fallback, _ = filepath.Abs(fallback)
	paths := ProjectPaths(func(name string) bool {
		_, err := os.Stat(name)
		return err == nil
	}, fallback)
	return func(path string) string {
		abs, _ := filepath.Abs(path)
		return paths(abs)
	}
}
