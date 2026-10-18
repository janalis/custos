package index

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	phpversion "custos/internal/php/version"
)

func TestMethodCacheTracksEveryIndexLayer(t *testing.T) {
	base := New(nil)
	ix := New(base)
	if ix.FindMethod("Child", "run", phpversion.PHP84) != nil {
		t.Fatal("unexpected initial declaration")
	}
	source := func(result string) *FileSymbols {
		return extract(t, "classes.php", `<?php class ParentType{function run():`+result+`{}} class Child extends ParentType{}`)
	}
	base.Add(source("int"))
	require := func(want string) {
		t.Helper()
		for range 2 {
			method := ix.FindMethod("Child", "run", phpversion.PHP84)
			if method == nil || method.Return != want {
				t.Fatalf("method=%+v, want %s", method, want)
			}
		}
	}
	require("int")
	ix.Add(source("string"))
	require("string")
	ix.Remove("classes.php")
	require("int")
	base.Add(source("float"))
	require("float")
	base.Remove("classes.php")
	if ix.FindMethod("Child", "run", phpversion.PHP84) != nil {
		t.Fatal("removed base method remained cached")
	}
}

func TestMethodCacheSeparatesVersionsAndBoundsStorage(t *testing.T) {
	ix := New(nil)
	symbols := extract(t, "versioned.php", `<?php class Versioned{function run():int{}}`)
	symbols.Classes[0].Methods["run"].Avail.From = phpversion.PHP80
	ix.Add(symbols)
	for range 2 {
		if ix.FindMethod("Versioned", "run", phpversion.PHP74) != nil {
			t.Fatal("unavailable method resolved")
		}
		if ix.FindMethod("Versioned", "run", phpversion.PHP80) == nil {
			t.Fatal("available method unresolved")
		}
	}
	ix.methods = make(map[resolvedMethodKey]resolvedMethodEntry)
	for i := range maxResolvedMethods {
		ix.methods[resolvedMethodKey{class: fmt.Sprint(i)}] = resolvedMethodEntry{}
	}
	if ix.FindMethod("Versioned", "run", phpversion.PHP80) == nil || len(ix.methods) != 1 {
		t.Fatal("cache limit changed resolution or retained excess entries")
	}
}

func TestOversizedMethodKeysAreNotRetained(t *testing.T) {
	ix := New(nil)
	long := strings.Repeat("Name", maxResolvedMethodKeyBytes)
	if ix.FindMethod(long, "run", phpversion.PHP84) != nil || ix.FindMethod("Missing", long, phpversion.PHP84) != nil || len(ix.methods) != 0 {
		t.Fatal("oversized unresolved names retained")
	}
}

func TestMethodCacheReleasesOldGenerations(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "changing.php", `<?php class Before{function run():int{}}`))
	if ix.FindMethod("Before", "run", phpversion.PHP84) == nil {
		t.Fatal("initial method missing")
	}
	ix.Add(extract(t, "changing.php", `<?php class After{function run():string{}}`))
	if len(ix.methods) != 0 {
		t.Fatal("old methods retained after declaration replacement")
	}
	if ix.FindMethod("After", "run", phpversion.PHP84) == nil || len(ix.methods) != 1 {
		t.Fatal("stale-generation symbols retained")
	}
	if ix.FindMethod("Before", "run", phpversion.PHP84) != nil {
		t.Fatal("removed class remained resolved")
	}
	ix.Remove("changing.php")
	if len(ix.methods) != 0 {
		t.Fatal("removed declaration retained")
	}
}

func TestMethodCacheConcurrentIndexChanges(t *testing.T) {
	ix := New(nil)
	symbols := extract(t, "shared.php", `<?php class Shared{function run():int{}}`)
	var group sync.WaitGroup
	for range 4 {
		group.Add(1)
		go func() {
			defer group.Done()
			for range 1000 {
				ix.FindMethod("Shared", "run", phpversion.PHP84)
			}
		}()
	}
	for range 100 {
		ix.Add(symbols)
		ix.Remove("shared.php")
	}
	group.Wait()
	ix.methodMu.RLock()
	defer ix.methodMu.RUnlock()
	for _, entry := range ix.methods {
		if entry.generation != ix.generation() || entry.method != nil {
			t.Fatal("stale method survived the final removal")
		}
	}
}
