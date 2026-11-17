package names

import (
	"strconv"
	"strings"
	"testing"
)

func TestFold(t *testing.T) {
	want := map[string]string{`App\Thing`: `app\thing`, "lower": "lower", "": "", `app\thing`: `app\thing`, "ÉTÉ": "été", "été": "été", "İ": "i"}
	for _, name := range []string{`App\Thing`, `App\Thing`, "lower", "", `app\thing`, "ÉTÉ", "été", "İ"} {
		if got := Fold(name); got != want[name] {
			t.Fatalf("Fold(%q) = %q", name, got)
		}
	}
	long := "L" + strings.Repeat("o", maxFoldedBytes)
	if got := Fold(long); got != strings.ToLower(long) {
		t.Fatal(got)
	}
	if _, ok := folded.Load(long); ok {
		t.Fatal("remembered a name past the length bound")
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

func TestLower(t *testing.T) {
	for name, want := range map[string]string{`app\thing::run`: `app\thing::run`, `App\Thing::Run`: `app\thing::run`, "": "", "ÉTÉ": "été", "İ": "i"} {
		if got := Lower(name); got != want || got != strings.ToLower(name) {
			t.Fatalf("Lower(%q) = %q", name, got)
		}
	}
}

func TestLowerASCII(t *testing.T) {
	// Every byte at every offset of a word and of the tail after it.
	for c := range 256 {
		for at := range 12 {
			name := []byte(`abc\defg_h0z`)
			name[at] = byte(c)
			want := c < 0x80 && (c < 'A' || c > 'Z')
			if got := lowerASCII(string(name)); got != want {
				t.Fatalf("lowerASCII(%q) = %v", name, got)
			}
		}
	}
}

func BenchmarkFold(b *testing.B) {
	names := []string{`Symfony\Component\DependencyInjection\ContainerBuilder`, `psr\log\loggerinterface`, `Closure`}
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		Fold(names[i%len(names)])
	}
}

func BenchmarkLower(b *testing.B) {
	keys := []string{`illuminate\database\eloquent\builder::where`, `app\http\controllers\usercontroller::store`, `str_replace`}
	b.Run("Lower", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; b.Loop(); i++ {
			Lower(keys[i%len(keys)])
		}
	})
	b.Run("strings.ToLower", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; b.Loop(); i++ {
			strings.ToLower(keys[i%len(keys)])
		}
	})
}
