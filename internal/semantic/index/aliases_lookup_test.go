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
		ix := New(nil)
		ix.Add(&FileSymbols{Path: "a.php", ClassAliases: [][2]string{{"Alias", tc.original}}})
		if got := len(ix.AliasesOf(tc.query)) == 1; got != tc.want {
			t.Errorf("%s/%s=%v want%v", tc.original, tc.query, got, tc.want)
		}
	}
}

func TestAliasesOfFollowsChanges(t *testing.T) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "a.php", ClassAliases: [][2]string{{"First", "Original"}}})
	if got := ix.AliasesOf("Original"); !slices.Equal(got, []string{"first"}) {
		t.Fatal(got)
	}
	ix.Add(&FileSymbols{Path: "b.php", ClassAliases: [][2]string{{"Second", "Original"}, {"Again", "Original"}}})
	if got := ix.AliasesOf("Original"); !slices.Equal(got, []string{"again", "first", "second"}) {
		t.Fatal(got)
	}
	ix.Remove("a.php")
	if got := ix.AliasesOf("Original"); !slices.Equal(got, []string{"again", "second"}) {
		t.Fatal(got)
	}
	ix.Freeze()
	if got := ix.AliasesOf(`\ORIGINAL`); !slices.Equal(got, []string{"again", "second"}) {
		t.Fatal(got)
	}
}

func TestAliasesOfSameAliasTwice(t *testing.T) {
	ix := New(nil)
	ix.Add(&FileSymbols{Path: "a.php", ClassAliases: [][2]string{{"Alias", "Original"}}})
	ix.Add(&FileSymbols{Path: "b.php", ClassAliases: [][2]string{{"Alias", "Original"}}})
	if got := ix.AliasesOf("Original"); !slices.Equal(got, []string{"alias", "alias"}) {
		t.Fatal(got)
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
