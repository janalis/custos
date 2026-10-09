package index

import "testing"

func TestClassCount(t *testing.T) {
	base := New(nil)
	base.Add(extract(t, "a.php", `<?php class Dup {} class Once {}`))
	base.Add(extract(t, "b.php", `<?php class Dup {}`))
	top := New(base)
	top.Add(extract(t, "a.php", `<?php class Once {}`))
	if n := top.ClassCount(`\Dup`); n != 2 {
		t.Fatalf("Dup = %d", n)
	}
	if n := top.ClassCount("once"); n != 1 {
		t.Fatalf("Once = %d (layers must not be summed)", n)
	}
	if n := top.ClassCount("Missing"); n != 0 {
		t.Fatalf("Missing = %d", n)
	}
}
