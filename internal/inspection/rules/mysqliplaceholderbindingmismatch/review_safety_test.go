package mysqliplaceholderbindingmismatch

import "testing"

func TestReviewAmbiguousComment(t *testing.T) {
	for _, tc := range []struct {
		sql  string
		want bool
	}{
		{"SELECT ?", false},
		{"SELECT ? -- note", false},
		{"SELECT ? --", false},
		{"SELECT ? --\n+ ?", false},
		{"SELECT ? --\t?\n-- ?", false},
		{"SELECT ? /* ? */", false},
		{"SELECT 5--?", true},
		{"SELECT ? -- a\n, 5--?", true},
		{"SELECT ? /*! + ? */", true},
		{"SELECT ? /*M!100100 + ? */", true},
		{"SELECT /*+ NO_MERGE(?) */ ?", true},
	} {
		if got := ambiguousComment(tc.sql); got != tc.want {
			t.Errorf("ambiguousComment(%q) = %v, want %v", tc.sql, got, tc.want)
		}
	}
}
