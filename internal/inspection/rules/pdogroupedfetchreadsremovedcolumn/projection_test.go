package pdogroupedfetchreadsremovedcolumn

import "testing"

// TestProjection reads the columns of a SELECT list, refusing anything but
// plain and aliased column names.
func TestProjection(t *testing.T) {
	for _, tc := range []struct {
		sql   string
		known bool
	}{
		{"UPDATE x SET y=1", false},
		{"SELECT", false},
		{"SELECT FROM x", false},
		{"SELECT x,y FROM t", true},
		{"SELECT 'x'", false},
		{"SELECT x AS 'alias'", false},
		{"SELECT x+1", false},
		{"SELECT `id`, value FROM t", true},
		{"SELECT x AS a,y AS a", false},
		{"SELECT fn(x)", false},
	} {
		_, known := selectProjection(tc.sql)
		if known != tc.known {
			t.Fatalf("projection %q: known=%v", tc.sql, known)
		}
	}
}
