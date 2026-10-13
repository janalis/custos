<?php
class Store {
    private $items = [];
    public function &slot() {
        $s = $this->items;
        return $s;                    // method returns by reference
    }
}
function noConsumer() {
    $v = make();
    process($v);
}
function deadRead() {
    <warning descr="Variable $v is used only once; inline its value.">$v</warning> = make();
    return $v;
    echo $v;
}
function nestedFirstWrite() {
    f($v = 2);
    $v = 3;
    return $v;
}
function varFirst() {
    /** @var $g Gadget */
    $g = factory();
    return $g;
}
function longArray() {
    $pair = make();
    array($a, $b) = $pair;
}
