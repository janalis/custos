package flow

import (
	"errors"

	"custos/internal/semantic/wire"
)

// The binary form of summaries (see package wire), which a cache kept
// between runs stores. Every field of every type a File or a Basis reaches
// is written, and nil slices and maps stay distinct from empty ones, so
// that a decoded value is reflect.DeepEqual to the one encoded: a decoded
// summary answers Current and Reextract as the one encoded. Options are not
// written: a decoder gives every file and snapshot it reads its own.
// TestWireFields fails when a field is added without being encoded here.
//
// Summaries keep no syntax: a value still holding a node cannot be encoded,
// and fails the encoder instead of losing it.

// errSyntax fails an encoder given a value that holds a syntax node.
var errSyntax = errors.New("flow: summary holds a syntax node")

// errOverlay fails an encoder given a snapshot that is not a Basis.
var errOverlay = errors.New("flow: snapshot is not a basis")

// Encoder writes files and the snapshots they were summarized against.
// They share one table of strings and one of summaries: a summary many
// snapshots hold is written once.
type Encoder struct {
	strings *wire.Strings
	body    *wire.Writer
	table   *wire.Writer
	scratch *wire.Writer
	count   uint64
	// byContent numbers the summaries by their encoding, byPointer by the
	// pointer a snapshot holds them through.
	byContent map[string]uint64
	byPointer map[*Summary]uint64
}

// NewEncoder returns an empty encoder.
func NewEncoder() *Encoder {
	s := wire.NewStrings()
	return &Encoder{strings: s, body: wire.NewWriter(s), table: wire.NewWriter(s), scratch: wire.NewWriter(s), byContent: map[string]uint64{}, byPointer: map[*Summary]uint64{}}
}

// Writer returns the writer of the encoder's body, for the data a caller
// keeps around files and snapshots.
func (e *Encoder) Writer() *wire.Writer { return e.body }

// Mark returns the position Rollback returns to.
func (e *Encoder) Mark() int { return len(e.body.Buf) }

// Rollback drops what the body holds past mark and clears its error: an
// entry that failed to encode leaves nothing. Summaries the table gained
// meanwhile stay, unused.
func (e *Encoder) Rollback(mark int) {
	buf := e.body.Buf[:mark]
	e.body.Reset()
	e.body.Buf = buf
}

// Bytes returns the encoding: the strings, the summaries, then the body.
func (e *Encoder) Bytes() []byte {
	out := e.strings.Append(nil)
	header := wire.NewWriter(e.strings)
	header.Uint(e.count)
	header.Uint(uint64(len(e.table.Buf)))
	out = append(out, header.Buf...)
	out = append(out, e.table.Buf...)
	return append(out, e.body.Buf...)
}

// File writes f; nil is written too.
func (e *Encoder) File(f *File) {
	w := e.body
	w.Bool(f != nil)
	if f == nil {
		return
	}
	w.String(f.Path)
	w.Len(len(f.Summaries), f.Summaries == nil)
	for key, s := range f.Summaries {
		w.String(key)
		w.Uint(e.summary(&s, false))
	}
	encodeFlags(w, f.consulted)
	w.Strings(f.held)
	encodeFlags(w, f.served)
	w.Len(len(f.callables), f.callables == nil)
	for key, reads := range f.callables {
		w.String(key)
		w.Len(len(reads), reads == nil)
		for read, bits := range reads {
			w.String(read)
			w.Uint(uint64(bits))
		}
	}
}

// Snapshot writes s, which must be nil or a Basis.
func (e *Encoder) Snapshot(s *Snapshot) {
	w := e.body
	w.Bool(s != nil)
	if s == nil {
		return
	}
	if s.base != nil || s.own != nil || s.files != nil {
		w.Fail(errOverlay)
		return
	}
	w.Len(len(s.summaries), s.summaries == nil)
	for key, summary := range s.summaries {
		w.String(key)
		if summary == nil {
			w.Uint(0)
		} else {
			w.Uint(e.summary(summary, true) + 1)
		}
	}
	w.Len(len(s.shared), s.shared == nil)
	for key, n := range s.shared {
		w.String(key)
		w.Int(int64(n))
	}
}

// Err returns the error that failed the body since the last Rollback.
func (e *Encoder) Err() error { return e.body.Err() }

// summary returns the number of s in the table, adding it when no summary
// encoded alike is there; held tells a pointer a snapshot holds, which the
// encoder may number by itself. A summary that cannot be encoded fails the
// body.
func (e *Encoder) summary(s *Summary, held bool) uint64 {
	if held {
		if id, ok := e.byPointer[s]; ok {
			return id
		}
	}
	e.scratch.Reset()
	encodeSummary(e.scratch, s)
	if err := e.scratch.Err(); err != nil {
		e.body.Fail(err)
		return 0
	}
	id, ok := e.byContent[string(e.scratch.Buf)]
	if !ok {
		id = e.count
		e.count++
		e.byContent[string(e.scratch.Buf)] = id
		e.table.Buf = append(e.table.Buf, e.scratch.Buf...)
	}
	if held {
		e.byPointer[s] = id
	}
	return id
}
