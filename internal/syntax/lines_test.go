package syntax

import "testing"

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
