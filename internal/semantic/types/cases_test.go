package types

import "testing"

// The cases of an enum member survive unions and removals of other atoms,
// read as `\E::A|\E::B`, and narrow by keeping or removing one case.
func TestEnumCases(t *testing.T) {
	all := []string{"A", "B", "C"}
	ab := EnumCases(`\E`, "B", "A", "A")
	for _, tc := range []struct {
		name string
		got  Type
		want string
	}{
		{"seed", ab, `\E::A|\E::B`},
		{"none", EnumCases(`\E`), "?unknown"},
		{"only", ab.CasesOnly(`\E`, "B"), `\E::B`},
		{"only excluded", Union(ab, Null).CasesOnly(`\E`, "C"), "null"},
		{"only from any", Of(`\E`, "null").CasesOnly(`\E`, "C"), `\E::C|null`},
		{"only other enum", ab.CasesOnly(`\F`, "A"), `\E::A|\E::B`},
		{"without", ab.CasesWithout(`\E`, "A", all), `\E::B`},
		{"without last", EnumCases(`\E`, "A").CasesWithout(`\E`, "A", all), "?unknown"},
		{"without from any", Of(`\E`).CasesWithout(`\E`, "B", all), `\E::A|\E::C`},
		{"without other enum", ab.CasesWithout(`\F`, "A", all), `\E::A|\E::B`},
		{"union", Union(EnumCases(`\E`, "A"), EnumCases(`\E`, "C"), Null), `\E::A|\E::C|null`},
		{"union with any case", Union(EnumCases(`\E`, "A"), Of(`\E`)), `\E`},
		{"union with other enum", Union(EnumCases(`\E`, "A"), EnumCases(`\F`, "B")), `\E::A|\F`},
		{"without other atom", Union(ab, Null).Without("null"), `\E::A|\E::B`},
		{"without the enum", Union(ab, Null).Without(`\E`), "null"},
		{"relative", Union(ab, Of("static")).BindRelative(`\A`, "", `\A`), `\A|\E::A|\E::B`},
	} {
		if got := tc.got.ShapeString(); got != tc.want {
			t.Errorf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
	if ab.String() != `\E` || !ab.Equal(Of(`\E`)) {
		t.Error("String and Equal ignore the cases")
	}
	if names, ok := ab.EnumCaseNames(`\E`); !ok || len(names) != 2 {
		t.Errorf("EnumCaseNames: %v %v", names, ok)
	}
	if _, ok := ab.EnumCaseNames(`\F`); ok {
		t.Error("another enum has no known cases")
	}
	if _, ok := Of(`\E`).EnumCaseNames(`\E`); ok {
		t.Error("a plain enum is unconstrained")
	}
}
