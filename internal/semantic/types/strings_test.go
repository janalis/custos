package types

import (
	"strconv"
	"testing"
)

// Tracked string values narrow by listing or ruling out values, survive
// unions and removals of other atoms, and read as `string<"a"|"b">` or
// `string<not ".">`; untracked strings stay as they are.
func TestStringValues(t *testing.T) {
	many := make([]string, maxStringValues+1)
	for i := range many {
		many[i] = strconv.Itoa(i)
	}
	ab := StringValues("b", "a", "a")
	notDot := TrackedStrings().StringsWithout(".")
	for _, tc := range []struct {
		name string
		got  Type
		want string
	}{
		{"list", ab, `string<"a"|"b">`},
		{"none", StringValues(), "?unknown"},
		{"too many", StringValues(many...), "string"},
		{"tracked any", TrackedStrings(), "string"},
		{"only", ab.StringsOnly("b", "c"), `string<"b">`},
		{"only none", Union(ab, Null).StringsOnly("c"), "null"},
		{"only from any", notDot.StringsOnly(".", "x"), `string<"x">`},
		{"only untracked", String.StringsOnly("a"), "string"},
		{"without listed", ab.StringsWithout("a"), `string<"b">`},
		{"without last", StringValues("a").StringsWithout("a"), "?unknown"},
		{"without excluded", ab.StringsWithout("c"), `string<"a"|"b">`},
		{"without from any", notDot.StringsWithout(".."), `string<not "."|"..">`},
		{"without untracked", String.StringsWithout("a"), "string"},
		{"union of lists", Union(StringValues("a"), StringValues("b"), Null), `null|string<"a"|"b">`},
		{"union of exclusions", Union(notDot, notDot.StringsWithout("..")), `string<not ".">`},
		{"union of other exclusions", Union(notDot.StringsWithout(".."), notDot), `string<not ".">`},
		{"union list and exclusions", Union(StringValues("."), notDot.StringsWithout("..")), `string<not "..">`},
		{"union with untracked", Union(ab, String), "string"},
		{"union past the cap", Union(StringValues(many[:9]...), StringValues(many[9:]...)), "string"},
		{"union without strings", Union(ab, Int), `int|string<"a"|"b">`},
		{"without other atom", Union(ab, Null).Without("null"), `string<"a"|"b">`},
		{"relative", Union(ab, Of("static")).BindRelative(`\A`, "", `\A`), `\A|string<"a"|"b">`},
	} {
		if got := tc.got.ShapeString(); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
	capped := TrackedStrings()
	for _, v := range many {
		capped = capped.StringsWithout(v)
	}
	if capped.StringExcludes(many[maxStringValues]) || !capped.StringExcludes(many[0]) {
		t.Error("exclusions past the cap are given up")
	}
	if ab.String() != "string" || !ab.Equal(String) {
		t.Error("String and Equal ignore the values")
	}
	if l, ok := ab.StringValueList(); !ok || len(l) != 2 {
		t.Errorf("StringValueList: %v %v", l, ok)
	}
	if _, ok := notDot.StringValueList(); ok {
		t.Error("exclusions list no values")
	}
	if !Int.StringExcludes("a") || String.StringExcludes("a") || !notDot.StringExcludes(".") || notDot.StringExcludes("x") {
		t.Error("StringExcludes")
	}
	if !ab.StringsTracked() || String.StringsTracked() || Union(ab, Int).Without("string").StringsTracked() {
		t.Error("StringsTracked")
	}
}
