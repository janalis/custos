package readmap

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMapStoresOnce(t *testing.T) {
	var m Map[string, int]
	if v, ok := m.Load("a"); ok || v != 0 || m.Len() != 0 {
		t.Fatalf("empty map: %d %v %d", v, ok, m.Len())
	}
	if !m.Store("a", 1) || m.Store("a", 2) {
		t.Fatal("a stored twice, or not at all")
	}
	if v, ok := m.Load("a"); !ok || v != 1 {
		t.Fatalf("a=%d %v from the stored entries", v, ok)
	}
	// The miss above promoted the stored entries: they are now read.
	if m.dirty != nil || m.read.Load().amended {
		t.Fatal("entries not promoted")
	}
	if m.Len() != 1 || m.Store("a", 3) {
		t.Fatal("read entry stored again")
	}
	if v, ok := m.Load("b"); ok || v != 0 {
		t.Fatal("absent key found")
	}
	for i := range 4 {
		m.Store(fmt.Sprint(i), i)
	}
	if m.Len() != 5 {
		t.Fatalf("len=%d", m.Len())
	}
	// Read entries stay found while others wait to be promoted, which takes
	// as many misses as there are entries.
	if v, ok := m.Load("a"); !ok || v != 1 {
		t.Fatal("read entry lost")
	}
	for range 5 {
		if v, ok := m.Load("2"); !ok || v != 2 {
			t.Fatal("stored entry lost")
		}
	}
	if m.dirty != nil {
		t.Fatal("misses did not promote the stored entries")
	}
	for i := range 4 {
		if v, ok := m.Load(fmt.Sprint(i)); !ok || v != i {
			t.Fatalf("%d lost", i)
		}
	}
}

func TestMapConcurrent(t *testing.T) {
	var m Map[[2]string, int]
	var wg sync.WaitGroup
	var added atomic.Int64
	for w := range 8 {
		wg.Go(func() {
			for i := range 2000 {
				k := [2]string{fmt.Sprint(i % 500), "x"}
				if v, ok := m.Load(k); ok {
					if v != i%500 {
						panic(fmt.Sprintf("worker %d: %v=%d", w, k, v))
					}
					continue
				}
				if m.Store(k, i%500) {
					added.Add(1)
				}
			}
		})
	}
	wg.Wait()
	if added.Load() != 500 || m.Len() != 500 {
		t.Fatalf("added=%d len=%d", added.Load(), m.Len())
	}
}

// BenchmarkMapLoad reads stored keys from parallel workers, as the caches
// of a run do once warm (compare at -cpu 4,16).
func BenchmarkMapLoad(b *testing.B) {
	keys := make([]string, 4096)
	var m Map[string, string]
	for i := range keys {
		keys[i] = fmt.Sprintf(`Illuminate\Database\Eloquent\Relations\Model%d`, i)
		m.Store(keys[i], keys[i])
	}
	for _, k := range keys {
		m.Load(k)
	}
	var seed atomic.Uint32
	b.RunParallel(func(pb *testing.PB) {
		i := int(seed.Add(7919))
		for pb.Next() {
			i++
			if _, ok := m.Load(keys[i%len(keys)]); !ok {
				panic("stored key lost")
			}
		}
	})
}
