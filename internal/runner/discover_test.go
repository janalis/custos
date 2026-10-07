package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverWithPatterns(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.php", "composer.json", "composer.lock", "sub/b.php", "sub/composer.json", "vendor/x/composer.json"} {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rel := func(files []string) []string {
		var out []string
		for _, f := range files {
			r, _ := filepath.Rel(dir, f)
			out = append(out, filepath.ToSlash(r))
		}
		return out
	}
	got, err := Discover([]string{dir}, []string{"vendor"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.php", "sub/b.php"}; !reflect.DeepEqual(rel(got), want) {
		t.Fatalf("Discover: got %v want %v", rel(got), want)
	}
	got, err = DiscoverWith([]string{dir}, []string{"vendor"}, []string{"composer.json"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.php", "composer.json", "sub/b.php", "sub/composer.json"}; !reflect.DeepEqual(rel(got), want) {
		t.Fatalf("DiscoverWith: got %v want %v", rel(got), want)
	}
}
