<?php
class Store
{
    public $items = [];
    public function &all() { return $this->items; }
    public function fill(&$target, $extra = null) { $target = []; }
    public function plain() { return []; }
}

function push_one(array &$list) { $list[] = 1; }

$s = new Store();
$s->fill($s->items, $s->plain());
$s->fill($buf);
$s->fill($s->all());
push_one($s->items['k']);
push_one(unknown_fn());
count($s->plain());
key($s->plain());
current($s->plain());
$s->fill($a, $b);
$unknown->fill($s->plain());
$fn = 'push_one';
$fn($s->plain());
$m = 'fill';
$s->$m($s->plain());
Store::$m($s->plain());
$cb = $s->fill(...);
$s->fill($a, $b, $s->plain());
push_one(nope: $s->plain());
