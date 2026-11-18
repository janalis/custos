package project

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/project/config"
)

// Discovery of a captured root's inputs, confined to the root.

// listing is one deterministic walk of the project and vendor inputs.
type listing struct {
	own, selected, php, vendor []string
	stats                      map[string]FileStat
	skipped                    []string
	skippedCount               int
	// unseen is what the walks left out below the paths they scanned (see
	// Skipped), for the index world.
	unseen Skipped
}

// found is what scanning one directory produced, in listing order.
type found struct {
	names  []string
	infos  []fs.FileInfo
	dirs   []string
	links  []string
	pruned []string
	err    error
}

// splitLevels is the depth at which a walk stops listing single directories
// and descends into whole subtrees.
const splitLevels = 2

func (l *listing) skip(name, reason string) {
	l.skippedCount++
	if len(l.skipped) < maxSkipped {
		l.skipped = append(l.skipped, name+" ("+reason+")")
	}
}

// discover lists inputs without reading them. Symlinks are never followed:
// one under a project path is an error, one under vendor is skipped.
func discover(ctx context.Context, root *os.Root, cfg *config.Config, patterns []string, vendor string, paths []string) (*listing, error) {
	if len(paths) == 0 {
		paths = cfg.Paths
	}
	l := &listing{stats: map[string]FileStat{}}
	packageToSkip := path.Join(vendor, "janalis/custos")
	own, selected, php := map[string]bool{}, map[string]bool{}, map[string]bool{}
	vendorBytes := int64(0)
	walk := func(base string, symbolsOnly bool) error {
		base = path.Clean(filepath.ToSlash(base))
		if base != "." && !fs.ValidPath(base) {
			return fmt.Errorf("invalid relative path")
		}
		inVendor := symbolsOnly && base == vendor
		scanned := base
		file := func(name string, info fs.FileInfo) {
			match := strings.HasSuffix(name, ".php")
			if !symbolsOnly {
				match = match || matchesAny(path.Base(name), patterns)
			}
			if !match {
				if loadable(name) {
					l.unseen.Files = append(l.unseen.Files, name)
				}
				return
			}
			if inVendor {
				if info.Size() > syntax.MaxFileSize {
					l.skip(name, "too large")
					return
				}
				vendorBytes += info.Size()
				l.vendor = append(l.vendor, name)
			} else {
				own[name] = true
				if strings.HasSuffix(name, ".php") {
					php[name] = true
				}
				// Declaration files below a scanned path are indexed, not
				// analyzed (see declarationOnly).
				if !symbolsOnly && (name == scanned || !declarationOnly(path.Base(name))) {
					selected[name] = true
				}
			}
			l.stats[name] = FileStat{Size: info.Size(), ModTime: info.ModTime().UnixNano()}
		}
		// Each directory is opened once and its entries are reached by a
		// single path component, which keeps confined traversal cheap.
		scan := func(base string, recurse bool) *found {
			f := &found{}
			var visit func(parent *os.Root, entry, name string) error
			visit = func(parent *os.Root, entry, name string) error {
				dir, err := parent.OpenRoot(entry)
				if err != nil {
					return err
				}
				defer dir.Close()
				entries, err := fs.ReadDir(dir.FS(), ".")
				for _, d := range entries {
					child := path.Join(name, d.Name())
					if child == packageToSkip && d.IsDir() {
						f.pruned = append(f.pruned, child)
					}
					if child == packageToSkip || child == vendor {
						continue
					}
					if d.Type()&os.ModeSymlink != 0 {
						if inVendor {
							f.links = append(f.links, child)
							continue
						}
						return fmt.Errorf("symlink input is not supported: %s", child)
					}
					if !d.IsDir() {
						// A file removed since the listing is simply not an input.
						if info, err := d.Info(); err == nil {
							f.names, f.infos = append(f.names, child), append(f.infos, info)
						}
						continue
					}
					if inVendor && (d.Name() == "tests" || d.Name() == "Tests" || d.Name() == "test") || excluded(cfg.Exclude, strings.TrimPrefix(child, scanned+"/")) || hidden(d.Name()) {
						f.pruned = append(f.pruned, child)
						continue
					}
					if !recurse {
						f.dirs = append(f.dirs, child)
						continue
					}
					if err := visit(dir, d.Name(), child); err != nil {
						return err
					}
				}
				return err
			}
			f.err = visit(root, base, base)
			return f
		}
		// The top levels are listed one directory at a time so that the
		// subtrees below them can be walked in parallel; results are merged
		// in listing order.
		tree := func() error {
			dirs := []string{base}
			for level := 0; len(dirs) > 0; level++ {
				out := make([]*found, len(dirs))
				if err := Each(ctx, len(dirs), func(i int) { out[i] = scan(dirs[i], level == splitLevels) }); err != nil {
					return err
				}
				dirs = nil
				for _, f := range out {
					if f.err != nil {
						return f.err
					}
					for i, name := range f.names {
						file(name, f.infos[i])
					}
					l.unseen.Dirs = append(l.unseen.Dirs, f.pruned...)
					for _, name := range f.links {
						l.skip(name, "symlink")
					}
					dirs = append(dirs, f.dirs...)
				}
			}
			return nil
		}
		info, err := root.Lstat(base)
		switch {
		case os.IsNotExist(err):
			return fmt.Errorf("path not found: %s", base)
		case err != nil:
			return err
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("symlink input is not supported: %s", base)
		case !info.IsDir():
			file(base, info)
			return nil
		}
		return tree()
	}
	for _, p := range paths {
		if err := walk(p, false); err != nil {
			return nil, err
		}
	}
	for _, p := range cfg.Paths {
		if err := walk(p, true); err != nil {
			return nil, err
		}
	}
	if _, err := root.Stat(vendor); err == nil {
		if err := walk(vendor, true); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if len(l.vendor) > maxVendorFiles || vendorBytes > maxVendorBytes {
		return nil, fmt.Errorf("vendor input budget exceeded: %d files, %d MiB (limits %d files, %d MiB)", len(l.vendor), vendorBytes>>20, maxVendorFiles, maxVendorBytes>>20)
	}
	l.own, l.selected, l.php = sorted(own), sorted(selected), sorted(php)
	sort.Strings(l.vendor)
	return l, nil
}

// excluded reports whether rel, a slash-separated directory path below a
// scanned path, ends in one of the exclude entries: a directory name or a
// path of several names, both matched at any depth (a nested application's
// var/cache too).
func excluded(patterns []string, rel string) bool {
	for _, ex := range patterns {
		ex = strings.TrimSuffix(filepath.ToSlash(ex), "/")
		if rel == ex || strings.HasSuffix(rel, "/"+ex) {
			return true
		}
	}
	return false
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
