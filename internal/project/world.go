package project

import (
	"encoding/json"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/platform/safeio"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// ComposerWorld derives what the index knows of the project's world (see
// index.World) from Composer's metadata under root. scope lists the project
// directories whose PHP files are indexed, indexed the files of the index
// and skipped what the walks and the indexing left out (paths absolute or
// relative to the working directory).
func ComposerWorld(root string, scope, indexed []string, skipped Skipped) index.World {
	// Abs fails only when the working directory is gone; Rel then fails
	// for every path and the world stays incomplete.
	absRoot, _ := filepath.Abs(root)
	rel := func(paths []string) []string {
		out := make([]string, 0, len(paths))
		for _, p := range paths {
			abs, _ := filepath.Abs(p)
			if r, err := filepath.Rel(absRoot, abs); err == nil {
				out = append(out, filepath.ToSlash(r))
			}
		}
		return out
	}
	read := func(name string) ([]byte, error) {
		return safeio.ReadFile(filepath.Join(absRoot, filepath.FromSlash(name)), 4*config.MaxFileSize)
	}
	return composerWorld(read, rel(scope), rel(indexed), Skipped{Dirs: rel(skipped.Dirs), Files: rel(skipped.Files)})
}

// WorldInput reports whether path is a file ComposerWorld reads under
// root: composer.json, or the installed.json of the vendor directory it
// configures.
func WorldInput(root, path string) bool {
	path = filepath.Clean(path)
	return path == filepath.Join(root, "composer.json") || path == filepath.Join(root, vendorDir(root), "composer", "installed.json")
}

// composerManifest is the part of composer.json and of an installed
// package's entry the world needs.
type composerManifest struct {
	Name        string            `json:"name"`
	Require     map[string]string `json:"require"`
	RequireDev  map[string]string `json:"require-dev"`
	Replace     map[string]string `json:"replace"`
	Provide     map[string]string `json:"provide"`
	Suggest     map[string]string `json:"suggest"`
	Autoload    composerAutoload  `json:"autoload"`
	AutoloadDev composerAutoload  `json:"autoload-dev"`
	InstallPath string            `json:"install-path"`
}

type composerAutoload struct {
	PSR4     map[string]json.RawMessage `json:"psr-4"`
	PSR0     map[string]json.RawMessage `json:"psr-0"`
	Classmap []string                   `json:"classmap"`
	Files    []string                   `json:"files"`
}

func (a composerAutoload) empty() bool {
	return len(a.PSR4)+len(a.PSR0)+len(a.Classmap)+len(a.Files) == 0
}

// dirs returns the directories (or files) of an autoload mapping value,
// which Composer accepts as one string or a list.
func dirs(raw json.RawMessage) []string {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return []string{one}
	}
	var many []string
	_ = json.Unmarshal(raw, &many)
	return many
}

