package infer_test

import "testing"

// TestUntypedPropertyWrites covers the write forms that make an untyped
// private property untracked (unknown) or widen it, in the same file and
// through the index.
func TestUntypedPropertyWrites(t *testing.T) {
	lib := `<?php
namespace Lib;
class Sink {
    public static function take(&$x) {}
    public function grab(&$x) {}
}
class W {
    private $listed = 1;
    private $nested = 1;
    private $refSet = 1;
    private $refRead = 1;
    private $concat = 'a';
    private $counter = 0;
    private $elemInc = [];
    private $elemArr = [];
    private $elemStr = 's';
    private $byRefLoop = [];
    private $keyTarget = 1;
    private $unsetMe = 1;
    private $unsetElem = [];
    private $staticRef = 1;
    private $methodRef = 1;
    private $newArg = 1;
    private $incStr = 'a';
    private $guarded;
    private $setInIf;
    private $neverWritten;
    private $objCount = 0;
    private $viaList = 1;
    private $lazyArr;
    private $hooked { get => 1; }
    public function __construct(private int|string $prom = 1, private $ph = 1 { get => 2; }) {
        if (\rand()) { $this->setInIf = 1; }
        $this->guarded = 2;
    }
    public function m(Sink $s, array $xs) {
        [$this->listed, [$this->nested]] = $xs;
        $x = 1;
        $this->refSet = &$x;
        $y = &$this->refRead;
        $this->concat .= 'b';
        $this->counter++;
        $this->elemInc[0]++;
        $this->elemArr['k'] = 1;
        $this->elemStr[0] = 'x';
        foreach ($this->byRefLoop as &$v) {}
        foreach ($xs as $this->keyTarget => $w) {}
        unset($this->unsetMe, $this->unsetElem[0]);
        Sink::take($this->staticRef);
        $s->grab($this->methodRef);
        new Sink;
        \strlen(...);
        $this->incStr++;
        $o = new class { private $inner = 1; public function f() { $this->inner = 'x'; } };
        $f = function () { return 1; };
        $this->objCount += 1;
        list($this->viaList) = $xs;
        $this->lazyArr[] = 1;
    }
    public function all() {
        t('viaList', $this->viaList);
        t('lazyArr', $this->lazyArr);
        t('hooked', $this->hooked);
        t('ph', $this->ph);
        t('listed', $this->listed);
        t('nested', $this->nested);
        t('refSet', $this->refSet);
        t('refRead', $this->refRead);
        t('concat', $this->concat);
        t('counter', $this->counter);
        t('elemInc', $this->elemInc);
        t('elemArr', $this->elemArr);
        t('elemStr', $this->elemStr);
        t('byRefLoop', $this->byRefLoop);
        t('keyTarget', $this->keyTarget);
        t('unsetMe', $this->unsetMe);
        t('unsetElem', $this->unsetElem);
        t('staticRef', $this->staticRef);
        t('methodRef', $this->methodRef);
        t('incStr', $this->incStr);
        t('guarded', $this->guarded);
        t('setInIf', $this->setInIf);
        t('neverWritten', $this->neverWritten);
        t('prom', $this->prom);
        t('objCount', $this->objCount);
    }
}
class Dyn {
    private $a = 1;
    private $b = 1;
    private $c = 1;
    private $d = 1;
    public function m($k, array $xs) { $this->$k++; }
    public function n($k, array $xs) { [$this->$k] = $xs; }
    public function o($k) { unset($this->$k); }
    public function p($k) { $r = &$this->$k; }
    public function q($k) { foreach ([1] as $this->$k) {} }
    public function read() { t('dyn', $this->a); }
}
`
	checkWith(t, nil, lib, map[string]string{
		"listed": "?unknown", "nested": "?unknown", "refSet": "?unknown", "refRead": "?unknown",
		"concat": "string", "counter": "int", "elemInc": "array", "elemArr": "array",
		"elemStr": "?unknown", "byRefLoop": "?unknown", "keyTarget": "?unknown", "unsetMe": "?unknown",
		"unsetElem": "array", "staticRef": "?unknown", "methodRef": "?unknown", "incStr": "?unknown",
		"guarded": "int", "setInIf": "int|null", "neverWritten": "?unknown", "prom": "int|string",
		"objCount": "?unknown", "viaList": "?unknown", "lazyArr": "array|null", "hooked": "?unknown", "ph": "?unknown", "dyn": "?unknown",
	})
}
