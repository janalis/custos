package index

import (
	"reflect"
	"testing"
)

// A trait use statement's own doc gives the trait its generic arguments
// (`/** @use T<TKey, TValue> */ use T;`); other tags there are not the
// class's: neither its parent's arguments nor a trait requirement.
func TestTraitUseStatementGenerics(t *testing.T) {
	fs := extract(t, "c.php", `<?php
namespace App;
class Base {}
/** @template TValue */
trait Enumerates {}
/**
 * @template TValue
 */
trait Uses {
	/**
	 * @use Enumerates<TValue>
	 * @extends Base<int>
	 * @require-extends Base
	 */
	use Enumerates;
}
/** @template TValue */
class Coll extends Base {
	/** @phpstan-use Enumerates<TValue> */
	use Enumerates;
	/** Not generic. */
	use Uses;
}
`)
	want := map[string][]SuperArgs{
		`App\Uses`: {{Class: `App\Enumerates`, Args: []string{`\~TValue`}}},
		`App\Coll`: {{Class: `App\Enumerates`, Args: []string{`\~TValue`}}},
	}
	for _, c := range fs.Classes {
		if w, ok := want[c.FQN]; ok && (!reflect.DeepEqual(c.Supers, w) || len(c.Requires) != 0) {
			t.Errorf("%s: supers %+v, requires %v", c.FQN, c.Supers, c.Requires)
		}
	}
}
