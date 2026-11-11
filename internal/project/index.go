package project

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/stubs"
)

// IndexSources returns files to index: the analysed files first, then the
// other PHP files under the project paths, then those of the project's vendor
// directory (symbols only; neither is analysed). Discovery applies exclude
// everywhere; vendor also leaves out test directories and the custos package.
// Each file appears once, so a project path that holds vendor adds nothing
// twice.
func IndexSources(root string, files, paths, exclude []string) []string {
	out, _ := IndexSourcesSkipping(root, files, paths, exclude)
	return out
}

// IndexSourcesSkipping is IndexSources also returning what the walks left
// out (see Skipped): a path they could not walk counts as a skipped
// directory, the vendor directory never does for the project walk (the
// vendor walk covers it).
func IndexSourcesSkipping(root string, files, paths, exclude []string) ([]string, Skipped) {
	out := append([]string(nil), files...)
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		seen[absolute(f)] = true
	}
	var skipped Skipped
	vendor := filepath.Join(root, vendorDir(root))
	vendorAbs := absolute(vendor)
	under := func(p, dir string) bool {
		return p == dir || strings.HasPrefix(p, dir+string(filepath.Separator))
	}
	add := func(found []string, skip string) (dropped bool) {
		for _, f := range found {
			key := absolute(f)
			if under(key, skip) {
				dropped = true
				continue
			}
			if !seen[key] {
				seen[key] = true
				out = append(out, f)
			}
		}
		return dropped
	}
	// Vendor files are left to the vendor walk and its policy.
	var own Skipped
	if found, err := discoverSkipping(paths, exclude, nil, &own); err == nil {
		add(found, vendorAbs)
		for _, d := range own.Dirs {
			if !under(absolute(d), vendorAbs) {
				skipped.Dirs = append(skipped.Dirs, d)
			}
		}
		for _, f := range own.Files {
			if !under(absolute(f), vendorAbs) {
				skipped.Files = append(skipped.Files, f)
			}
		}
	} else {
		skipped.Dirs = append(skipped.Dirs, paths...)
	}
	if st, err := os.Stat(vendor); err == nil && st.IsDir() {
		if vf, err := discoverSkipping([]string{vendor}, append([]string{"node_modules", "tests", "Tests", "test"}, exclude...), nil, &skipped); err == nil {
			if self := filepath.Join(vendor, "janalis", "custos"); add(vf, absolute(self)) {
				skipped.Dirs = append(skipped.Dirs, self)
			}
		} else {
			skipped.Dirs = append(skipped.Dirs, vendor)
		}
	}
	return out, skipped
}

// absolute is path made absolute and clean, to compare listed files.
func absolute(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
}

// absoluteFrom is absolute for a caller that already knows the working
// directory cwd, without asking the system for it again.
func absoluteFrom(cwd, path string) string {
	// A Windows volume-relative path (C:foo) is left to absolute.
	if filepath.IsAbs(path) || filepath.VolumeName(path) != "" {
		return absolute(path)
	}
	return filepath.Join(cwd, path)
}

// vendorDir returns composer.json's `config.vendor-dir` (Joomla:
// `libraries/vendor`) when it is a path inside root, else "vendor".
func vendorDir(root string) string {
	b, err := safeio.ReadFile(filepath.Join(root, "composer.json"), config.MaxFileSize)
	if err != nil {
		return "vendor"
	}
	return vendorDirBytes(b)
}

func vendorDirBytes(b []byte) string {
	var c struct {
		Config struct {
			VendorDir string `json:"vendor-dir"`
		} `json:"config"`
	}
	if json.Unmarshal(b, &c) != nil || c.Config.VendorDir == "" || !filepath.IsLocal(c.Config.VendorDir) {
		return "vendor"
	}
	return filepath.Clean(c.Config.VendorDir)
}

// BuildIndex parses files in parallel and returns their symbols layered over
// the embedded stubs. Files are added in input order (Discover sorts paths),
// so when a symbol is declared more than once (polyfills) the same
// declaration wins on every run.
func BuildIndex(files []string, opt syntax.Options, options infer.Options) *index.Index {
	ix, _ := BuildIndexKeep(files, 0, opt, options)
	return ix
}

// BuildIndexKeep is BuildIndex that also returns the sources it read for the
// first keep files (nil where unreadable), so a following Run over them does
// not read them again (IndexSources lists the analysed files first).
func BuildIndexKeep(files []string, keep int, opt syntax.Options, options infer.Options) (*index.Index, [][]byte) {
	return BuildIndexCached(files, keep, opt, options, nil)
}

// BuildIndexCached is BuildIndexKeep that takes the declarations of the
// files past keep from cache while they are unchanged, records the others
// in it, and saves it.
func BuildIndexCached(files []string, keep int, opt syntax.Options, options infer.Options, cache *SymbolCache) (*index.Index, [][]byte) {
	ix, srcs, _, _ := buildIndex(files, keep, opt, options, cache, false)
	return ix, srcs
}

