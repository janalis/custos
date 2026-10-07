package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

// IndexSources returns files to index: the analysed files plus the
// project's vendor directory (symbols only; vendor is not analysed).
func IndexSources(root string, files []string) []string {
	out := append([]string(nil), files...)
	vendor := filepath.Join(root, "vendor")
	if st, err := os.Stat(vendor); err == nil && st.IsDir() {
		if vf, err := Discover([]string{vendor}, []string{".git", "node_modules", "tests", "Tests", "test"}); err == nil {
			out = append(out, vf...)
		}
	}
	return out
}

// BuildIndex parses files in parallel and returns their symbols layered over
// the embedded stubs. Files are added in input order (Discover sorts paths),
// so when a symbol is declared more than once (polyfills) the same
// declaration wins on every run.
func BuildIndex(files []string, opt syntax.Options) *index.Index {
	ix := index.New(stubs.Index())
	results := make([]*index.FileSymbols, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < runtime.GOMAXPROCS(0); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = indexOne(files[i], opt)
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for _, fs := range results {
		if fs != nil && len(fs.Classes)+len(fs.Functions)+len(fs.Constants) > 0 {
			ix.Add(fs)
		}
	}
	for _, fs := range results {
		if fs != nil {
			ix.DropStaleInferred(fs)
		}
	}
	return ix
}

// extract is ExtractSymbols; tests replace it to simulate a crash.
var extract = ExtractSymbols

// indexOne reads and extracts one file. A crash (parser or inference bug on
// unusual input) drops that file's symbols instead of taking down the CLI
// or the language server; the file itself still gets an "internal" finding
// when it is analysed.
func indexOne(path string, opt syntax.Options) (fs *index.FileSymbols) {
	defer func() {
		if recover() != nil {
			fs = nil
		}
	}()
	src, err := ReadSource(path)
	if err != nil {
		return nil
	}
	return extract(path, src, opt)
}

// ExtractSymbols parses one file and returns its symbols, with the return
// types of untyped functions and methods inferred from their bodies
// (infer.AnnotateReturns). Before adding the result to a project index,
// call DropStaleInferred on it once the project symbols are known.
func ExtractSymbols(path string, src []byte, opt syntax.Options) *index.FileSymbols {
	f := syntax.ParseBest(path, src, opt)
	fs := index.Extract(f)
	infer.AnnotateReturns(f, fs, stubs.Index(), opt.Version)
	return fs
}
