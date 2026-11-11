package project

import (
	"encoding/json"
	"path"
	"path/filepath"
	"strings"

	"custos/internal/platform/safeio"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
)

// ComposerWorld derives what the index knows of the project's world (see
// index.World) from Composer's metadata under root. scope lists the project
// directories whose PHP files are indexed and indexed the files of the
// index (paths absolute or relative to the working directory).
func ComposerWorld(root string, scope, indexed []string) index.World {
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
	return composerWorld(read, rel(scope), rel(indexed))
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
// installed, and every classmap and files entry of the
// project's autoload configuration lies in scope. Without the development
// packages, production code may still name them (tool extensions, optional
// integrations), so the world is not complete. PSR-4 and PSR-0 prefixes
// whose directories are out of scope (or excluded from the vendor walk)
// are blind.
func composerWorld(read func(string) ([]byte, error), scope, indexed []string) index.World {
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
	files := map[string]bool{}
	for _, f := range indexed {
		for d := path.Dir(f); d != "." && d != "/" && !files[d]; d = path.Dir(d) {
			files[d] = true
		}
	}
	w := index.World{Complete: true}
	inScope := func(p string) bool {
		p = path.Clean(p)
		for _, s := range scope {
			s = path.Clean(s)
			if s == "." || p == s || strings.HasPrefix(p, s+"/") {
				return true
			}
		}
		return false
	}
	// The project's own autoload configuration.
	for _, a := range []composerAutoload{root.Autoload, root.AutoloadDev} {
		for _, mapping := range []map[string]json.RawMessage{a.PSR4, a.PSR0} {
			for prefix, raw := range mapping {
				for _, d := range dirs(raw) {
					if inScope(d) {
						continue
					}
					// A fallback directory or a PSR-0 prefix of class names
					// may hold any name.
					if prefix == "" || !strings.HasSuffix(prefix, `\`) {
						return index.World{}
					}
					w.Blind = append(w.Blind, strings.ToLower(prefix))
				}
			}
		}
		for _, p := range append(append([]string{}, a.Classmap...), a.Files...) {
			if !inScope(p) {
				return index.World{}
			}
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
		if !files[dir] {
			return index.World{}
		}
		// The vendor walk leaves test directories out.
		for _, mapping := range []map[string]json.RawMessage{p.Autoload.PSR4, p.Autoload.PSR0} {
			for prefix, raw := range mapping {
				for _, d := range dirs(raw) {
					if !testDir(d) {
						continue
					}
					if prefix == "" || !strings.HasSuffix(prefix, `\`) {
						return index.World{}
					}
					w.Blind = append(w.Blind, strings.ToLower(prefix))
				}
			}
		}
		for _, p := range append(append([]string{}, p.Autoload.Classmap...), p.Autoload.Files...) {
			if testDir(p) {
				return index.World{}
			}
		}
	}
	return w
}

// testDir reports whether a package path crosses a directory the vendor
// walk skips.
func testDir(p string) bool {
	for _, part := range strings.Split(path.Clean(p), "/") {
		if part == "tests" || part == "Tests" || part == "test" {
			return true
		}
	}
	return false
}
