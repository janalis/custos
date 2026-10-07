package diff

import "testing"

func TestUnified(t *testing.T) {
	old := "a\nb\nc\nd\ne\nf\ng\n"
	nw := "a\nb\nC\nd\ne\nf\ng\nh\n"
	got := Unified("x.php", old, nw, 1)
	want := "--- a/x.php\n+++ b/x.php\n@@ -2,3 +2,3 @@\n b\n-c\n+C\n d\n@@ -7,1 +7,2 @@\n g\n+h\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if Unified("x", "same", "same", 3) != "" {
		t.Fatal("equal texts must give empty diff")
	}
}
