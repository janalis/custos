// Package stubs embeds the builtin PHP symbol index generated from
// JetBrains phpstorm-stubs (Apache-2.0, see NOTICE) by tools/genstubs.
package stubs

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/gob"
	"strings"
	"sync"

	"custos/internal/semantic/index"
)

//go:embed stubs.gob.gz
var data []byte

//go:embed VERSION
var Version string

var (
	once sync.Once
	ix   *index.Index
	// exts holds the extensions the stubs describe (see Extensions).
	exts map[string]bool
)

// Index returns the builtin symbol index (decoded on first use).
func Index() *index.Index {
	load()
	return ix
}

func load() {
	once.Do(func() { ix, exts = must(decode(data)) })
}

// Extensions returns the names of the extensions the stubs describe, as
// Composer spells a platform requirement after `ext-` (lower case, spaces
// as dashes: `zend-opcache`).
func Extensions() map[string]bool {
	load()
	return exts
}

// decode builds an index from gzip-compressed, gob-encoded FileSymbols.
func decode(b []byte) (*index.Index, map[string]bool, error) {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}
	var files []*index.FileSymbols
	if err := gob.NewDecoder(zr).Decode(&files); err != nil {
		return nil, nil, err
	}
	addPre80Failures(files)
	addMissingProps(files)
	markBuiltin(files)
	out := index.New(nil)
	extensions := map[string]bool{}
	for _, f := range files {
		out.Add(f)
		if dir, _, ok := strings.Cut(strings.TrimPrefix(f.Path, "stubs/"), "/"); ok {
			extensions[strings.ReplaceAll(strings.ToLower(dir), " ", "-")] = true
		}
	}
	out.Freeze()
	return out, extensions, nil
}

// must panics on a decode error: the embedded data is a build artefact, so a
// failure is a broken build, not a runtime condition.
func must(x *index.Index, exts map[string]bool, err error) (*index.Index, map[string]bool) {
	if err != nil {
		panic("stubs: " + err.Error())
	}
	return x, exts
}
