package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// SymbolCache keeps, between runs, the declarations of the files a run
// indexes without analysing them (vendor and the rest of the project
// paths), so a run over a few files does not parse every dependency again.
//
// An entry is trusted while its file keeps its size and modification time;
// a file modified within racyWindow of the run is never stored, as it may
// change again without either moving. The cache key covers the custos
// binary and the parse and inference options, so a rebuild or another PHP
// version starts afresh. Entries hold declarations as extracted, before the
// index drops inferred returns that depend on project functions (see
// DropStaleInferred). A nil *SymbolCache caches nothing.
type SymbolCache struct {
	path, key string
	started   time.Time
	old       map[string]cachedSymbols
	mu        sync.Mutex
	fresh     map[string]cachedSymbols
	missed    bool
	// order is the flow order FlowOrder read.
	order map[string]int
}

type cachedSymbols struct {
	Size, ModTime int64
	Symbols       *index.FileSymbols
}

type symbolCacheFile struct {
	Key   string
	Files map[string]cachedSymbols
}

// maxSymbolCacheBytes bounds the cache file read back.
const maxSymbolCacheBytes = 1 << 30

// executable is os.Executable and userCacheDir os.UserCacheDir; tests
// replace them.
var (
	executable   = os.Executable
	userCacheDir = os.UserCacheDir
)

// CacheDir is where adapters keep declarations between runs:
// $CUSTOS_CACHE_DIR, else custos under the user cache directory; "" when
// neither is known.
func CacheDir() string {
	if dir := os.Getenv("CUSTOS_CACHE_DIR"); dir != "" {
		return dir
	}
	dir, err := userCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "custos")
}

// OpenSymbolCache loads the cache of the project at root from dir. A
// missing, unreadable or outdated cache file starts empty; nil is returned
// when the running binary cannot be identified.
func OpenSymbolCache(dir, root string, opt syntax.Options, options infer.Options) *SymbolCache {
	exe, err := executable()
	if err != nil {
		return nil
	}
	info, err := os.Stat(exe)
	if err != nil {
		return nil
	}
	settings := fmt.Sprint(opt.Version, opt.ShortOpenTag, options)
	key := fmt.Sprint(exe, info.Size(), info.ModTime().UnixNano(), settings)
	// The file is named by project and options only, so a new custos build
	// replaces its cache instead of adding one beside it.
	sum := sha256.Sum256([]byte(absolute(root) + "\x00" + settings))
	c := &SymbolCache{
		path:    filepath.Join(dir, "symbols-"+hex.EncodeToString(sum[:8])+".gob"),
		key:     key,
		started: time.Now(),
		fresh:   map[string]cachedSymbols{},
	}
	if b, err := safeio.ReadFile(c.path, maxSymbolCacheBytes); err == nil {
		var f symbolCacheFile
		if gob.NewDecoder(bytes.NewReader(b)).Decode(&f) == nil && f.Key == key {
			c.old = f.Files
		}
	}
	return c
}

// lookup returns the cached declarations of path when its stat is unchanged.
func (c *SymbolCache) lookup(path string, info os.FileInfo) *index.FileSymbols {
	if c == nil {
		return nil
	}
	abs := absolute(path)
	e, ok := c.old[abs]
	if !ok || e.Size != info.Size() || e.ModTime != info.ModTime().UnixNano() {
		return nil
	}
	relocate(e.Symbols, path)
	c.mu.Lock()
	c.fresh[abs] = e
	c.mu.Unlock()
	return e.Symbols
}

// store records the declarations extracted from path, read with info.
func (c *SymbolCache) store(path string, info os.FileInfo, fs *index.FileSymbols) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.missed = true
	if fs != nil && info.ModTime().Before(c.started.Add(-racyWindow)) {
		c.fresh[absolute(path)] = cachedSymbols{Size: info.Size(), ModTime: info.ModTime().UnixNano(), Symbols: fs}
	}
}

// Save writes the entries this run used, unless they are the ones it read.
// Files no longer indexed drop out.
func (c *SymbolCache) Save() error {
	if c == nil || !c.missed && len(c.fresh) == len(c.old) {
		return nil
	}
	return writeCache(c.path, symbolCacheFile{Key: c.key, Files: c.fresh})
}

// writeCache writes v to path as gob.
func writeCache(path string, v any) error {
	var buf bytes.Buffer
	// Encoding plain exported data into memory cannot fail.
	_ = gob.NewEncoder(&buf).Encode(v)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	// Write aside and rename, so a concurrent run reads a whole file.
	tmp := fmt.Sprintf("%s.%d", path, os.Getpid())
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// flowOrderFile is the order the last flow build of the project saved,
// by absolute path (see FlowBuild.Order).
type flowOrderFile struct {
	Key   string
	Order map[string]int
}

// orderPath is where the cache keeps the flow order.
func (c *SymbolCache) orderPath() string {
	return strings.TrimSuffix(c.path, ".gob") + "-order.gob"
}

// FlowOrder returns, by path as files name them, the order the last flow
// build of these files with options saved; nil when there is none.
func (c *SymbolCache) FlowOrder(files []string, options flow.Options) map[string]int {
	if c == nil {
		return nil
	}
	b, err := safeio.ReadFile(c.orderPath(), maxSymbolCacheBytes)
	if err != nil {
		return nil
	}
	var f flowOrderFile
	if gob.NewDecoder(bytes.NewReader(b)).Decode(&f) != nil || f.Key != c.key+fmt.Sprint(options) {
		return nil
	}
	c.order = f.Order
	order := make(map[string]int, len(files))
	for _, path := range files {
		if depth, ok := f.Order[absolute(path)]; ok {
			order[path] = depth
		}
	}
	return order
}

// SaveFlowOrder keeps order (by path, see FlowBuild.Order) for the next
// build with options, unless it is the one FlowOrder read.
func (c *SymbolCache) SaveFlowOrder(order map[string]int, options flow.Options) error {
	if c == nil {
		return nil
	}
	abs := make(map[string]int, len(order))
	for path, depth := range order {
		abs[absolute(path)] = depth
	}
	if maps.Equal(abs, c.order) {
		return nil
	}
	return writeCache(c.orderPath(), flowOrderFile{Key: c.key + fmt.Sprint(options), Order: abs})
}

// relocate points cached declarations at path when an earlier run named
// the same file differently (relative to another working directory).
func relocate(fs *index.FileSymbols, path string) {
	if fs.Path == path {
		return
	}
	old := fs.Path
	fs.Path = path
	for _, c := range fs.Classes {
		if c.File == old {
			c.File = path
		}
	}
	for _, f := range fs.Functions {
		if f.File == old {
			f.File = path
		}
	}
	for _, k := range fs.Constants {
		if k.File == old {
			k.File = path
		}
	}
}
