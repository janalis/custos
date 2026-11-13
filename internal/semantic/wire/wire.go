// Package wire is the compact binary form semantic values take in caches
// kept between runs: varints, strings shared through one table, and lengths
// that tell a nil slice or map from an empty one, so that a decoded value is
// reflect.DeepEqual to the one encoded. Each semantic package encodes its own
// types with a Writer and decodes them with a Reader; this package knows none
// of them.
package wire

import (
	"encoding/binary"
	"errors"
	"math"
	"unsafe"
)

// MaxDepth bounds the nesting a Writer or Reader follows: deeper values are
// refused (a cycle) or corrupt.
const MaxDepth = 512

// ErrCorrupt is the error of a Reader whose input does not decode.
var ErrCorrupt = errors.New("wire: corrupt input")

// Strings is the table of the strings writers share: each is written once,
// and every use refers to it by number.
type Strings struct {
	ids  map[string]uint64
	list []string
}

// NewStrings returns an empty table.
func NewStrings() *Strings { return &Strings{ids: map[string]uint64{}} }

// Append appends the table to buf: its size, then each string.
func (s *Strings) Append(buf []byte) []byte {
	buf = binary.AppendUvarint(buf, uint64(len(s.list)))
	for _, v := range s.list {
		buf = binary.AppendUvarint(buf, uint64(len(v)))
		buf = append(buf, v...)
	}
	return buf
}

// Writer appends values to Buf. A value that cannot be encoded fails it
// (see Fail); it then writes nothing more.
type Writer struct {
	Buf     []byte
	strings *Strings
	err     error
	depth   int
}

// NewWriter returns a writer whose strings go to table.
func NewWriter(table *Strings) *Writer { return &Writer{strings: table} }

// Err returns the error that failed the writer, if any.
func (w *Writer) Err() error { return w.err }

// Fail fails the writer with err unless it already failed.
func (w *Writer) Fail(err error) {
	if w.err == nil {
		w.err = err
	}
}

// Reset empties Buf and clears the error, keeping the string table.
func (w *Writer) Reset() {
	w.Buf, w.err, w.depth = w.Buf[:0], nil, 0
}

// Uint writes v.
func (w *Writer) Uint(v uint64) {
	if w.err == nil {
		w.Buf = binary.AppendUvarint(w.Buf, v)
	}
}

// Int writes v.
func (w *Writer) Int(v int64) {
	if w.err == nil {
		w.Buf = binary.AppendVarint(w.Buf, v)
	}
}

// Float writes v bit for bit.
func (w *Writer) Float(v float64) { w.Uint(math.Float64bits(v)) }

// Bool writes v.
func (w *Writer) Bool(v bool) {
	if v {
		w.Uint(1)
	} else {
		w.Uint(0)
	}
}

// Bools writes up to 64 flags as one number.
func (w *Writer) Bools(flags ...bool) {
	var v uint64
	for i, f := range flags {
		if f {
			v |= 1 << i
		}
	}
	w.Uint(v)
}

// String writes v as its number in the table.
func (w *Writer) String(v string) {
	id, ok := w.strings.ids[v]
	if !ok {
		id = uint64(len(w.strings.list))
		w.strings.ids[v] = id
		w.strings.list = append(w.strings.list, v)
	}
	w.Uint(id)
}

// Len writes the length of a slice or map, nil being distinct from empty.
func (w *Writer) Len(n int, isNil bool) {
	if isNil {
		w.Uint(0)
	} else {
		w.Uint(uint64(n) + 1)
	}
}

// Strings writes a slice of strings.
func (w *Writer) Strings(v []string) {
	w.Len(len(v), v == nil)
	for _, s := range v {
		w.String(s)
	}
}

// Enter starts a nested value; false (and the writer failed) when it nests
// deeper than MaxDepth. Leave ends it.
func (w *Writer) Enter() bool {
	if w.depth++; w.depth > MaxDepth {
		w.Fail(errors.New("wire: value nested too deep"))
		return false
	}
	return true
}

// Leave ends a nested value Enter started.
func (w *Writer) Leave() { w.depth-- }

// Reader decodes values a Writer wrote. Once the input fails to decode,
// every read returns zero and Err returns ErrCorrupt.
type Reader struct {
	buf     []byte
	off     int
	strings []string
	err     error
	depth   int
	// alloc is the memory, in bytes, the decoded data may still take (see
	// Make): a corrupt length cannot make a small input allocate much.
	alloc int
}

// allocPerByte bounds the memory decoded data takes per input byte. A
// summary of a few hundred bytes encodes in a few dozen, so valid input
// stays far below it; corrupt input fails instead of exhausting memory.
const allocPerByte = 64

