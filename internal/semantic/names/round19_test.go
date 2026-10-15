package names

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestImportsBeginAtDeclaration(t *testing.T) {
	src := `<?php namespace N { new Alias; use X\Y as Alias; use function X\f as fun; use const X\C; new Alias; fun(); echo C; } namespace M { new Alias; }`
	r := New(syntax.Parse("probe.php", []byte(src), syntax.Options{}))
	first, last := uint32(strings.Index(src, "new Alias")), uint32(strings.LastIndex(src, "new Alias"))
	middle := uint32(strings.Index(src, "new Alias; fun"))
	if r.Class("Alias", first) != `N\Alias` || r.Class("Alias", middle) != `X\Y` || r.Class("Alias", last) != `M\Alias` {
		t.Fatal("class import position")
	}
	if f, _ := r.Function("fun", middle); f != `X\f` {
		t.Fatal(f)
	}
	if c, _ := r.Const("C", middle); c != `X\C` {
		t.Fatal(c)
	}
}

func BenchmarkManyImports(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php namespace App;\n")
	for i := range 500 {
		fmt.Fprintf(&src, "use Lib\\Class%d;\n", i)
	}
	f := syntax.Parse("imports.php", []byte(src.String()), syntax.Options{})
	b.Run("construct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			New(f)
		}
	})
	r := New(f)
	b.Run("lookup", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r.Class("Class499", uint32(len(f.Src)-1))
		}
	})
}

func TestImportHistory(t *testing.T) {
	src := `<?php use A\C; echo 1; use B\C; echo 2; use D\C; echo 3;`
	r := New(syntax.Parse("duplicate.php", []byte(src), syntax.Options{}))
	if r.Class("!C", 1) != "" || r.Class("!C", uint32(strings.Index(src, "echo 1"))) != `A\C` || r.Class("!C", uint32(strings.Index(src, "echo 2"))) != `B\C` || r.Class("!C", uint32(strings.Index(src, "echo 3"))) != `D\C` {
		t.Fatal("import history")
	}
}

func BenchmarkDuplicateImportLookup(b *testing.B) {
	for _, count := range []int{500, 2000} {
		var src strings.Builder
		src.WriteString("<?php namespace App;\n")
		for i := range count {
			fmt.Fprintf(&src, "use Lib\\Class%d as Alias;\n", i)
		}
		f := syntax.Parse("duplicates.php", []byte(src.String()), syntax.Options{})
		r := New(f)
		at := uint32(strings.Index(src.String(), "use Lib\\Class1 as"))
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				r.Class("Alias", at)
			}
		})
	}
}
