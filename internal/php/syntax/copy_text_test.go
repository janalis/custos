package syntax

import (
	"bytes"
	"testing"
)

func TestCopyTextOwnership(t *testing.T) {
	src := []byte("<?php class OwnedName {}")
	file := Parse("owned.php", src, Options{CopyText: true})
	name := file.Stmts[0].(*ClassLike).Name.Value
	// Only the captured name is inspected after this deliberate ownership probe.
	// The modified source is never parsed or analyzed again.
	at := bytes.Index(src, []byte("OwnedName"))
	src[at] = 'X'
	if name != "OwnedName" {
		t.Fatal("published name retains source text")
	}
}

func BenchmarkCopiedTextParse(b *testing.B) {
	src := []byte("<?php class Example { public function method(string $value): string { return $value; } }")
	b.ReportAllocs()
	for b.Loop() {
		Parse("copy.php", src, Options{CopyText: true})
	}
}
