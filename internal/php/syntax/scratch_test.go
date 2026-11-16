package syntax

import (
	"reflect"
	"strings"
	"testing"

	phpversion "custos/internal/php/version"
)

// recycledSource allocates from every node slab.
const recycledSource = `<?php
namespace App;
use Other\Thing;
function make(int $n, ...$rest): array {
	$items = [1 => 'a', 'b' => $n, ...$rest];
	$items[] = strlen("x{$n}y") + Thing::count($n) + Thing::LIMIT;
	$obj = new Thing($n, named: 2);
	$obj->call($items)->prop = ($n * 2) . PHP_EOL;
	$items[0][1] = $obj?->prop;
	return $items;
}
class Box extends Thing {
	public function __construct(private int $size = 3) { parent::__construct(); }
	public function size(): int { return $this->size; }
}
`

func TestReleasedArenaParsesLikeFresh(t *testing.T) {
	// A bigger tree than the next one, so the next parse runs on spare chunks
	// of every slab; released twice to check the second call is a no-op.
	big := Parse("big.php", []byte(strings.Repeat(recycledSource+"\n", 3)), Options{})
	big.Release()
	big.Release()
	if big.Tokens != nil || big.Stmts != nil {
		t.Fatal("released tree still exposes its tokens or statements")
	}
	for _, v := range []phpversion.Version{phpversion.PHP74, phpversion.Max} {
		src := []byte(recycledSource)
		fresh := Parse("fresh.php", src, Options{Version: v})
		reused := Parse("fresh.php", src, Options{Version: v, Recycled: true})
		if len(fresh.arena.slabs.sVariable.spare) != 0 {
			t.Fatalf("PHP %s: a parse not asking for released memory took some", v)
		}
		// Parent links reach the files, whose arenas differ in spare chunks.
		arenas := [2]*arena{fresh.arena, reused.arena}
		fresh.arena, reused.arena = nil, nil
		if !reflect.DeepEqual(fresh.Stmts, reused.Stmts) || !reflect.DeepEqual(fresh.Tokens, reused.Tokens) || !reflect.DeepEqual(fresh.Errors, reused.Errors) {
			t.Fatalf("PHP %s: a parse on a released arena differs from a fresh one", v)
		}
		fresh.arena, reused.arena = arenas[0], arenas[1]
		fresh.Release()
		reused.Release()
	}
}

func TestReleaseWithoutArena(t *testing.T) {
	f := Parse("huge.php", make([]byte, MaxFileSize+1), Options{})
	f.Release() // the oversized file never lexed: nothing to recycle
	if len(f.Errors) != 1 {
		t.Fatal("oversized file lost its error")
	}
}

func TestSlabRecycleClearsAndReuses(t *testing.T) {
	var s slab[Variable]
	first := put(&s, Variable{Name: "a"})
	for range 8 { // fills the first 8-node chunk and starts a second
		put(&s, Variable{Name: "b"})
	}
	partial := put(&s, Variable{Name: "c"})
	s.recycle()
	if first.Name != "" || partial.Name != "" {
		t.Fatal("recycled node kept its contents")
	}
	if len(s.spare) != 2 || len(s.used) != 0 || len(s.buf) != 0 {
		t.Fatalf("spare=%d used=%d buf=%d after recycle", len(s.spare), len(s.used), len(s.buf))
	}
	// The newest chunk (16 nodes) is taken first; then the 8-node one.
	for range 16 {
		put(&s, Variable{Name: "c"})
	}
	if again := put(&s, Variable{Name: "d"}); again != first {
		t.Fatal("recycled chunk not reused")
	}
	if len(s.spare) != 0 {
		t.Fatal("spare chunks left after reuse")
	}
}

func TestSigBufferReuse(t *testing.T) {
	putSig(make([]int32, 3, 64))
	// The pool may drop items (it does at random under -race): only the
	// capacity contract is checked.
	if b := sigBuffer(10); len(b) != 0 || cap(b) < 10 {
		t.Fatalf("sigBuffer = len %d cap %d", len(b), cap(b))
	}
	if b := sigBuffer(1 << 20); cap(b) < 1<<20 {
		t.Fatal("small pooled buffer returned for a larger need")
	}
}

func BenchmarkParseReleased(b *testing.B) {
	src := []byte(strings.Repeat(recycledSource, 20))
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for b.Loop() {
		Parse("released.php", src, Options{Version: phpversion.PHP85, CopyText: true, Recycled: true}).Release()
	}
}
