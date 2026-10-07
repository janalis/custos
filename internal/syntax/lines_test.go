package syntax

import (
	"testing"
	"unicode/utf8"
)

func TestLineIndex(t *testing.T) {
	src := []byte("ab\r\nc€d\n\ne😀f")
	l := NewLineIndex(src)
	if l.Lines() != 4 {
		t.Fatalf("lines = %d", l.Lines())
	}
	if line, col := l.Position(4); line != 1 || col != 0 {
		t.Fatalf("Position(4) = %d,%d", line, col)
	}
	// 'd' after the 3-byte euro sign.
	if line, col := l.RuneColumn(8); line != 1 || col != 2 {
		t.Fatalf("RuneColumn = %d,%d", line, col)
	}
	// 'f' after a 4-byte emoji = 2 UTF-16 units.
	off := uint32(len(src) - 1)
	if line, col := l.UTF16Column(off); line != 3 || col != 3 {
		t.Fatalf("UTF16Column = %d,%d", line, col)
	}
	if got := l.OffsetUTF16(3, 3); got != off {
		t.Fatalf("OffsetUTF16 = %d want %d", got, off)
	}
}

// TestLineIndexLongLines checks the checkpointed column lookups against a
// direct count on long lines mixing ASCII, multi-byte, astral and invalid
// UTF-8 (minified code), including offsets inside a multi-byte sequence.
func TestLineIndexLongLines(t *testing.T) {
	var b []byte
	for i := 0; i < 3000; i++ {
		b = append(b, "a€😀\xff"[i%4:]...)
		if i%1500 == 1499 {
			b = append(b, "\r\n"...)
		}
	}
	b = append(b, '\n')
	l := NewLineIndex(b)
	if len(l.cps) == 0 {
		t.Fatal("no checkpoints built")
	}
	naive := func(off uint32) (runes, units int) {
		line, _ := l.Position(off)
		s := b[l.starts[line]:off]
		runes = utf8.RuneCount(s)
		for _, r := range string(s) {
			units++
			if r >= 0x10000 {
				units++
			}
		}
		return
	}
	for off := uint32(0); off <= uint32(len(b)); off += 7 {
		wr, wu := naive(off)
		if _, c := l.RuneColumn(off); c != wr {
			t.Fatalf("RuneColumn(%d) = %d, want %d", off, c, wr)
		}
		line, c := l.UTF16Column(off)
		if c != wu {
			t.Fatalf("UTF16Column(%d) = %d, want %d", off, c, wu)
		}
		if got, want := l.OffsetUTF16(line, c), naiveOffset(b, l, line, c); got != want {
			t.Fatalf("OffsetUTF16(%d,%d) = %d, want %d", line, c, got, want)
		}
		if got, want := l.OffsetUTF16(line, c+5), naiveOffset(b, l, line, c+5); got != want {
			t.Fatalf("OffsetUTF16(%d,%d) = %d, want %d", line, c+5, got, want)
		}
	}
	// past the end of a long line stops before the terminator
	if got := l.OffsetUTF16(0, 1<<30); got != l.starts[1]-2 {
		t.Fatalf("OffsetUTF16 past line end = %d, want %d", got, l.starts[1]-2)
	}
}

// naiveOffset is the uncheckpointed OffsetUTF16.
func naiveOffset(src []byte, l *LineIndex, line, col int) uint32 {
	off, end := l.starts[line], l.lineEnd(line)
	units := 0
	for off < end && units < col {
		r, size := utf8.DecodeRune(src[off:])
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
