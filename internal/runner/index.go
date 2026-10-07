package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"custos/internal/index"
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
				src, err := os.ReadFile(files[i])
				if err != nil {
					continue
				}
				results[i] = index.Extract(syntax.ParseBest(files[i], src, opt))
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
	return ix
}
