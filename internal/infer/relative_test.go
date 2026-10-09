package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestRelativeMemberContracts(t *testing.T) {
	checkAnywhere(t, `<?php
class Root {}
class Base extends Root {
 public self $same;
 public parent $up;
 public static self $shared;
 public function same(): self {}
 public function up(): parent {}
 public function late(): static {}
 /** @return self[] */ public function items() {}
 /** @return callable(): parent */ public function callback() {}
 /** @return array{same: self, late: static, up: parent} */ public function shape() {}
 public static function factory(): static {}
}
class Child extends Base {
 public function forward() { t('forward', parent::factory()); }
}
trait Inner {
 public self $owner;
 public parent $ancestor;
 public function mine(): self {}
 public function above(): parent {}
 public function next(): static {}
}
trait Outer { use Inner { mine as alias; } }
class Importer extends Root { use Outer; }
class Descendant extends Importer {}
function inspect(Child $c, Descendant $d) {
 t('same', $c->same()); t('up', $c->up()); t('late', $c->late());
 t('property', $c->same); t('parentProperty', $c->up); t('staticProperty', Child::$shared);
 t('items', $c->items()[0]); $callback = $c->callback(); t('callback', $callback());
 $shape = $c->shape(); t('shapeSelf', $shape['same']); t('shapeLate', $shape['late']); t('shapeParent', $shape['up']);
 $self = $c->same(...); t('firstSelf', $self()); $late = $c->late(...); t('firstLate', $late());
 t('traitSelf', $d->mine()); t('traitAlias', $d->alias()); t('traitParent', $d->above()); t('traitStatic', $d->next());
 t('traitProperty', $d->owner); t('traitParentProperty', $d->ancestor);
}
`, map[string]string{
		"same": `\Base`, "up": `\Root`, "late": `\Child`, "forward": `\Child`,
		"property": `\Base`, "parentProperty": `\Root`, "staticProperty": `\Base`,
		"items": `\Base`, "callback": `\Root`, "shapeSelf": `\Base`, "shapeLate": `\Child`, "shapeParent": `\Root`,
		"firstSelf": `\Base`, "firstLate": `\Child`,
		"traitSelf": `\Importer`, "traitAlias": `\Importer`, "traitParent": `\Root`, "traitStatic": `\Descendant`,
		"traitProperty": `\Importer`, "traitParentProperty": `\Root`,
	})
}

func TestRelativeNativeAndTRules(t *testing.T) {
	src := `<?php
class Root {}
class Base extends Root {
 public self $p;
 public function selfType(): self {}
 public function parentType(): parent {}
 public function staticType(): static {}
 /** @return string */ public function documented(): self {}
}

class Child extends Base { function run() { t('property', $this->p); } }
function run(Child $c) {
 t('self', $c->selfType()); t('parent', $c->parentType()); t('static', $c->staticType());
 t('doc', $c->documented()); t('staticCall', Child::selfType());
 $f = $c->selfType(...); t('first', $f());
}`
	f := syntax.Parse("relative.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	want := map[string]string{"property": `\Base`, "self": `\Base`, "parent": `\Root`, "static": `\Child`, "doc": `\Base`, "staticCall": `\Base`, "first": `\Base`}
	for _, native := range []bool{false, true} {
		e := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
		if native {
			e = e.Native()
		}
		r := infer.NewTRules(e)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			call, ok := n.(*syntax.FuncCall)
			if !ok {
				return true
			}
			name, ok := call.Name.(*syntax.Name)
			if !ok || name.Value != "t" {
				return true
			}
			label := strings.Trim(call.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal).Raw, "'")
			x := call.Args.Args[1].(*syntax.Arg).Value
			if got := e.TypeOf(x).String(); got != want[label] {
				t.Errorf("native=%v %s: %s", native, label, got)
			}
			if !native && label != "doc" && label != "first" {
				if got := r.TypeOf(x).String(); got != want[label] {
					t.Errorf("T-rules %s: %s", label, got)
				}
			}
			return true
		})
	}
}

func TestRelativeCrossFile(t *testing.T) {
	checkWith(t, map[string]string{"library.php": `<?php
namespace Library;
class Root {}
trait Inner { public self $owner; public function same(): self {} public function above(): parent {} public function next(): static {} }
trait Outer { use Inner { same as alias; } }
class Base extends Root { use Outer; }
`}, `<?php
class Child extends \Library\Base {}
function run(Child $c) { t('self', $c->same()); t('alias', $c->alias()); t('parent', $c->above()); t('static', $c->next()); t('property', $c->owner); }
`, map[string]string{"self": `\Library\Base`, "alias": `\Library\Base`, "parent": `\Library\Root`, "static": `\Child`, "property": `\Library\Base`})
}
