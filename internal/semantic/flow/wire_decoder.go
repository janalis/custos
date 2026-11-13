package flow

import (
	"custos/internal/semantic/wire"
)

// Decoder reads what an Encoder wrote, giving every file and snapshot
// options.
type Decoder struct {
	r         *wire.Reader
	summaries []Summary
	options   Options
}

// NewDecoder reads the tables of b; the error is wire.ErrCorrupt when they
// do not decode.
func NewDecoder(b []byte, options Options) (*Decoder, error) {
	table, rest, err := wire.ReadStrings(b)
	if err != nil {
		return nil, err
	}
	header := wire.NewReader(rest, table)
	count, size := header.Uint(), header.Uint()
	start := header.Offset()
	// Every summary takes at least one byte.
	if header.Err() != nil || size > uint64(len(rest)-start) || count > size {
		return nil, wire.ErrCorrupt
	}
	sums := wire.NewReader(rest[start:start+int(size)], table)
	d := &Decoder{summaries: wire.Make[Summary](sums, int(count)), options: options}
	for i := range d.summaries {
		d.summaries[i] = decodeSummary(sums)
	}
	if !sums.Done() {
		return nil, wire.ErrCorrupt
	}
	d.r = wire.NewReader(rest[start+int(size):], table)
	return d, nil
}

// Reader returns the reader of the body, for what the caller wrote around
// files and snapshots.
func (d *Decoder) Reader() *wire.Reader { return d.r }

// ref returns the summary numbered id, failing the reader when there is none.
func (d *Decoder) ref(id uint64) *Summary {
	if id >= uint64(len(d.summaries)) {
		d.r.Fail()
		return &Summary{}
	}
	return &d.summaries[id]
}

// File reads a file File wrote.
func (d *Decoder) File() *File {
	r := d.r
	if !r.Bool() {
		return nil
	}
	f := &File{options: d.options, Path: r.String()}
	if n, isNil := r.Len(); !isNil {
		f.Summaries = wire.MakeMap[string, Summary](r, n)
		for range n {
			key := r.String()
			f.Summaries[key] = *d.ref(r.Uint())
		}
	}
	f.consulted = decodeFlags(r)
	f.held = r.Strings()
	f.served = decodeFlags(r)
	if n, isNil := r.Len(); !isNil {
		f.callables = wire.MakeMap[string, map[string]uint8](r, n)
		for range n {
			key := r.String()
			var reads map[string]uint8
			if m, isNil := r.Len(); !isNil {
				reads = wire.MakeMap[string, uint8](r, m)
				for range m {
					read := r.String()
					reads[read] = uint8(r.Uint())
				}
			}
			f.callables[key] = reads
		}
	}
	return f
}

// Snapshot reads a snapshot Snapshot wrote.
func (d *Decoder) Snapshot() *Snapshot {
	r := d.r
	if !r.Bool() {
		return nil
	}
	s := &Snapshot{options: d.options}
	if n, isNil := r.Len(); !isNil {
		s.summaries = wire.MakeMap[string, *Summary](r, n)
		for range n {
			key := r.String()
			var summary *Summary
			if id := r.Uint(); id > 0 {
				summary = d.ref(id - 1)
			}
			s.summaries[key] = summary
		}
	}
	if n, isNil := r.Len(); !isNil {
		s.shared = wire.MakeMap[string, int](r, n)
		for range n {
			key := r.String()
			s.shared[key] = int(r.Int())
		}
	}
	return s
}
