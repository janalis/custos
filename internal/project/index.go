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
	out := append([]string(nil), files...)
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		seen[absolute(f)] = true
	}
	vendor := filepath.Join(root, vendorDir(root))
	add := func(found []string, skip ...string) {
	next:
		for _, f := range found {
			key := absolute(f)
			for _, dir := range skip {
				if strings.HasPrefix(key, dir+string(filepath.Separator)) {
					continue next
				}
			}
			if !seen[key] {
				seen[key] = true
				out = append(out, f)
			}
		}
	}
	// Vendor files are left to the vendor walk and its policy.
	if own, err := Discover(paths, exclude); err == nil {
		add(own, absolute(vendor))
	}
	if st, err := os.Stat(vendor); err == nil && st.IsDir() {
		if vf, err := Discover([]string{vendor}, append([]string{"node_modules", "tests", "Tests", "test"}, exclude...)); err == nil {
			add(vf, absolute(filepath.Join(vendor, "janalis", "custos")))
		}
	}
	return out
}

// absolute is path made absolute and clean, to compare listed files.
func absolute(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
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
	ix := index.New(stubs.Index())
	ix.DeferTypeAliases()
	results := make([]*index.FileSymbols, len(files))
	srcs := make([][]byte, keep)
	// indexOne isolates crashes of its own.
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		if i < keep {
			results[i], srcs[i] = indexOne(files[i], opt, options)
			return
		}
		results[i] = indexCached(files[i], opt, options, cache)
	}, nil)
	for _, fs := range results {
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
	return ix, srcs
}

// extract is ExtractSymbols; tests replace it to simulate a crash.
var extract = ExtractSymbols

// indexOne reads and extracts one file. A crash (parser or inference bug on
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
	return extract(path, src, opt, options), src
}

// indexCached is indexOne for a file whose source is not kept: its
// declarations come from cache while its stat is unchanged.
func indexCached(path string, opt syntax.Options, options infer.Options, cache *SymbolCache) *index.FileSymbols {
	if cache == nil {
		fs, _ := indexOne(path, opt, options)
		return fs
	}
	info, err := os.Stat(path)
	if err != nil {
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
	f := syntax.ParseBest(path, src, opt)
	fs := index.Extract(f)
	infer.AnnotateReturns(f, fs, stubs.Index(), opt.Version, options)
	return fs
}
