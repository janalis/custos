package project

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/janalis/custos/internal/platform/safeio"
)

// The format cache keeps, between runs of custos format, the files known to
// be already formatted, so a run skips formatting them again. An entry is
// the SHA-256 of the settings, the PHP version and the source of a file
// formatting left unchanged, without anything to report; only such
// outcomes are kept, so a hit always means "unchanged". The file is named
// after the project root and starts with the identity of the binary that
// wrote it: a new build starts afresh.
//
// Layout: formatCacheMagic, a uvarint-prefixed binary identity, then
// entries of sha256.Size bytes. A file that does not read so is ignored.

// formatCacheMagic starts a format cache file; a new layout changes it.
const formatCacheMagic = "custos format 1\n"

// maxFormatEntries bounds the entries kept (32 bytes each: 8 MiB).
const maxFormatEntries = 1 << 18

// FormatKey identifies the outcome of formatting one source.
type FormatKey [sha256.Size]byte

// FormatCache is the set of sources known to be formatted. A nil cache
// keeps nothing. It is safe for concurrent use.
type FormatCache struct {
	path, id string
	old      map[FormatKey]bool

	mu sync.Mutex
	// used holds the entries this run hit or added, in order.
	used  []FormatKey
	seen  map[FormatKey]bool
	added bool
}

// OpenFormatCache loads the format cache of the project at root from dir;
// nil when dir is "" or the running binary cannot be identified.
func OpenFormatCache(dir, root string) *FormatCache {
	if dir == "" {
		return nil
	}
	exe, err := executable()
	if err != nil {
		return nil
	}
	info, err := os.Stat(exe)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256([]byte(absolute(root)))
	c := &FormatCache{
		path: filepath.Join(dir, "format-"+hex.EncodeToString(sum[:8])+".bin"),
		id:   fmt.Sprint(exe, "\x00", info.Size(), "\x00", info.ModTime().UnixNano()),
		old:  map[FormatKey]bool{},
		seen: map[FormatKey]bool{},
	}
	if b, err := safeio.ReadFile(c.path, int64(len(formatCacheMagic)+binary.MaxVarintLen64+len(c.id)+maxFormatEntries*sha256.Size)); err == nil {
		c.decode(b)
	}
	return c
}

// decode reads the entries of b when it was written by this binary.
func (c *FormatCache) decode(b []byte) {
	id, rest, ok := cutField(b, formatCacheMagic)
	if !ok || string(id) != c.id || len(rest)%sha256.Size != 0 {
		return
	}
	for ; len(rest) > 0; rest = rest[sha256.Size:] {
		c.old[FormatKey(rest[:sha256.Size])] = true
	}
}

// Key returns the key of formatting src with settings, a fingerprint of the
// resolved settings and target version.
func (c *FormatCache) Key(settings string, src []byte) FormatKey {
	h := sha256.New()
	fmt.Fprintf(h, "%d\x00%s\x00", len(settings), settings)
	h.Write(src)
	return FormatKey(h.Sum(nil))
}

// Formatted reports whether the source of key is known to be formatted.
func (c *FormatCache) Formatted(key FormatKey) bool {
	if c == nil || !c.old[key] {
		return false
	}
	c.keep(key)
	return true
}

// Store records that the source of key is formatted.
func (c *FormatCache) Store(key FormatKey) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.added = c.added || !c.old[key]
	c.mu.Unlock()
	c.keep(key)
}

func (c *FormatCache) keep(key FormatKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.seen[key] {
		c.seen[key] = true
		c.used = append(c.used, key)
	}
}

// Save writes the entries of this run, then those of earlier runs it did
// not use, up to the bound; nothing when the run added none.
func (c *FormatCache) Save() error {
	if c == nil || !c.added {
		return nil
	}
	out := binary.AppendUvarint([]byte(formatCacheMagic), uint64(len(c.id)))
	out = append(out, c.id...)
	n := 0
	add := func(k FormatKey) {
		if n < maxFormatEntries {
			out = append(out, k[:]...)
			n++
		}
	}
	for _, k := range c.used {
		add(k)
	}
	// Map order is random: which unused entries survive the bound does
	// not matter.
	for k := range c.old {
		if !c.seen[k] {
			add(k)
		}
	}
	return writeCacheBytes(c.path, out)
}
