package semanticquery

import "testing"

func TestNativeSQLPositionalCount(t *testing.T) {
	for _, tc := range []struct {
		sql   string
		n     int
		known bool
	}{
		{"SELECT ? + ?", 2, true},
		{"SELECT '?' /* ? */ , ? -- ?\n", 1, true},
		{"SELECT :name", 0, false},
		{"SELECT $$?$$", 0, false},
		{"SELECT ??", 0, false},
	} {
		n, known := NativeSQLPositionalCount(tc.sql)
		if n != tc.n || known != tc.known {
			t.Errorf("%q got %d %v", tc.sql, n, known)
		}
	}
}
