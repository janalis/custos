package infer_test

import "testing"

// A type alias imported from another class reads as that class declares
// it, following the aliases it imports in turn.
func TestImportedTypeAliases(t *testing.T) {
	checkWith(t, map[string]string{
		"journal.php": `<?php
namespace App;
class Field {}
/** @phpstan-type SchemaField array{name: string, field: Field} */
class Journal {}
`,
		"access.php": `<?php
namespace App;
/**
 * @phpstan-import-type SchemaField from Journal
 * @phpstan-type Scope array{fields: list<SchemaField>, n?: int}
 * @phpstan-import-type Gone from Nowhere
 * @phpstan-type Partial array{g: Gone, s: string}
 */
class Access {}
/** @phpstan-import-type Cycle from Pong */
class Ping {}
/** @phpstan-import-type Cycle from Ping */
class Pong {}
`,
	}, `<?php
namespace App;
/**
 * @phpstan-import-type Scope from Access
 * @phpstan-import-type SchemaField from Journal as Column
 * @psalm-import-type Missing from Access
 * @phpstan-import-type Other from Nowhere
 * @phpstan-import-type Cycle from Ping
 * @phpstan-import-type Partial from Access
 */
class Editor {
    /** @param Scope $scope */
    public function edit(array $scope, $column, $missing, $other, $cycle, $partial) {
        foreach ($scope['fields'] as $f) { t('field', $f['field']); }
        t('n', $scope['n']);
        /** @var Column $column */
        t('column', $column['name']);
        /** @var Missing $missing */
        t('missing', $missing);
        /** @var Other $other */
        t('other', $other);
        /** @var Cycle $cycle */
        t('cycle', $cycle);
        /** @var Partial $partial */
        t('partialGone', $partial['g']);
        t('partialKept', $partial['s']);
    }
}
`, map[string]string{
		"field":   `\App\Field`,
		"n":       "int",
		"column":  "string",
		"missing": "mixed",
		"other":   "mixed",
		"cycle":   "mixed",
		// An alias it cannot resolve reads as mixed in the one it declares.
		"partialGone": "mixed",
		"partialKept": "string",
	})
}

// Members of another file whose types use an imported alias read it as its
// class declares it.
func TestImportedAliasesAcrossFiles(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": `<?php
namespace Lib;
class Field {}
/** @phpstan-type Row array{name: string, field: Field} */
class Journal {}
/** @phpstan-import-type Row from Journal */
class Access {
    /** @var list<Row> */
    public array $rows = [];
    /** @return Row|null */
    public function first(): ?array { return null; }
}
`}, `<?php
function run(\Lib\Access $a) {
    t('first', $a->first());
    foreach ($a->rows as $row) { t('field', $row['field']); }
}
`, map[string]string{"first": `array|null{name: string, field: \Lib\Field}`, "field": `\Lib\Field`})
}
