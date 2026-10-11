<?php
function probes() {
    $a = new SplFixedArray(3);
    var_dump(isset($a[5]));
    var_dump($a[5] ?? 'default');
    var_dump(empty($a[7]));
    var_dump(isset($a[5][0], $a[6]->name));
    var_dump(($a[5]) ?? 'default');
    var_dump($a[5]['key']->name ?? 'default');
}
function closureResize() {
    $a = new SplFixedArray(3);
    $grow = function () use ($a) { $a->setSize(10); };
    $grow();
    $a[5] = 1;
}
function arrowResize() {
    $a = new SplFixedArray(3);
    $grow = fn () => $a->setSize(10);
    $grow();
    $a[5] = 1;
}
function conditionalResize($n) {
    $a = new SplFixedArray(3);
    if ($n > 3) { $a->setSize($n); }
    $a[5] = 1;
}
function conditionalLiteralResize($n) {
    $a = new SplFixedArray(3);
    if ($n) { $a->setSize(10); }
    if ($n) { $a[5] = 1; }
}
function laterIterationResize() {
    $a = new SplFixedArray(3);
    for ($i = 0; $i < 2; $i++) {
        if ($i === 1) { echo $a[4]; }
        $a->setSize(5);
    }
}
function laterIterationReplacement() {
    $a = new SplFixedArray(3);
    while (more()) {
        echo $a[4];
        $a = new SplFixedArray(9);
    }
}
function aliasResize() {
    $a = new SplFixedArray(3);
    $b = $a;
    $b->setSize(10);
    $a[5] = 1;
}
function referenceAlias() {
    $a = new SplFixedArray(3);
    $b = &$a;
    $b = new SplFixedArray(10);
    $a[5] = 1;
}
function escapedArgument(Resizer $resizer) {
    $a = new SplFixedArray(3);
    $resizer->grow($a);
    $a[5] = 1;
    $b = new SplFixedArray(3);
    call_user_func('grow', $b);
    $b[5] = 1;
    $c = new SplFixedArray(3);
    var_dump(...$c);
    $c[5] = 1;
}
function dynamicMethod($method) {
    $a = new SplFixedArray(3);
    $a->$method(10);
    $a[5] = 1;
}
function byReferenceIteration() {
    $a = new SplFixedArray(3);
    foreach ($a as &$slot) { $slot = 1; }
    $a[5] = 1;
}
function iterationTarget($lists) {
    $a = new SplFixedArray(3);
    foreach ($lists as $a) { echo count($a); }
    $a[5] = 1;
}
function wholeUnset() {
    $c = SplFixedArray::fromArray([1, 2, 3]);
    unset($c[0]);
    unset($c);
    echo $c[3];
}
function factoryResizedAfterUnset() {
    $c = SplFixedArray::fromArray([1, 2, 3]);
    unset($c[0]);
    $c->setSize(10);
    echo $c[5];
}
function unknownOrigin($c) {
    unset($c[0]);
    echo $c[3];
    $d = make();
    echo $d[3];
}
class Owner {
    private $items;
    private static $shared;
    function viaMethod() {
        $this->items = new SplFixedArray(2);
        $this->grow();
        $this->items[3] = 1;
    }
    function viaFunction() {
        $this->items = new SplFixedArray(2);
        grow_items($this);
        $this->items[3] = 1;
    }
    function viaStatic() {
        $this->items = new SplFixedArray(2);
        self::growAll();
        $this->items[3] = 1;
    }
    function viaConstructor() {
        $this->items = new SplFixedArray(2);
        new Grower($this);
        $this->items[3] = 1;
    }
    function viaDynamicClass($class) {
        $this->items = new SplFixedArray(2);
        new $class($this);
        $this->items[3] = 1;
    }
    function viaInclude() {
        $this->items = new SplFixedArray(2);
        include 'grow.php';
        $this->items[3] = 1;
    }
    function viaClosure() {
        $this->items = new SplFixedArray(2);
        $grow = function () { $this->items->setSize(10); };
        $this->items[3] = 1;
    }
    function viaArrow() {
        $this->items = new SplFixedArray(2);
        $keep = fn () => $this->items;
        $this->items[3] = 1;
    }
    function viaDynamicMethod($method) {
        $this->items = new SplFixedArray(2);
        $this->items->$method(10);
        $this->items[3] = 1;
    }
    function grow() { $this->items->setSize(10); }
    static function growAll() {}
}
