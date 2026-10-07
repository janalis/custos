package syntax

import (
	"sort"
	"unicode/utf8"
)

// LineIndex maps byte offsets to line/column positions.
type LineIndex struct {
	src    []byte
	starts []uint32 // byte offset of each line start
	// cps holds column checkpoints for long lines (minified code), so that
	// rune / UTF-16 columns cost O(checkpointEvery) instead of O(line
	// length) per query — otherwise reporting many findings on one huge
	// line is quadratic.
	cps map[int][]colCheckpoint
}

// colCheckpoint records, at a rune boundary of a line, the byte offset and
// the rune / UTF-16 columns reached there.
type colCheckpoint struct {
	off, runes, units uint32
}

const (
	longLine        = 4096
	checkpointEvery = 1024
)

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
	l := &LineIndex{src: src, starts: starts}
	for i := range starts {
		if end := l.lineEnd(i); end-starts[i] > longLine {
			if l.cps == nil {
				l.cps = map[int][]colCheckpoint{}
			}
			l.cps[i] = checkpoints(src, starts[i], end)
		}
	}
	return l
}

func (l *LineIndex) lineEnd(line int) uint32 {
	if line+1 < len(l.starts) {
		return l.starts[line+1]
	}
	return uint32(len(l.src))
}

func checkpoints(src []byte, start, end uint32) []colCheckpoint {
	// stop before the line terminator: checkpoints stay inside the line
	for end > start && (src[end-1] == '\n' || src[end-1] == '\r') {
		end--
	}
	cps := []colCheckpoint{{off: start}}
	var runes, units uint32
	last := start
	for off := start; off < end; {
		r, size := utf8.DecodeRune(src[off:end])
		off += uint32(size)
		runes++
		units++
		if r >= 0x10000 {
			units++
		}
		if off-last >= checkpointEvery {
			cps = append(cps, colCheckpoint{off: off, runes: runes, units: units})
			last = off
		}
	}
	return cps
}

// checkpoint returns the last checkpoint of line at or before offset (the
// line start when the line has none).
func (l *LineIndex) checkpoint(line int, offset uint32) colCheckpoint {
	cps := l.cps[line]
	if len(cps) == 0 {
		return colCheckpoint{off: l.starts[line]}
	}
	i := sort.Search(len(cps), func(i int) bool { return cps[i].off > offset }) - 1
	if i < 0 {
		i = 0
	}
	return cps[i]
}

// Position returns the 0-based line and 0-based byte column of offset.
func (l *LineIndex) Position(offset uint32) (line, col int) {
	line = sort.Search(len(l.starts), func(i int) bool { return l.starts[i] > offset }) - 1
	return line, int(offset - l.starts[line])
}

// RuneColumn returns the 0-based column of offset counted in runes.
func (l *LineIndex) RuneColumn(offset uint32) (line, col int) {
	line, _ = l.Position(offset)
	cp := l.checkpoint(line, offset)
	return line, int(cp.runes) + utf8.RuneCount(l.src[cp.off:offset])
}

// UTF16Column returns the 0-based line and column in UTF-16 code units (LSP default).
func (l *LineIndex) UTF16Column(offset uint32) (line, col int) {
	line, _ = l.Position(offset)
	cp := l.checkpoint(line, offset)
	col = int(cp.units)
	for _, r := range string(l.src[cp.off:offset]) {
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
	end := l.lineEnd(line)
	units := 0
	if cps := l.cps[line]; len(cps) > 0 && col > 0 {
		// last checkpoint not past col (checkpoints precede any line break)
		i := sort.Search(len(cps), func(i int) bool { return int(cps[i].units) > col }) - 1
		if i > 0 {
			off, units = cps[i].off, int(cps[i].units)
		}
	}
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