// composerWorld is ComposerWorld on paths relative to the project root
// (slash-separated); read reads a file by such a path.
//
// The world is complete when composer.json and the installed.json Composer
// writes into the vendor directory exist, every package the project
// requires, development ones included, is installed (or replaced or
// provided by an installed one), every installed package that autoloads
// code has indexed files, every package the project suggests is
// installed, every classmap entry of the project and of the packages was
// walked whole and every files entry was indexed. Without the development
// packages, production code may still name them (tool extensions, optional
// integrations), so the world is not complete. PSR-4 and PSR-0 prefixes
// whose directories are out of scope, or lie in a directory the walks
// skipped, are blind; below a PSR-4 directory, so are the namespaces of
// the directories that hold what the walks or the indexing left out.
func composerWorld(read func(string) ([]byte, error), scope, indexed []string, skipped Skipped) index.World {
	b, err := read("composer.json")
	if err != nil {
		return index.World{}
	}
	var root composerManifest
	if json.Unmarshal(b, &root) != nil {
		return index.World{}
	}
	vendor := vendorDirBytes(b)
	b, err = read(path.Join(vendor, "composer", "installed.json"))
	if err != nil {
		return index.World{}
	}
	var packages []composerManifest
	var v2 struct {
		Packages []composerManifest `json:"packages"`
	}
	if json.Unmarshal(b, &v2) == nil && v2.Packages != nil {
		packages = v2.Packages
	} else if json.Unmarshal(b, &packages) != nil {
		return index.World{}
	}
	installed := map[string]bool{}
	for _, p := range packages {
		installed[strings.ToLower(p.Name)] = true
		for name := range p.Replace {
			installed[strings.ToLower(name)] = true
		}
		for name := range p.Provide {
			installed[strings.ToLower(name)] = true
		}
	}
	// Code may name a package the project only suggests.
	for _, require := range []map[string]string{root.Require, root.RequireDev, root.Suggest} {
		for name := range require {
			// Platform requirements (php, ext-*, lib-*, composer-*) have no
			// vendor name.
			if strings.Contains(name, "/") && !installed[strings.ToLower(name)] {
				return index.World{}
			}
		}
	}
	// An extension the stubs lack may declare any name.
	platform := []map[string]string{root.Require, root.RequireDev, root.Suggest}
	for _, p := range packages {
		platform = append(platform, p.Require)
	}
	for _, require := range platform {
		for name := range require {
			if ext, ok := strings.CutPrefix(strings.ToLower(name), "ext-"); ok && !knownExtension(ext) {
				return index.World{}
			}
		}
	}
	files := map[string]bool{}
	for _, f := range indexed {
		for d := path.Dir(f); d != "." && d != "/" && !files[d]; d = path.Dir(d) {
			files[d] = true
		}
	}
	g := newGaps(indexed, skipped, vendor)
	w := index.World{Complete: true}
	inScope := func(p string) bool {
		for _, s := range scope {
			s = path.Clean(s)
			if s == "." || p == s || strings.HasPrefix(p, s+"/") {
				return true
			}
		}
		return false
	}
	// autoload adds the blind prefixes of a, whose paths are relative to
	// dir, and reports whether the world stays complete. For the project's
	// own configuration (own), out of scope counts as skipped.
	autoload := func(a composerAutoload, dir string, own bool) bool {
		for _, psr4 := range []bool{true, false} {
			mapping := a.PSR4
			if !psr4 {
				mapping = a.PSR0
			}
			for prefix, raw := range mapping {
				for _, d := range dirs(raw) {
					d = path.Join(dir, d)
					var blind, names []string
					if own && !inScope(d) || g.within(d) {
						blind = []string{""}
					} else {
						blind, names = g.psr(d)
					}
					if len(blind)+len(names) == 0 {
						continue
					}
					// A fallback directory or a PSR-0 prefix of class names
					// may hold any name.
					if prefix == "" || !strings.HasSuffix(prefix, `\`) {
						return false
					}
					if !psr4 {
						// PSR-0 maps underscores to directories too.
						blind, names = []string{""}, nil
					}
					for _, sub := range blind {
						w.Blind = append(w.Blind, strings.ToLower(prefix+sub))
					}
					for _, name := range names {
						w.Names = append(w.Names, strings.ToLower(prefix+name))
					}
				}
			}
		}
		for _, p := range a.Classmap {
			p = path.Join(dir, p)
			if own && !inScope(p) || g.within(p) || g.holds(p) {
				return false
			}
		}
		for _, p := range a.Files {
			// Only what was indexed (and read) can be seen.
			if p = path.Join(dir, p); !g.indexed[p] || g.file[p] {
				return false
			}
		}
		return true
	}
	// The project's own autoload configuration.
	for _, a := range []composerAutoload{root.Autoload, root.AutoloadDev} {
		if !autoload(a, ".", true) {
			return index.World{}
		}
	}
	// Installed packages: each that autoloads code must have been indexed.
	for _, p := range packages {
		if p.Autoload.empty() {
			continue
		}
		dir := path.Join(vendor, p.Name)
		if p.InstallPath != "" {
			dir = path.Join(vendor, "composer", p.InstallPath)
		}
		if !files[dir] || !autoload(p.Autoload, dir, false) {
			return index.World{}
		}
	}
	return w
}

// knownExtension reports whether the stubs describe the extension a
// Composer platform requirement names (lower case, after `ext-`). PDO
// drivers belong to PDO, mysqlnd declares nothing of its own.
func knownExtension(name string) bool {
	switch {
	case strings.HasPrefix(name, "pdo_"):
		name = "pdo"
	case name == "opcache":
		name = "zend-opcache"
	case name == "mysqlnd":
		return true
	}
	return stubs.Extensions()[strings.ReplaceAll(name, " ", "-")]
}

// gaps answers where the walks for the index left something out (paths
// relative to the project root, slash-separated and clean).
type gaps struct {
	indexed, dir, file map[string]bool
	// dirs and files list the skipped directories and files, sorted.
	dirs, files []string
	// vendor is the vendor directory, which its own walk reads.
	vendor string
}

func newGaps(indexed []string, skipped Skipped, vendor string) *gaps {
	g := &gaps{indexed: map[string]bool{}, dir: map[string]bool{}, file: map[string]bool{}, vendor: path.Clean(filepath.ToSlash(vendor))}
	for _, f := range indexed {
		g.indexed[path.Clean(f)] = true
	}
	for _, d := range skipped.Dirs {
		d = path.Clean(d)
		g.dir[d] = true
		g.dirs = append(g.dirs, d)
	}
	for _, f := range skipped.Files {
		f = path.Clean(f)
		g.file[f] = true
		g.files = append(g.files, f)
	}
	sort.Strings(g.dirs)
	sort.Strings(g.files)
	return g
}

// within reports whether p (clean) is, or lies in, a skipped directory. A
// directory skipped by the project walk that holds the vendor directory
// (vendor-dir lib/vendor, lib excluded) does not count for vendor paths:
// the vendor walk reads them.
func (g *gaps) within(p string) bool {
	inVendor := p == g.vendor || strings.HasPrefix(p, g.vendor+"/")
	for {
		if g.dir[p] && !(inVendor && strings.HasPrefix(g.vendor, p+"/")) {
			return true
		}
		parent := path.Dir(p)
		if parent == p || p == "." {
			return false
		}
		p = parent
	}
}

// under returns the skipped directories and files strictly below p
// (clean), and the prefix their paths share.
func (g *gaps) under(p string) (dirs, files []string, prefix string) {
	prefix = p + "/"
	if p == "." {
		prefix = ""
	}
	span := func(list []string) []string {
		i := sort.SearchStrings(list, prefix)
		j := i
		for j < len(list) && strings.HasPrefix(list[j], prefix) {
			j++
		}
		return list[i:j]
	}
	return span(g.dirs), span(g.files), prefix
}

// holds reports whether the classmap entry p (clean) is, or holds, a
// skipped file or directory.
func (g *gaps) holds(p string) bool {
	dirs, files, _ := g.under(p)
	return g.file[p] || len(dirs)+len(files) > 0
}

// psr returns, relative to the namespace a PSR directory p (clean) maps,
// what may name what the walks left out below it: the namespace prefixes
// of skipped directories (their paths with backslashes and a trailing
// one), and the class names of skipped files a PSR autoloader can load
// (*.php or *.hh named after a class: their paths without the extension).
func (g *gaps) psr(p string) (prefixes, names []string) {
	dirs, files, prefix := g.under(p)
	ns := func(rel string) string { return strings.ReplaceAll(strings.TrimPrefix(rel, prefix), "/", `\`) }
	for _, d := range dirs {
		prefixes = append(prefixes, ns(d)+`\`)
	}
	for _, f := range files {
		ext := path.Ext(f)
		name := strings.TrimSuffix(f, ext)
		if (ext == ".php" || ext == ".hh") && !strings.Contains(path.Base(name), ".") {
			names = append(names, ns(name))
		}
	}
	return prefixes, names
}
