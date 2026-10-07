// Package stubs embeds the builtin PHP symbol index generated from
// JetBrains phpstorm-stubs (Apache-2.0, see NOTICE) by tools/genstubs.
package stubs

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/gob"
	"sync"

	"custos/internal/index"
)

//go:embed stubs.gob.gz
var data []byte

//go:embed VERSION
var Version string

var (
	once sync.Once
	ix   *index.Index
)

// Index returns the builtin symbol index (decoded on first use).
func Index() *index.Index {
	once.Do(func() {
		ix = index.New(nil)
		zr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			panic("stubs: " + err.Error())
		}
		var files []*index.FileSymbols
		if err := gob.NewDecoder(zr).Decode(&files); err != nil {
			panic("stubs: " + err.Error())
		}
		for _, f := range files {
			ix.Add(f)
		}
	})
	return ix
}