// Make returns a slice of n elements of T, charging their memory to r: nil,
// with r failed, past what r's input can hold.
func Make[T any](r *Reader, n int) []T {
	var zero T
	if !r.charge(n, int(unsafe.Sizeof(zero))) {
		return nil
	}
	return make([]T, n)
}

// MakeMap returns a map sized for n entries, charging their memory to r: an
// empty map, with r failed, past what r's input can hold.
func MakeMap[K comparable, V any](r *Reader, n int) map[K]V {
	var k K
	var v V
	if !r.charge(n, int(unsafe.Sizeof(k))+int(unsafe.Sizeof(v))+8) {
		return map[K]V{}
	}
	return make(map[K]V, n)
}

// charge takes n elements of size bytes from r's allocation budget.
func (r *Reader) charge(n, size int) bool {
	if r.err != nil || n > r.alloc/max(size, 1) {
		r.Fail()
		return false
	}
	r.alloc -= n * max(size, 1)
	return true
}

// NewReader returns a reader of buf whose strings come from table.
func NewReader(buf []byte, table []string) *Reader {
	return &Reader{buf: buf, strings: table, alloc: allocPerByte * len(buf)}
}

// ReadStrings reads a string table Strings.Append wrote at the start of buf
// and returns it with the rest of buf.
func ReadStrings(buf []byte) ([]string, []byte, error) {
	r := &Reader{buf: buf}
	n := r.Uint()
	if n > uint64(len(buf)) {
		return nil, nil, ErrCorrupt
	}
	table := make([]string, 0, n)
	for range n {
		size := r.Uint()
		if r.err != nil || size > uint64(len(r.buf)-r.off) {
			return nil, nil, ErrCorrupt
		}
		table = append(table, string(r.buf[r.off:r.off+int(size)]))
		r.off += int(size)
	}
	if r.err != nil {
		return nil, nil, r.err
	}
	return table, r.buf[r.off:], nil
}

// Err returns ErrCorrupt once a read failed.
func (r *Reader) Err() error { return r.err }

// Done reports that the whole input was read without error.
func (r *Reader) Done() bool { return r.err == nil && r.off == len(r.buf) }

// Offset returns how many bytes of the input were read.
func (r *Reader) Offset() int { return r.off }

// Fail marks the input corrupt.
func (r *Reader) Fail() { r.err = ErrCorrupt }

// Uint reads a number Writer.Uint wrote.
func (r *Reader) Uint() uint64 {
	if r.err != nil {
		return 0
	}
	v, n := binary.Uvarint(r.buf[r.off:])
	if n <= 0 {
		r.Fail()
		return 0
	}
	r.off += n
	return v
}

// Int reads a number Writer.Int wrote.
func (r *Reader) Int() int64 {
	if r.err != nil {
		return 0
	}
	v, n := binary.Varint(r.buf[r.off:])
	if n <= 0 {
		r.Fail()
		return 0
	}
	r.off += n
	return v
}

// Float reads a number Writer.Float wrote.
func (r *Reader) Float() float64 { return math.Float64frombits(r.Uint()) }

// Bool reads a flag Writer.Bool wrote.
func (r *Reader) Bool() bool { return r.Uint() != 0 }

// Bools reads n flags Writer.Bools wrote into the pointers given.
func (r *Reader) Bools(flags ...*bool) {
	v := r.Uint()
	for i, f := range flags {
		*f = v&(1<<i) != 0
	}
}

// String reads a string Writer.String wrote.
func (r *Reader) String() string {
	id := r.Uint()
	if r.err != nil || id >= uint64(len(r.strings)) {
		r.Fail()
		return ""
	}
	return r.strings[id]
}

// Len reads a length Writer.Len wrote. Every element takes at least one
// byte, so a length past the rest of the input is corrupt: a corrupt input
// cannot make the caller allocate more than it holds.
func (r *Reader) Len() (n int, isNil bool) {
	v := r.Uint()
	if v == 0 {
		return 0, true
	}
	if v-1 > uint64(len(r.buf)-r.off) {
		r.Fail()
		return 0, true
	}
	return int(v - 1), false
}

// Strings reads a slice Writer.Strings wrote.
func (r *Reader) Strings() []string {
	n, isNil := r.Len()
	if isNil {
		return nil
	}
	out := make([]string, n)
	for i := range out {
		out[i] = r.String()
	}
	return out
}

// Enter starts a nested value; false (and the input corrupt) past MaxDepth.
func (r *Reader) Enter() bool {
	if r.depth++; r.depth > MaxDepth {
		r.Fail()
		return false
	}
	return true
}

// Leave ends a nested value Enter started.
func (r *Reader) Leave() { r.depth-- }