// BuildIndexUnread is BuildIndexCached keeping no source and also returning
// the files whose declarations it could not read (unreadable, too large or
// crashing the extraction).
func BuildIndexUnread(files []string, opt syntax.Options, options infer.Options, cache *SymbolCache) (*index.Index, []string) {
	ix, _, _, unread := buildIndex(files, 0, opt, options, cache, false)
	return ix, unread
}

// buildIndex is BuildIndexCached also returning, with keepTrees, the syntax
// trees of the first keep files (nil where unread or crashed), and the
// files whose declarations it could not read.
func buildIndex(files []string, keep int, opt syntax.Options, options infer.Options, cache *SymbolCache, keepTrees bool) (*index.Index, [][]byte, []*syntax.File, []string) {
	ix := index.New(stubs.Index())
	ix.DeferTypeAliases()
	results := make([]*index.FileSymbols, len(files))
	srcs := make([][]byte, keep)
	var trees []*syntax.File
	if keepTrees {
		trees = make([]*syntax.File, keep)
	}
	// indexOne and indexTree isolate crashes of their own.
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		switch {
		case i < keep && keepTrees:
			results[i], srcs[i], trees[i] = indexTree(files[i], opt, options)
			cache.declared(i, results[i])
			return
		case i < keep:
			results[i], srcs[i] = indexOne(files[i], opt, options)
			cache.declared(i, results[i])
			return
		}
		results[i] = indexCached(files[i], opt, options, cache)
	}, nil)
	var unread []string
	for i, fs := range results {
		if fs == nil {
			unread = append(unread, files[i])
		}
		if fs != nil && len(fs.Classes)+len(fs.Functions)+len(fs.Constants)+len(fs.ClassAliases) > 0 {
			ix.Add(fs)
		}
	}
	// Saved before DropStaleInferred rewrites declarations for this project.
	// The cache only saves time: a failure to write it changes no result.
	_ = cache.Save()
	for _, fs := range results {
		if fs != nil {
			ix.DropStaleInferred(fs)
		}
	}
	ix.LinkTypeAliases()
	return ix, srcs, trees, unread
}

// extract is ExtractSymbols; tests replace it to simulate a crash.
var extract = ExtractSymbols

// indexOne reads and extracts one file; a file too large to analyse
// declares nothing it can read (nil). A crash (parser or inference bug on
// unusual input) drops that file's symbols instead of taking down the CLI
// or the language server; the file itself still gets an "internal" finding
// when it is analysed.
func indexOne(path string, opt syntax.Options, options infer.Options) (fs *index.FileSymbols, src []byte) {
	defer func() {
		if recover() != nil {
			fs = nil
		}
	}()
	src, err := ReadSource(path)
	if err != nil {
		return nil, nil
	}
	if len(src) > syntax.MaxFileSize {
		return nil, src
	}
	return extract(path, src, opt, options), src
}

// indexTree is indexOne keeping the file's syntax tree.
func indexTree(path string, opt syntax.Options, options infer.Options) (fs *index.FileSymbols, src []byte, tree *syntax.File) {
	defer func() {
		if recover() != nil {
			fs, tree = nil, nil
		}
	}()
	src, err := ReadSource(path)
	if err != nil {
		return nil, nil, nil
	}
	tree = syntax.ParseBest(path, src, opt)
	if len(src) > syntax.MaxFileSize {
		return nil, src, tree
	}
	return extractTree(tree, opt, options), src, tree
}

// indexCached is indexOne for a file whose source is not kept: its
// declarations come from cache while its stat is unchanged.
func indexCached(path string, opt syntax.Options, options infer.Options, cache *SymbolCache) *index.FileSymbols {
	if cache == nil {
		fs, _ := indexOne(path, opt, options)
		return fs
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() > syntax.MaxFileSize {
		return nil
	}
	if fs := cache.lookup(path, info); fs != nil {
		return fs
	}
	fs, _ := indexOne(path, opt, options)
	cache.store(path, info, fs)
	return fs
}

// ExtractSymbols parses one file and returns its symbols, with the return
// types of untyped functions and methods inferred from their bodies
// (infer.AnnotateReturns). Before adding the result to a project index,
// call DropStaleInferred on it once the project symbols are known.
func ExtractSymbols(path string, src []byte, opt syntax.Options, options infer.Options) *index.FileSymbols {
	return extractTree(syntax.ParseBest(path, src, opt), opt, options)
}

// extractTree is ExtractSymbols for a file already parsed; tests replace it
// to simulate a crash.
var extractTree = func(f *syntax.File, opt syntax.Options, options infer.Options) *index.FileSymbols {
	fs := index.Extract(f)
	fs.Digest = index.Digest(f)
	infer.AnnotateReturns(f, fs, stubs.Index(), opt.Version, options)
	return fs
}
