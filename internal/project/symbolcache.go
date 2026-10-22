package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
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

// executable is os.Executable; tests replace it.
var executable = os.Executable

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
	var buf bytes.Buffer
	// Encoding plain exported data into memory cannot fail.
	_ = gob.NewEncoder(&buf).Encode(symbolCacheFile{Key: c.key, Files: c.fresh})
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	// Write aside and rename, so a concurrent run reads a whole file.
	tmp := fmt.Sprintf("%s.%d", c.path, os.Getpid())
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, c.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
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
