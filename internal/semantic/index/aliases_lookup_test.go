package index

import (
	"fmt"
	"slices"
	"testing"
)

func TestAliasesOfQualifiedCaseInsensitive(t *testing.T) {
	base := New(nil)
	base.Add(&FileSymbols{Path: "base.php", ClassAliases: [][2]string{{"BaseAlias", `\Vendor\Original`}, {"OtherAlias", "Other"}}})
	ix := New(base)
	ix.Add(&FileSymbols{Path: "local.php", ClassAliases: [][2]string{{"LocalAlias", `vendor\original`}}})
	for _, name := range []string{`Vendor\Original`, `\VENDOR\ORIGINAL`} {
		if got := ix.AliasesOf(name); !slices.Equal(got, []string{"basealias", "localalias"}) {
			t.Fatalf("%s: %v", name, got)
		}
	}
	if got := ix.AliasesOf("Absent"); len(got) != 0 {
		t.Fatal(got)
	}
}

func TestAliasLowercaseSemantics(t *testing.T) {
	for _, tc := range []struct {
		original, query string
		want            bool
	}{
		{"ſ", "S", false}, {"ς", "Σ", false}, {"É", "é", true}, {"Foo", "foo", true}, {"Foo", "bar", false}, {"Foo", "fo", false},
	} {
		if got := aliasMatches(tc.original, key(tc.query)); got != tc.want {
			t.Errorf("%s/%s=%v want%v", tc.original, tc.query, got, tc.want)
		}
	}
}

func BenchmarkAliasesOfMixedCase(b *testing.B) {
	ix := New(nil)
	fs := &FileSymbols{Path: "aliases.php"}
	for i := range 500 {
		fs.ClassAliases = append(fs.ClassAliases, [2]string{fmt.Sprintf("Alias%d", i), fmt.Sprintf("Vendor\\Original%d", i)})
	}
	ix.Add(fs)
	b.ReportAllocs()
	for b.Loop() {
		ix.AliasesOf(`\Vendor\Original499`)
	}
}
