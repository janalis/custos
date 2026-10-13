<?php
class Store
{
    public $items = [];
    public function &all() { return $this->items; }
    public function fill(&$target, $extra = null) { $target = []; }
    public function plain() { return []; }
    public static function make() { return []; }
}

function push_one(array &$list) { $list[] = 1; }

$s = new Store();
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">$s->plain()</warning>);
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">Store::make()</warning>, 1);
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">new ArrayObject()</warning>, 1);
push_one(<warning descr="Pass a variable here: this parameter is taken by reference.">array_values(...[$buf])</warning>);
push_one(list: <warning descr="Pass a variable here: this parameter is taken by reference.">$s->plain()</warning>);
