<?php
class Item
{
    public $id;
    public function __get($k) { return null; }
}

class Magic
{
    private $data = [];
    public function __get($k) { return $this->data[$k] ?? null; }
    public function __set($k, $v) { $this->data[$k] = $v; }
}

function import_item(array $row): ?Item
{
    $item = new Item();
    foreach ($row as $field => $value) {
        $item->$field = $value; // real (dynamic) properties: no __set()
    }
    if (!isset($item->shop)) {
        $item->shop = 1;
    }
    $magic = new Magic();
    $magic->color = 'red'; // goes to __set(): isset() cannot see it
    return isset(<error descr="\Magic has no __isset(); this isset/empty check is always false.">$magic->color</error>) ? $item : null;
}
