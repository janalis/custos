package syntax

import (
	"sort"
	"unicode/utf8"
)

// LineIndex maps byte offsets to line/column positions.
type LineIndex struct {
	src    []byte
	starts []uint32 // byte offset of each line start
}

// NewLineIndex indexes line starts of src (\n, \r\n and \r terminate lines).
func NewLineIndex(src []byte) *LineIndex {
	starts := make([]uint32, 1, len(src)/32+1)
	for i := 0; i < len(src); i++ {
		switch src[i] {
		case '\n':
			starts = append(starts, uint32(i+1))
		case '\r':
			if i+1 < len(src) && src[i+1] == '\n' {
				i++
			}
			starts = append(starts, uint32(i+1))
		}
	}
	return &LineIndex{src: src, starts: starts}
}

// Position returns the 0-based line and 0-based byte column of offset.
func (l *LineIndex) Position(offset uint32) (line, col int) {
	line = sort.Search(len(l.starts), func(i int) bool { return l.starts[i] > offset }) - 1
	return line, int(offset - l.starts[line])
}

// RuneColumn returns the 0-based column of offset counted in runes.
func (l *LineIndex) RuneColumn(offset uint32) (line, col int) {
	line, _ = l.Position(offset)
	return line, utf8.RuneCount(l.src[l.starts[line]:offset])
}

// UTF16Column returns the 0-based line and column in UTF-16 code units (LSP default).
func (l *LineIndex) UTF16Column(offset uint32) (line, col int) {
	line, _ = l.Position(offset)
	for _, r := range string(l.src[l.starts[line]:offset]) {
		if r >= 0x10000 {
			col += 2
		} else {
			col++
		}
	}
	return line, col
}

// OffsetUTF16 converts an LSP position (0-based line, UTF-16 column) to a byte offset.
func (l *LineIndex) OffsetUTF16(line, col int) uint32 {
	if line < 0 {
		return 0
	}
	if line >= len(l.starts) {
		return uint32(len(l.src))
	}
	off := l.starts[line]
	end := uint32(len(l.src))
	if line+1 < len(l.starts) {
		end = l.starts[line+1]
	}
	units := 0
	for off < end && units < col {
		r, size := utf8.DecodeRune(l.src[off:])
		if r == '\n' || r == '\r' {
			break
		}
		if r >= 0x10000 {
			units += 2
		} else {
			units++
		}
		off += uint32(size)
	}
	return off
}

// Lines returns the number of lines.
func (l *LineIndex) Lines() int { return len(l.starts) }
