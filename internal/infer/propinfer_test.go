package infer_test

import "testing"

func TestUntypedPropertyInference(t *testing.T) {
	lib := `<?php
namespace Lib;
final class Counter {
    private $count = 0;
    private $name;
    private $items = [];
    private $maybe;
    protected $prot = 'x';
    public $pub = 1;
    private $byRef = [];
    private $mixed;
    private static $st = 1;
    private $hydrated;
    public function __construct(private $promoted = null, string $n = '') {
        $this->name = $n;
    }
    public function inc(): void { $this->count++; }
    public function add(Thing $t): void { $this->items[] = $t; }
    public function set(): void { $this->maybe = new Thing(); }
    public function withName(string $n): static { $c = clone $this; $c->name = $n . '!'; return $c; }
    public function sortIt(): void { sort($this->byRef); }
    public function setMixed($v): void { $this->mixed = $v; }
    public function readCount() { return $this->count; }
    public function readName() { return $this->name; }
    public function readItems() { return $this->items; }
    public function readMaybe() { return $this->maybe; }
    public function readProt() { return $this->prot; }
    public function readPub() { return $this->pub; }
    public function readByRef() { return $this->byRef; }
    public function readMixed() { return $this->mixed; }
    public function readPromoted() { return $this->promoted; }
    public function readHydrated() { return $this->hydrated; }
}
class Thing {}
/** @property $magic */
class Dyn {
    private $a = 1;
    public function hydrate(array $data): void { foreach ($data as $k => $v) { $this->$k = $v; } }
    public function readA() { return $this->a; }
}
`
	want := map[string]string{
		"count":     "int",
		"name":      "string",
		"items":     "array",
		"maybe":     `\Lib\Thing|null`,
		"prot":      "string",
		"pub":       "?unknown",
		"byRef":     "?unknown",
		"mixed":     "?unknown",
		"promoted":  "?unknown",
		"dyn":       "?unknown",
		"hydrated":  "?unknown",
		"pubDirect": "?unknown",
		"magic":     "?unknown",
	}
	src := `<?php
use Lib\{Counter, Dyn};
function run(Counter $c, Dyn $d) {
    t('count', $c->readCount());
    t('name', $c->readName());
    t('items', $c->readItems());
    t('maybe', $c->readMaybe());
    t('prot', $c->readProt());
    t('pub', $c->readPub());
    t('byRef', $c->readByRef());
    t('mixed', $c->readMixed());
    t('promoted', $c->readPromoted());
    t('dyn', $d->readA());
    t('hydrated', $c->readHydrated());
    t('pubDirect', $c->pub);
    t('magic', $d->magic);
}
`
	// Cross-file (index-time inference) and same-file.
	checkWith(t, map[string]string{"lib.php": lib}, src, want)
	sameFile := lib + `
function run(Counter $c, Dyn $d) {
    t('count', (new Counter())->readCount());
}
`
	checkWith(t, nil, sameFile, map[string]string{"count": "int"})
}
