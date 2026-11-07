package names

import (
	"strconv"
	"testing"
)

func TestFold(t *testing.T) {
	want := map[string]string{`App\Thing`: `app\thing`, "lower": "lower", "": "", `app\thing`: `app\thing`, "ÉTÉ": "été", "été": "été", "İ": "i"}
	for _, name := range []string{`App\Thing`, `App\Thing`, "lower", "", `app\thing`, "ÉTÉ", "été", "İ"} {
		if got := Fold(name); got != want[name] {
			t.Fatalf("Fold(%q) = %q", name, got)
		}
	}
	// Past the bound, names are still folded but no longer remembered.
	before := foldedCount.Load()
	foldedCount.Store(maxFolded)
	defer foldedCount.Store(before)
	if got := Fold("Unremembered" + strconv.Itoa(1)); got != "unremembered1" {
		t.Fatal(got)
	}
	if _, ok := folded.Load("Unremembered1"); ok {
		t.Fatal("remembered past the bound")
	}
}

func BenchmarkFold(b *testing.B) {
	names := []string{`Symfony\Component\DependencyInjection\ContainerBuilder`, `psr\log\loggerinterface`, `Closure`}
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		Fold(names[i%len(names)])
	}
}
