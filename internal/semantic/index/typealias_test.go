package index

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func aliasFile(t *testing.T, path, src string) *FileSymbols {
	t.Helper()
	return Extract(syntax.Parse(path, []byte(src), syntax.Options{Version: phpversion.PHP84}))
}

const aliasSource = `<?php
namespace App;
/** @phpstan-type Row array{name: string} */
class Journal {}
`

const aliasImporter = `<?php
namespace App;
/**
 * @phpstan-import-type Row from Journal
 * @phpstan-import-type Gone from Nowhere
 * @template C
 */
class Access {
    /** @var Row */
    public $row;
    /** @return list<Row> */
    public function rows(): array { return []; }
    /**
     * @param Row $r
     * @param-out Row $out
     * @return array<C, Row>
     */
    public function keyed($r, &$out): array { return []; }
    /**
     * @template T
     * @param T $x
     * @return array{t: T, r: Row, g: Gone}
     */
    public function gen($x): array { return []; }
}
`

// The types an importing class records read the alias as its class
// declares it once linked: when added after that class, or once the whole
// project is linked whatever the order; mixed when it cannot be found.
func TestLinkTypeAliases(t *testing.T) {
	recorded := func(c *Class) string {
		m := c.Methods["keyed"]
		return strings.Join([]string{c.Props["row"].DocType, c.Methods["rows"].DocReturn, m.Params[0].DocType, m.Params[1].Out, m.GenReturn, c.Methods["gen"].Tpl.Return}, " ; ")
	}
	linked := `(array{name:string}) ; array<int,(array{name:string})> ; (array{name:string}) ; (array{name:string}) ; array<\~C,(array{name:string})> ; array{t:\~~T,r:(array{name:string}),g:mixed}`
	for _, tc := range []struct {
		name        string
		sourceFirst bool
		deferred    bool
		want        string
	}{
		{"source first", true, false, linked},
		{"source after", false, false, ""},
		{"deferred", false, true, linked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ix := New(nil)
			if tc.deferred {
				ix.DeferTypeAliases()
			}
			files := []*FileSymbols{aliasFile(t, "access.php", aliasImporter), aliasFile(t, "journal.php", aliasSource)}
			if tc.sourceFirst {
				files[0], files[1] = files[1], files[0]
			}
			for _, fs := range files {
				ix.Add(fs)
			}
			if tc.deferred {
				ix.LinkTypeAliases()
			}
			got := recorded(ix.Class(`App\Access`, phpversion.PHP84))
			if strings.Contains(got, AliasMark) || tc.want != "" && got != tc.want || tc.want == "" && !strings.Contains(got, "mixed") {
				t.Fatalf("recorded: %s", got)
			}
		})
	}
	// Classes of one file read each other's aliases.
	ix := New(nil)
	ix.Add(aliasFile(t, "both.php", aliasImporter+strings.TrimPrefix(aliasSource, "<?php\nnamespace App;")))
	if got := recorded(ix.Class(`App\Access`, phpversion.PHP84)); got != linked {
		t.Fatalf("same file: %s", got)
	}
	if def, ok := ix.TypeAlias(`App\Journal`, "Row"); !ok || def != "array{name:string}" {
		t.Fatalf("TypeAlias: %q %v", def, ok)
	}
}
