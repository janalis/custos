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

func TestUnifiedEdges(t *testing.T) {
	// Empty old text, change at the start (context clipped), no final newline.
	got := Unified("x", "", "a\nb", 3)
	want := "--- a/x\n+++ b/x\n@@ -0,0 +1,2 @@\n+a\n+b\n\\ No newline at end of file\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
	got = Unified("x", "a\nb\n", "a\nc\n", 3)
	want = "--- a/x\n+++ b/x\n@@ -1,2 +1,2 @@\n a\n-b\n+c\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestUnifiedEmptyRanges(t *testing.T) {
	// Without context, pure deletions and insertions have an empty range
	// that names the preceding line (as diff -U0).
	if got, want := Unified("x", "a\nb\nc\n", "a\nc\n", 0), "--- a/x\n+++ b/x\n@@ -2,1 +1,0 @@\n-b\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := Unified("x", "a\nc\n", "a\nb\nc\n", 0), "--- a/x\n+++ b/x\n@@ -1,0 +2,1 @@\n+b\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestUnifiedMergesCloseHunks(t *testing.T) {
	// Two changes separated by no more than 2*context equal lines share a hunk.
	got := Unified("x", "a\nb\nc\nd\ne\n", "a\nB\nc\nD\ne\n", 1)
	want := "--- a/x\n+++ b/x\n@@ -1,5 +1,5 @@\n a\n-b\n+B\n c\n-d\n+D\n e\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
