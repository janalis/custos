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
	once.Do(func() { ix = must(decode(data)) })
	return ix
}

// decode builds an index from gzip-compressed, gob-encoded FileSymbols.
func decode(b []byte) (*index.Index, error) {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var files []*index.FileSymbols
	if err := gob.NewDecoder(zr).Decode(&files); err != nil {
		return nil, err
	}
	addPre80Failures(files)
	markBuiltin(files)
	out := index.New(nil)
	for _, f := range files {
		out.Add(f)
	}
	return out, nil
}

// must panics on a decode error: the embedded data is a build artefact, so a
// failure is a broken build, not a runtime condition.
func must(x *index.Index, err error) *index.Index {
	if err != nil {
		panic("stubs: " + err.Error())
	}
	return x
}
