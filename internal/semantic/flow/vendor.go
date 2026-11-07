package flow

import (
	"strconv"
	"strings"
	"sync"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// MaxVendorBytes bounds the dependency files summarized on demand: larger ones
// are generated code or data, and keep their calls unresolved.
const MaxVendorBytes = 512 << 10

// Vendor summarizes dependency files on demand. A contract that no file of
// the snapshot declares is answered from the single dependency file declaring
// the callable. Each such file is summarized once, alone: without a snapshot
// and without further dependency lookups, so its summaries depend on nothing
// but the file and the index, whichever lookup asks first.
type Vendor struct {
	index  *index.Index
	php    phpversion.Version
	parse  syntax.Options
	assert bool
	// within tells a dependency file; read loads one.
	within func(path string) bool
	read   func(path string) ([]byte, error)

	mu    sync.Mutex
	files map[string]*vendorFile
	// answers memoizes lookups by key: flow asks the same contracts and class
	// metadata for every call.
	answers sync.Map
}

type vendorAnswer struct {
	summary Summary
	ok      bool
}

type vendorFile struct {
	once      sync.Once
	summaries map[string]Summary
}

// NewVendor returns a provider for the dependency files that within accepts,
// read with read and summarized against ix.
func NewVendor(ix *index.Index, parse syntax.Options, assertions bool, within func(string) bool, read func(string) ([]byte, error)) *Vendor {
	return &Vendor{index: ix, php: parse.Version, parse: parse, assert: assertions, within: within, read: read, files: map[string]*vendorFile{}}
}

// String keeps option printing (cache keys) independent of the provider.
func (o Options) String() string { return "{" + strconv.FormatBool(o.Assertions) + "}" }

// lookup answers a folded summary key from the dependency file declaring it.
func (v *Vendor) lookup(key string) (Summary, bool) {
	if known, ok := v.answers.Load(key); ok {
		answer := known.(vendorAnswer)
		return answer.summary, answer.ok
	}
	summary, ok := v.answer(key)
	v.answers.Store(key, vendorAnswer{summary, ok})
	return summary, ok
}

func (v *Vendor) answer(key string) (Summary, bool) {
	path := v.declaringFile(key)
	if path == "" {
		return Summary{}, false
	}
	v.mu.Lock()
	f := v.files[path]
	if f == nil {
		f = &vendorFile{}
		v.files[path] = f
	}
	v.mu.Unlock()
	f.once.Do(func() { f.summaries = v.summarize(path) })
	summary, ok := f.summaries[key]
	return summary, ok
}

// declaringFile is the dependency file declaring the callable, or the class
// whose metadata the key names, exactly once; "" otherwise.
func (v *Vendor) declaringFile(key string) string {
	path := ""
	if class, metadata := strings.CutPrefix(key, sinkClassKey("")); metadata {
		path = v.classFile(class)
	} else if class, method, isMethod := strings.Cut(key, "::"); isMethod {
		if method != "" {
			path = v.classFile(class)
		}
	} else if decls := v.index.FunctionDecls(key, v.php); len(decls) == 1 {
		path = decls[0].File
	}
	if path == "" || !v.within(path) {
		return ""
	}
	return path
}

func (v *Vendor) classFile(name string) string {
	if decls := v.index.ClassDecls(name, v.php); len(decls) == 1 {
		return decls[0].File
	}
	return ""
}

func (v *Vendor) summarize(path string) map[string]Summary {
	src, err := v.read(path)
	if err != nil || len(src) > MaxVendorBytes {
		return nil
	}
	tree := syntax.ParseBest(path, src, v.parse)
	return Extract(tree, v.index, v.php, nil, Options{Assertions: v.assert}).Summaries
}
