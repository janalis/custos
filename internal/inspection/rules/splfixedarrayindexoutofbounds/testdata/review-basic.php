<?php
function reviewCoalesceAssign() {
    $a = new SplFixedArray(3);
    <error descr="Use an index within the fixed array size.">$a[5]</error> ??= 1;
    echo $a[<error descr="Use an index within the fixed array size.">$a[4]</error>] ?? 'inner index is a plain read';
    echo <error descr="Use an index within the fixed array size.">$a[6]</error>->name() ?? 'method call reads first';
    echo 'x' ?? <error descr="Use an index within the fixed array size.">$a[7]</error>;
    echo $o->{<error descr="Use an index within the fixed array size.">$a[8]</error>} ?? 'dynamic name is a plain read';
}
function reviewFactoryUnset() {
    $c = SplFixedArray::fromArray([1, 2, 3]);
    unset(<error descr="Use an index within the fixed array size.">$c[9]</error>);
    echo <error descr="Use an index within the fixed array size.">$c[3]</error>;
}
function reviewInertMentions($rows) {
    $a = new SplFixedArray(3);
    echo count($a), isset($a), $a->getSize();
    $twin = clone $a;
    foreach ($a as $value) {
        echo $value;
    }
    $copy = array_map(function ($a) { return $a; }, $rows);
    function reviewNested() { $a = 1; return $a; }
    <error descr="Use an index within the fixed array size.">$a[3]</error> = 1;
}
function reviewDominatingResize() {
    $a = new SplFixedArray(8);
    for ($i = 0; $i < 2; $i++) {
        $a->setSize(2);
        <error descr="Use an index within the fixed array size.">$a[5]</error> = $i;
    }
}
class ReviewOwner {
    private $items;
    private $other;
    function fill() {
        $this->items = new SplFixedArray(2);
        $this->other = SplFixedArray::fromArray([1]);
        echo strlen('builtin calls cannot reach the property');
        $noop = function () { return 1; };
        $this->items->setSize(3);
        <error descr="Use an index within the fixed array size.">$this->items[3]</error> = 1;
    }
}
